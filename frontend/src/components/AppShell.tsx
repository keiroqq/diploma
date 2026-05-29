import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type PointerEvent,
  type WheelEvent
} from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowDown,
  Bookmark,
  CalendarDays,
  Check,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Compass,
  Database,
  Filter,
  Globe2,
  LogOut,
  Menu,
  Rss,
  Search,
  UserCircle,
  X
} from "lucide-react";
import {
  NavLink,
  Outlet,
  useLocation,
  useNavigate,
  useSearchParams
} from "react-router-dom";

import {
  getMe,
  listCategories,
  listFeedCategories,
  listFeeds,
  saveItem,
  searchItems,
  unsaveItem
} from "../api/client";
import type { Item } from "../api/types";
import { ArticleCard } from "./ArticleCard";
import { ArticleReaderModal } from "./ArticleReaderModal";
import { useAuthStore } from "../store/auth";
import { useUiStore } from "../store/ui";
import { errorMessage } from "../utils/errors";
import {
  categoryFilterLabel,
  getDateFilter,
  getSelectedCategorySlugs,
  localDateString,
  type DatePreset
} from "../utils/filters";
import { searchLocalItems, toggleLocalItemSaved } from "../utils/localItems";

function feedPillStyle(themeColor?: string): CSSProperties {
  return {
    "--feed-pill-color": themeColor || "#2563eb"
  } as CSSProperties;
}

function byPublishedDesc(left: Item, right: Item) {
  return new Date(right.published_at).getTime() - new Date(left.published_at).getTime();
}

const FEED_SWITCH_BOTTOM_THRESHOLD = 8;
const FEED_SWITCH_WHEEL_THRESHOLD = 340;
const FEED_SWITCH_TOUCH_THRESHOLD = 150;
const FEED_SWITCH_COOLDOWN_MS = 800;
const FEED_SWITCH_RESET_DELAY_MS = 520;

export function AppShell() {
  const location = useLocation();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const queryClient = useQueryClient();
  const token = useAuthStore((state) => state.token);
  const user = useAuthStore((state) => state.user);
  const setUser = useAuthStore((state) => state.setUser);
  const logout = useAuthStore((state) => state.logout);
  const drawerOpen = useUiStore((state) => state.drawerOpen);
  const searchOpen = useUiStore((state) => state.searchOpen);
  const searchQuery = useUiStore((state) => state.searchQuery);
  const searchScope = useUiStore((state) => state.searchScope);
  const readerItem = useUiStore((state) => state.readerItem);
  const setDrawerOpen = useUiStore((state) => state.setDrawerOpen);
  const setSearchOpen = useUiStore((state) => state.setSearchOpen);
  const setSearchQuery = useUiStore((state) => state.setSearchQuery);
  const setSearchScope = useUiStore((state) => state.setSearchScope);
  const closeSearch = useUiStore((state) => state.closeSearch);
  const closeReader = useUiStore((state) => state.closeReader);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const dateFromInputRef = useRef<HTMLInputElement>(null);
  const dateToInputRef = useRef<HTMLInputElement>(null);
  const feedStripRef = useRef<HTMLDivElement>(null);
  const feedScrollTrackRef = useRef<HTMLDivElement>(null);
  const feedPillRefs = useRef<Record<string, HTMLElement | null>>({});
  const feedScrollHoldRef = useRef<number | null>(null);
  const feedSwitchResetTimerRef = useRef<number | null>(null);
  const feedSwitchWheelDistanceRef = useRef(0);
  const feedSwitchLockedRef = useRef(false);
  const feedSwitchTouchRef = useRef({
    active: false,
    distance: 0,
    lastY: 0,
    ready: false
  });
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [dateMenuOpen, setDateMenuOpen] = useState(false);
  const [categoryMenuOpen, setCategoryMenuOpen] = useState(false);
  const [draftDateFrom, setDraftDateFrom] = useState("");
  const [draftDateTo, setDraftDateTo] = useState("");
  const [draftCategories, setDraftCategories] = useState<string[]>([]);
  const [feedStripDragging, setFeedStripDragging] = useState(false);
  const [feedStripScroll, setFeedStripScroll] = useState({
    canScroll: false,
    canScrollLeft: false,
    canScrollRight: false,
    progress: 0,
    thumbWidth: 100
  });
  const [feedSwitchProgress, setFeedSwitchProgress] = useState(0);

  const feedsQuery = useQuery({
    queryKey: ["feeds"],
    queryFn: listFeeds,
    enabled: Boolean(token)
  });

  const meQuery = useQuery({
    queryKey: ["me"],
    queryFn: getMe,
    enabled: Boolean(token),
    staleTime: 5 * 60_000
  });

  useEffect(() => {
    if (meQuery.data) {
      setUser(meQuery.data);
    }
  }, [meQuery.data, setUser]);

  const feedId = location.pathname.match(/^\/feeds\/([^/]+)/)?.[1];
  const isFeedRoute = Boolean(feedId);
  const hasFeedTools = isFeedRoute || searchOpen;
  const currentFeed = feedsQuery.data?.find((feed) => feed.id === feedId);
  const searchScopeFeed = searchScope.type === "feed"
    ? feedsQuery.data?.find((feed) => feed.id === searchScope.feedId)
    : undefined;

  const categoriesQuery = useQuery({
    queryKey: ["categories"],
    queryFn: listCategories,
    enabled: Boolean(token) && !isFeedRoute
  });

  const feedCategoriesQuery = useQuery({
    queryKey: ["feedCategories", feedId],
    queryFn: () => listFeedCategories(feedId ?? ""),
    enabled: Boolean(token) && isFeedRoute
  });

  const dateFilter = getDateFilter(searchParams);
  const selectedCategories = getSelectedCategorySlugs(searchParams);
  const categories = isFeedRoute
    ? feedCategoriesQuery.data ?? []
    : categoriesQuery.data ?? [];
  const categoriesLoading = isFeedRoute
    ? feedCategoriesQuery.isLoading
    : categoriesQuery.isLoading;
  const categoryLabel = categoryFilterLabel(categories, selectedCategories);
  const headerOffset = hasFeedTools ? "132px" : "64px";
  const popoverOpen = dateMenuOpen || categoryMenuOpen;
  const normalizedSearchQuery = searchQuery.trim();
  const searchPlaceholder = searchScope.type === "feed"
    ? `Поиск в потоке "${searchScopeFeed?.name ?? currentFeed?.name ?? "Поток"}"`
    : "Поиск по всем доступным статьям";

  const pageTitle = useMemo(() => {
    if (location.pathname === "/catalog") {
      return "Каталог";
    }
    if (location.pathname === "/saved") {
      return "Избранное";
    }
    if (location.pathname === "/sources") {
      return "Источники";
    }
    if (feedId) {
      return currentFeed?.name ?? "Поток";
    }
    return "Мои ленты";
  }, [currentFeed?.name, feedId, location.pathname]);

  const searchAvailable =
    location.pathname.startsWith("/feeds/") || location.pathname === "/saved";
  const activeFeedPillId =
    searchOpen && searchScope.type === "feed" ? searchScope.feedId : feedId;
  const feeds = feedsQuery.data ?? [];
  const currentFeedIndex = feedId
    ? feeds.findIndex((feed) => feed.id === feedId)
    : -1;
  const nextFeed = currentFeedIndex >= 0 ? feeds[currentFeedIndex + 1] : undefined;
  const feedSwitchIndicatorStyle = {
    "--feed-switch-color": nextFeed?.theme_color || "#2563eb",
    "--feed-switch-progress": feedSwitchProgress
  } as CSSProperties;

  const searchItemsQuery = useQuery({
    queryKey: [
      "searchItems",
      searchScope.type,
      searchScope.type === "feed" ? searchScope.feedId : null,
      normalizedSearchQuery
    ],
    queryFn: () =>
      searchItems(normalizedSearchQuery, {
        feedId: searchScope.type === "feed" ? searchScope.feedId : undefined,
        limit: 200
      }),
    enabled:
      Boolean(token) &&
      searchOpen &&
      normalizedSearchQuery.length > 0 &&
      (searchScope.type === "feed" ? Boolean(searchScope.feedId) : true)
  });

  const localSearchItemsQuery = useQuery({
    queryKey: [
      "localSearchItems",
      searchScope.type,
      searchScope.type === "feed" ? searchScope.feedId : null,
      normalizedSearchQuery
    ],
    queryFn: () =>
      searchLocalItems(
        normalizedSearchQuery,
        searchScope.type === "feed" ? searchScope.feedId : undefined
      ),
    enabled:
      searchOpen &&
      normalizedSearchQuery.length > 0 &&
      (searchScope.type === "feed" ? Boolean(searchScope.feedId) : true)
  });

  const searchResults = useMemo(
    () =>
      [
        ...(searchItemsQuery.data?.items ?? []),
        ...(localSearchItemsQuery.data ?? [])
      ].sort(byPublishedDesc),
    [localSearchItemsQuery.data, searchItemsQuery.data?.items]
  );

  const toggleSearchSavedMutation = useMutation({
    mutationFn: (item: Item) => (item.is_saved ? unsaveItem(item.id) : saveItem(item.id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feedItems"] });
      queryClient.invalidateQueries({ queryKey: ["saved"] });
      queryClient.invalidateQueries({ queryKey: ["searchItems"] });
    }
  });

  const toggleLocalSearchSavedMutation = useMutation({
    mutationFn: toggleLocalItemSaved,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["localFeedItems"] });
      queryClient.invalidateQueries({ queryKey: ["localSearchItems"] });
    }
  });

  useEffect(() => {
    if (!searchAvailable && searchOpen) {
      closeSearch();
    }
  }, [closeSearch, searchAvailable, searchOpen]);

  useEffect(() => {
    if (searchOpen) {
      searchInputRef.current?.focus();
    }
  }, [searchOpen]);

  useEffect(() => {
    setDraftDateFrom(dateFilter.dateFrom ?? "");
    setDraftDateTo(dateFilter.dateTo ?? "");
  }, [dateFilter.dateFrom, dateFilter.dateTo]);

  useEffect(() => {
    setDraftCategories(selectedCategories);
  }, [selectedCategories.join(",")]);

  useEffect(() => {
    if (!isFeedRoute || !feedCategoriesQuery.isSuccess || !selectedCategories.length) {
      return;
    }

    const availableSlugs = new Set(categories.map((category) => category.slug));
    const nextSelected = selectedCategories.filter((slug) => availableSlugs.has(slug));
    if (nextSelected.length === selectedCategories.length) {
      return;
    }

    const nextParams = new URLSearchParams(searchParams);
    nextParams.delete("category");
    nextParams.delete("categories");
    if (nextSelected.length) {
      nextParams.set("categories", nextSelected.join(","));
    }

    setDraftCategories(nextSelected);
    setSearchParams(nextParams, { replace: true });
  }, [
    categories,
    feedCategoriesQuery.isSuccess,
    isFeedRoute,
    searchParams,
    selectedCategories.join(","),
    setSearchParams
  ]);

  useEffect(() => {
    if (!isFeedRoute) {
      setFiltersOpen(false);
      setDateMenuOpen(false);
      setCategoryMenuOpen(false);
    }
  }, [isFeedRoute]);

  useEffect(() => {
    if (!hasFeedTools) {
      return;
    }

    if (activeFeedPillId) {
      feedPillRefs.current[activeFeedPillId]?.scrollIntoView({
        behavior: "smooth",
        block: "nearest",
        inline: "center"
      });
    }

    window.requestAnimationFrame(updateFeedStripScrollState);
  }, [activeFeedPillId, feedsQuery.data?.length, hasFeedTools]);

  useEffect(() => {
    if (!hasFeedTools) {
      return;
    }

    const feedStrip = feedStripRef.current;
    if (!feedStrip) {
      return;
    }

    const update = () => updateFeedStripScrollState();
    update();
    feedStrip.addEventListener("scroll", update, { passive: true });

    const resizeObserver = new ResizeObserver(update);
    resizeObserver.observe(feedStrip);

    return () => {
      feedStrip.removeEventListener("scroll", update);
      resizeObserver.disconnect();
      stopFeedStripHold();
    };
  }, [feedsQuery.data?.length, hasFeedTools]);

  useEffect(() => {
    if (!isFeedRoute || !nextFeed || searchOpen) {
      resetFeedSwitchGesture();
      return;
    }

    function handleWheel(event: globalThis.WheelEvent) {
      if (event.defaultPrevented || feedSwitchLockedRef.current) {
        return;
      }
      if (event.deltaY <= 0 || Math.abs(event.deltaX) > Math.abs(event.deltaY)) {
        resetFeedSwitchGesture();
        return;
      }
      if (!canStartFeedSwitchGesture(event.target)) {
        resetFeedSwitchGesture();
        return;
      }

      feedSwitchWheelDistanceRef.current += event.deltaY;
      setFeedSwitchGestureProgress(
        feedSwitchWheelDistanceRef.current / FEED_SWITCH_WHEEL_THRESHOLD
      );

      if (feedSwitchWheelDistanceRef.current >= FEED_SWITCH_WHEEL_THRESHOLD) {
        event.preventDefault();
        switchToNextFeed();
      } else {
        scheduleFeedSwitchReset();
      }
    }

    function handleTouchStart(event: TouchEvent) {
      const touch = event.touches[0];
      if (!touch || !canStartFeedSwitchGesture(event.target)) {
        resetFeedSwitchGesture();
        return;
      }

      feedSwitchTouchRef.current = {
        active: true,
        distance: 0,
        lastY: touch.clientY,
        ready: false
      };
    }

    function handleTouchMove(event: TouchEvent) {
      const touch = event.touches[0];
      const state = feedSwitchTouchRef.current;
      if (!touch || !state.active || feedSwitchLockedRef.current) {
        return;
      }

      const deltaY = state.lastY - touch.clientY;
      state.lastY = touch.clientY;

      if (deltaY <= 0) {
        state.distance = Math.max(0, state.distance + deltaY);
        setFeedSwitchGestureProgress(state.distance / FEED_SWITCH_TOUCH_THRESHOLD);
        state.ready = state.distance >= FEED_SWITCH_TOUCH_THRESHOLD;
        return;
      }

      state.distance += deltaY;
      setFeedSwitchGestureProgress(state.distance / FEED_SWITCH_TOUCH_THRESHOLD);
      if (state.distance >= FEED_SWITCH_TOUCH_THRESHOLD) {
        event.preventDefault();
        state.ready = true;
      }
    }

    function handleTouchEnd() {
      if (feedSwitchTouchRef.current.ready) {
        switchToNextFeed();
      } else {
        resetFeedSwitchGesture();
      }
    }

    window.addEventListener("wheel", handleWheel, { passive: false });
    window.addEventListener("touchstart", handleTouchStart, { passive: true });
    window.addEventListener("touchmove", handleTouchMove, { passive: false });
    window.addEventListener("touchend", handleTouchEnd, { passive: true });
    window.addEventListener("touchcancel", resetFeedSwitchGesture, { passive: true });

    return () => {
      window.removeEventListener("wheel", handleWheel);
      window.removeEventListener("touchstart", handleTouchStart);
      window.removeEventListener("touchmove", handleTouchMove);
      window.removeEventListener("touchend", handleTouchEnd);
      window.removeEventListener("touchcancel", resetFeedSwitchGesture);
      clearFeedSwitchResetTimer();
    };
  }, [
    drawerOpen,
    filtersOpen,
    isFeedRoute,
    location.search,
    nextFeed?.id,
    popoverOpen,
    searchOpen
  ]);

  function handleLogout() {
    queryClient.clear();
    logout();
    navigate("/login", { replace: true });
  }

  function isPageAtBottom() {
    const scrollingElement = document.scrollingElement ?? document.documentElement;
    const scrollTop = scrollingElement.scrollTop;
    const viewportHeight = window.innerHeight || scrollingElement.clientHeight;
    const scrollHeight = Math.max(
      scrollingElement.scrollHeight,
      document.body.scrollHeight
    );

    return scrollTop + viewportHeight >= scrollHeight - FEED_SWITCH_BOTTOM_THRESHOLD;
  }

  function canStartFeedSwitchGesture(target: EventTarget | null) {
    if (
      !isFeedRoute ||
      searchOpen ||
      drawerOpen ||
      filtersOpen ||
      popoverOpen ||
      !nextFeed ||
      feedSwitchLockedRef.current ||
      !isPageAtBottom()
    ) {
      return false;
    }

    const element = target instanceof Element ? target : null;
    if (
      element?.closest(
        ".drawer, .feed-strip-frame, .filter-panel, .filter-popover, input, textarea, select"
      )
    ) {
      return false;
    }

    return true;
  }

  function resetFeedSwitchGesture() {
    clearFeedSwitchResetTimer();
    clearFeedSwitchGestureDistances();
    setFeedSwitchProgress(0);
  }

  function clearFeedSwitchGestureDistances() {
    feedSwitchWheelDistanceRef.current = 0;
    feedSwitchTouchRef.current = {
      active: false,
      distance: 0,
      lastY: 0,
      ready: false
    };
  }

  function clearFeedSwitchResetTimer() {
    if (feedSwitchResetTimerRef.current !== null) {
      window.clearTimeout(feedSwitchResetTimerRef.current);
      feedSwitchResetTimerRef.current = null;
    }
  }

  function scheduleFeedSwitchReset() {
    clearFeedSwitchResetTimer();
    feedSwitchResetTimerRef.current = window.setTimeout(() => {
      resetFeedSwitchGesture();
    }, FEED_SWITCH_RESET_DELAY_MS);
  }

  function setFeedSwitchGestureProgress(progress: number) {
    clearFeedSwitchResetTimer();
    const nextProgress = Math.min(1, Math.max(0, progress));
    setFeedSwitchProgress(nextProgress);
    if (nextProgress > 0) {
      scrollFeedSwitchZoneIntoView();
    }
  }

  function scrollFeedSwitchZoneIntoView() {
    window.requestAnimationFrame(() => {
      const scrollingElement = document.scrollingElement ?? document.documentElement;
      const maxScroll = scrollingElement.scrollHeight - scrollingElement.clientHeight;
      if (maxScroll > 0) {
        window.scrollTo({ top: maxScroll, behavior: "auto" });
      }
    });
  }

  function switchToNextFeed() {
    if (!nextFeed || feedSwitchLockedRef.current) {
      return;
    }

    feedSwitchLockedRef.current = true;
    clearFeedSwitchResetTimer();
    clearFeedSwitchGestureDistances();
    setFeedSwitchProgress(1);
    scrollFeedSwitchZoneIntoView();
    setFiltersOpen(false);
    closeFilterPopovers();

    feedPillRefs.current[nextFeed.id]?.scrollIntoView({
      behavior: "smooth",
      block: "nearest",
      inline: "center"
    });

    window.setTimeout(() => {
      navigate(`/feeds/${nextFeed.id}${location.search}`);

      window.requestAnimationFrame(() => {
        window.scrollTo({ top: 0, behavior: "auto" });
      });
      window.setTimeout(() => {
        setFeedSwitchProgress(0);
        feedSwitchLockedRef.current = false;
      }, FEED_SWITCH_COOLDOWN_MS);
    }, 120);
  }

  function closeDrawer() {
    setDrawerOpen(false);
  }

  function openDrawer() {
    setFiltersOpen(false);
    closeFilterPopovers();
    setDrawerOpen(true);
  }

  function openSearch() {
    closeFilterPopovers();
    setFiltersOpen(false);
    setDrawerOpen(false);

    if (feedId) {
      setSearchScope({ type: "feed", feedId });
    } else {
      setSearchScope({ type: "all" });
    }

    setSearchOpen(true);
  }

  function updateFilterParams(mutator: (next: URLSearchParams) => void) {
    const next = new URLSearchParams(searchParams);
    mutator(next);
    setSearchParams(next, { replace: false });
  }

  function applyDatePreset(preset: DatePreset) {
    updateFilterParams((next) => {
      next.delete("date_from");
      next.delete("date_to");

      if (preset === "all") {
        next.delete("date");
      } else {
        next.set("date", preset);
      }
    });
    setDateMenuOpen(false);
  }

  function applyCustomDate() {
    updateFilterParams((next) => {
      next.set("date", "custom");
      if (draftDateFrom) {
        next.set("date_from", draftDateFrom);
      } else {
        next.delete("date_from");
      }
      if (draftDateTo) {
        next.set("date_to", draftDateTo);
      } else {
        next.delete("date_to");
      }
    });
    setDateMenuOpen(false);
  }

  function toggleDraftCategory(slug: string) {
    setDraftCategories((current) =>
      current.includes(slug)
        ? current.filter((item) => item !== slug)
        : [...current, slug]
    );
  }

  function applyCategories() {
    updateFilterParams((next) => {
      next.delete("category");
      if (draftCategories.length) {
        next.set("categories", draftCategories.join(","));
      } else {
        next.delete("categories");
      }
    });
    setCategoryMenuOpen(false);
  }

  function closeFilterPopovers() {
    setDateMenuOpen(false);
    setCategoryMenuOpen(false);
  }

  function updateFeedStripScrollState() {
    const feedStrip = feedStripRef.current;
    if (!feedStrip) {
      return;
    }

    const maxScroll = feedStrip.scrollWidth - feedStrip.clientWidth;
    const canScroll = maxScroll > 1;
    const progress = canScroll ? feedStrip.scrollLeft / maxScroll : 0;
    const thumbWidth = canScroll
      ? Math.max(12, (feedStrip.clientWidth / feedStrip.scrollWidth) * 100)
      : 100;

    setFeedStripScroll({
      canScroll,
      canScrollLeft: feedStrip.scrollLeft > 1,
      canScrollRight: canScroll && feedStrip.scrollLeft < maxScroll - 1,
      progress,
      thumbWidth
    });
  }

  function scrollFeedStrip(direction: -1 | 1, distance = 120) {
    const feedStrip = feedStripRef.current;
    if (!feedStrip) {
      return;
    }

    feedStrip.scrollBy({ left: direction * distance, behavior: "auto" });
    window.requestAnimationFrame(updateFeedStripScrollState);
  }

  function startFeedStripHold(direction: -1 | 1) {
    stopFeedStripHold();
    scrollFeedStrip(direction, 148);
    feedScrollHoldRef.current = window.setInterval(() => {
      scrollFeedStrip(direction, 44);
    }, 55);
  }

  function stopFeedStripHold() {
    if (feedScrollHoldRef.current !== null) {
      window.clearInterval(feedScrollHoldRef.current);
      feedScrollHoldRef.current = null;
    }
  }

  function scrollFeedStripToPointer(clientX: number) {
    const track = feedScrollTrackRef.current;
    const feedStrip = feedStripRef.current;
    if (!track || !feedStrip || !feedStripScroll.canScroll) {
      return;
    }

    const trackRect = track.getBoundingClientRect();
    const thumbWidth = (feedStripScroll.thumbWidth / 100) * trackRect.width;
    const availableWidth = trackRect.width - thumbWidth;
    if (availableWidth <= 0) {
      return;
    }

    const rawLeft = clientX - trackRect.left - thumbWidth / 2;
    const progress = Math.min(1, Math.max(0, rawLeft / availableWidth));
    feedStrip.scrollLeft = progress * (feedStrip.scrollWidth - feedStrip.clientWidth);
    updateFeedStripScrollState();
  }

  function handleFeedScrollTrackPointerDown(event: PointerEvent<HTMLDivElement>) {
    if (!feedStripScroll.canScroll) {
      return;
    }

    event.preventDefault();
    event.currentTarget.setPointerCapture(event.pointerId);
    setFeedStripDragging(true);
    scrollFeedStripToPointer(event.clientX);
  }

  function handleFeedScrollTrackPointerMove(event: PointerEvent<HTMLDivElement>) {
    if (!feedStripDragging) {
      return;
    }

    event.preventDefault();
    scrollFeedStripToPointer(event.clientX);
  }

  function handleFeedScrollTrackPointerEnd(event: PointerEvent<HTMLDivElement>) {
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
    setFeedStripDragging(false);
  }

  function handleFeedStripWheel(event: WheelEvent<HTMLDivElement>) {
    const feedStrip = feedStripRef.current;
    if (!feedStrip || !feedStripScroll.canScroll) {
      return;
    }

    const delta = Math.abs(event.deltaX) > Math.abs(event.deltaY)
      ? event.deltaX
      : event.deltaY;
    if (delta === 0) {
      return;
    }

    event.preventDefault();
    feedStrip.scrollLeft += delta;
    updateFeedStripScrollState();
  }

  function openDatePicker(input: HTMLInputElement | null) {
    if (!input) {
      return;
    }

    const dateInput = input as HTMLInputElement & { showPicker?: () => void };

    if (dateInput.showPicker) {
      try {
        dateInput.showPicker();
        return;
      } catch {
        dateInput.focus();
        return;
      }
    }

    dateInput.focus();
  }

  function closeFiltersFromHeaderClick(event: React.MouseEvent<HTMLElement>) {
    const target = event.target as HTMLElement;
    if (target.closest(".filter-popover")) {
      return;
    }

    if (popoverOpen) {
      event.preventDefault();
      event.stopPropagation();
      closeFilterPopovers();
      return;
    }

    if (!filtersOpen) {
      return;
    }
    if (target.closest(".filter-panel") || target.closest(".filter-toggle")) {
      return;
    }

    closeFilterPopovers();
    setFiltersOpen(false);
  }

  function renderSearchSurface() {
    const searchLoading =
      normalizedSearchQuery.length > 0 &&
      (searchItemsQuery.isLoading || localSearchItemsQuery.isLoading);
    const searchError = searchItemsQuery.error ?? localSearchItemsQuery.error;

    if (!normalizedSearchQuery) {
      return (
        <div className="search-mode-surface">
          <div className="search-hint">
            <span className="search-hint-icon">
              <Search size={64} aria-hidden />
            </span>
            <p>
              Сузьте область поиска в пределах потока, нажав на его название, или
              нажмите{" "}
              <span className="search-inline-icon" aria-label="иконку планеты" role="img">
                <Globe2 size={15} aria-hidden />
              </span>
              , чтобы искать среди всех доступных статей.
            </p>
          </div>
        </div>
      );
    }

    if (searchLoading) {
      return (
        <div className="search-mode-surface">
          <div className="search-hint search-hint-muted">
            <span className="search-hint-icon">
              <Search size={48} aria-hidden />
            </span>
            <p>Ищем материалы...</p>
          </div>
        </div>
      );
    }

    if (searchError) {
      return (
        <div className="search-mode-surface">
          <div className="search-hint search-hint-muted">
            <span className="search-hint-icon">
              <Search size={48} aria-hidden />
            </span>
            <p>{errorMessage(searchError)}</p>
          </div>
        </div>
      );
    }

    if (!searchResults.length) {
      return (
        <div className="search-mode-surface">
          <div className="search-hint search-hint-muted">
            <span className="search-hint-icon">
              <Search size={48} aria-hidden />
            </span>
            <p>Ничего не найдено. Попробуйте другой запрос или выберите другую область поиска.</p>
          </div>
        </div>
      );
    }

    return (
      <div className="search-mode-surface">
        <div className="search-results" aria-live="polite">
          {searchResults.map((item) => (
            <ArticleCard
              key={item.id}
              item={item}
              isSaving={
                item.storage_mode === "local"
                  ? toggleLocalSearchSavedMutation.isPending &&
                    toggleLocalSearchSavedMutation.variables === item.id
                  : toggleSearchSavedMutation.isPending &&
                    toggleSearchSavedMutation.variables?.id === item.id
              }
              onToggleSaved={(nextItem) => {
                if (nextItem.storage_mode === "local") {
                  toggleLocalSearchSavedMutation.mutate(nextItem.id);
                } else {
                  toggleSearchSavedMutation.mutate(nextItem);
                }
              }}
            />
          ))}
        </div>
      </div>
    );
  }

  const shellStyle = {
    "--topbar-offset": headerOffset
  } as CSSProperties;

  return (
    <div className={`app-shell ${searchOpen ? "search-open" : ""}`} style={shellStyle}>
      {popoverOpen ? (
        <button
          className="filter-popover-backdrop"
          type="button"
          aria-label="Закрыть меню фильтра"
          onPointerDown={(event) => {
            event.preventDefault();
          }}
          onClick={(event) => {
            event.preventDefault();
            event.stopPropagation();
            closeFilterPopovers();
          }}
        />
      ) : null}
      <header
        className={`topbar ${hasFeedTools ? "with-feed-tools" : ""} ${searchOpen ? "search-mode" : ""}`}
        onClickCapture={closeFiltersFromHeaderClick}
      >
        <div className="topbar-main">
          {!searchOpen ? (
            <button
              className="icon-button"
              type="button"
              title="Открыть меню"
              aria-label="Открыть меню"
              onClick={() => {
                if (drawerOpen) {
                  setDrawerOpen(false);
                } else {
                  openDrawer();
                }
              }}
            >
              <Menu size={22} aria-hidden />
            </button>
          ) : null}

          <div className="topbar-title">
            {searchOpen && searchAvailable ? (
              <label className="search-field">
                <Search size={18} aria-hidden />
                <input
                  ref={searchInputRef}
                  type="search"
                  value={searchQuery}
                  onChange={(event) => setSearchQuery(event.target.value)}
                  placeholder={searchPlaceholder}
                />
              </label>
            ) : (
              <span>{pageTitle}</span>
            )}
          </div>

          <button
            className="icon-button"
            type="button"
            title={searchOpen ? "Закрыть поиск" : "Поиск"}
            aria-label={searchOpen ? "Закрыть поиск" : "Поиск"}
            disabled={!searchAvailable}
            onClick={() => {
              if (searchOpen) {
                closeSearch();
              } else {
                openSearch();
              }
            }}
          >
            {searchOpen ? <X size={21} aria-hidden /> : <Search size={21} aria-hidden />}
          </button>
        </div>

        {hasFeedTools ? (
          <div className="feed-toolbar">
            <div
              className={`feed-scroll-track ${feedStripScroll.canScroll ? "active" : ""} ${feedStripDragging ? "dragging" : ""}`}
              ref={feedScrollTrackRef}
              aria-hidden={!feedStripScroll.canScroll}
              onPointerDown={handleFeedScrollTrackPointerDown}
              onPointerMove={handleFeedScrollTrackPointerMove}
              onPointerUp={handleFeedScrollTrackPointerEnd}
              onPointerCancel={handleFeedScrollTrackPointerEnd}
            >
              <button
                className="feed-scroll-thumb"
                type="button"
                tabIndex={-1}
                style={{
                  left: `${feedStripScroll.progress * (100 - feedStripScroll.thumbWidth)}%`,
                  width: `${feedStripScroll.thumbWidth}%`
                }}
                aria-label="Прокрутка потоков"
              />
            </div>
            <div className="feed-toolbar-row">
              <button
                className={`icon-button filter-toggle ${filtersOpen || (searchOpen && searchScope.type === "all") ? "active" : ""}`}
                type="button"
                title={searchOpen ? "Все доступные статьи" : "Фильтры"}
                aria-label={searchOpen ? "Искать среди всех доступных статей" : "Фильтры"}
                aria-expanded={searchOpen ? undefined : filtersOpen}
                aria-pressed={searchOpen ? searchScope.type === "all" : undefined}
                onClick={() => {
                  if (searchOpen) {
                    setSearchScope({ type: "all" });
                    return;
                  }

                  setFiltersOpen((open) => !open);
                }}
              >
                {searchOpen ? <Globe2 size={19} aria-hidden /> : <Filter size={19} aria-hidden />}
              </button>
              <div
                className={`feed-strip-frame ${feedStripScroll.canScrollLeft ? "can-scroll-left" : ""} ${feedStripScroll.canScrollRight ? "can-scroll-right" : ""}`}
              >
                <button
                  className="feed-strip-arrow feed-strip-arrow-left"
                  type="button"
                  title="Назад"
                  aria-label="Прокрутить потоки назад"
                  disabled={!feedStripScroll.canScrollLeft}
                  onPointerDown={(event) => {
                    event.preventDefault();
                    startFeedStripHold(-1);
                  }}
                  onPointerUp={stopFeedStripHold}
                  onPointerCancel={stopFeedStripHold}
                  onPointerLeave={stopFeedStripHold}
                  onClick={(event) => event.preventDefault()}
                >
                  <ChevronLeft size={17} aria-hidden />
                </button>

                <div
                  className="feed-strip"
                  ref={feedStripRef}
                  aria-label="Быстрое переключение потоков"
                  onWheel={handleFeedStripWheel}
                >
                  {(feedsQuery.data ?? []).map((feed) =>
                    searchOpen ? (
                      <button
                        className={`feed-pill ${searchScope.type === "feed" && searchScope.feedId === feed.id ? "active" : ""}`}
                        key={feed.id}
                        ref={(node) => {
                          feedPillRefs.current[feed.id] = node;
                        }}
                        type="button"
                        style={feedPillStyle(feed.theme_color)}
                        onClick={() => setSearchScope({ type: "feed", feedId: feed.id })}
                      >
                        <span className="feed-pill-label">{feed.name}</span>
                      </button>
                    ) : (
                      <NavLink
                        className={({ isActive }) =>
                          `feed-pill ${isActive ? "active" : ""}`
                        }
                        key={feed.id}
                        ref={(node) => {
                          feedPillRefs.current[feed.id] = node;
                        }}
                        to={`/feeds/${feed.id}${location.search}`}
                        style={feedPillStyle(feed.theme_color)}
                      >
                        <span className="feed-pill-label">{feed.name}</span>
                      </NavLink>
                    )
                  )}
                </div>

                <button
                  className="feed-strip-arrow feed-strip-arrow-right"
                  type="button"
                  title="Вперед"
                  aria-label="Прокрутить потоки вперед"
                  disabled={!feedStripScroll.canScrollRight}
                  onPointerDown={(event) => {
                    event.preventDefault();
                    startFeedStripHold(1);
                  }}
                  onPointerUp={stopFeedStripHold}
                  onPointerCancel={stopFeedStripHold}
                  onPointerLeave={stopFeedStripHold}
                  onClick={(event) => event.preventDefault()}
                >
                  <ChevronRight size={17} aria-hidden />
                </button>
              </div>
            </div>

            {filtersOpen ? (
              <div
                className="filter-panel"
                aria-label="Фильтры ленты"
              >
                <div className="filter-row">
                  <span className="filter-row-label">Диапазон дат</span>
                  <div className="filter-control">
                    <button
                      className="filter-select"
                      type="button"
                      onClick={() => {
                        setDateMenuOpen((open) => !open);
                        setCategoryMenuOpen(false);
                      }}
                    >
                      <CalendarDays size={16} aria-hidden />
                      {dateFilter.label}
                      <ChevronDown size={16} aria-hidden />
                    </button>
                    {dateMenuOpen ? (
                      <div className="filter-popover date-popover" role="menu">
                        <button type="button" onClick={() => applyDatePreset("today")}>
                          Сегодня
                        </button>
                        <button type="button" onClick={() => applyDatePreset("7d")}>
                          Последние 7 дней
                        </button>
                        <button type="button" onClick={() => applyDatePreset("30d")}>
                          Последние 30 дней
                        </button>
                        <button type="button" onClick={() => applyDatePreset("all")}>
                          Все даты
                        </button>
                        <div className="date-input-grid">
                          <label>
                            С
                            <button
                              className="date-input-button"
                              type="button"
                              onClick={() => openDatePicker(dateFromInputRef.current)}
                            >
                              {draftDateFrom || "дд . мм . гггг"}
                              <CalendarDays size={16} aria-hidden />
                            </button>
                            <span className="date-native-holder">
                              <input
                                ref={dateFromInputRef}
                                type="date"
                                value={draftDateFrom}
                                max={draftDateTo || undefined}
                                onChange={(event) => setDraftDateFrom(event.target.value)}
                              />
                            </span>
                          </label>
                          <label>
                            По
                            <button
                              className="date-input-button"
                              type="button"
                              onClick={() => openDatePicker(dateToInputRef.current)}
                            >
                              {draftDateTo || "дд . мм . гггг"}
                              <CalendarDays size={16} aria-hidden />
                            </button>
                            <span className="date-native-holder">
                              <input
                                ref={dateToInputRef}
                                type="date"
                                value={draftDateTo}
                                min={draftDateFrom || undefined}
                                max={localDateString(new Date())}
                                onChange={(event) => setDraftDateTo(event.target.value)}
                              />
                            </span>
                          </label>
                        </div>
                        <button
                          className="primary-button compact"
                          type="button"
                          disabled={!draftDateFrom && !draftDateTo}
                          onClick={applyCustomDate}
                        >
                          Применить
                        </button>
                      </div>
                    ) : null}
                  </div>
                </div>

                <div className="filter-row">
                  <span className="filter-row-label">Темы</span>
                  <div className="filter-control">
                    <button
                      className="filter-select"
                      type="button"
                      onClick={() => {
                        setCategoryMenuOpen((open) => !open);
                        setDateMenuOpen(false);
                      }}
                    >
                      {categoryLabel}
                      <ChevronDown size={16} aria-hidden />
                    </button>
                    {categoryMenuOpen ? (
                      <div className="filter-popover category-popover" role="menu">
                        <div className="category-options">
                          {categoriesLoading ? (
                            <p className="category-empty">Загружаем темы...</p>
                          ) : null}
                          {!categoriesLoading && !categories.length ? (
                            <p className="category-empty">В этом потоке пока нет тем.</p>
                          ) : null}
                          {!categoriesLoading
                            ? categories.map((category) => (
                              <label className="checkbox-row" key={category.id}>
                                <input
                                  type="checkbox"
                                  checked={draftCategories.includes(category.slug)}
                                  onChange={() => toggleDraftCategory(category.slug)}
                                />
                                <span className="custom-checkbox">
                                  {draftCategories.includes(category.slug) ? (
                                    <Check size={14} aria-hidden />
                                  ) : null}
                                </span>
                                {category.name}
                              </label>
                            ))
                            : null}
                        </div>
                        <div className="popover-actions">
                          <button
                            className="text-button"
                            type="button"
                            onClick={() => setDraftCategories([])}
                          >
                            Сбросить
                          </button>
                          <button
                            className="primary-button compact"
                            type="button"
                            onClick={applyCategories}
                          >
                            Применить
                          </button>
                        </div>
                      </div>
                    ) : null}
                  </div>
                </div>
              </div>
            ) : null}
          </div>
        ) : null}
        {searchOpen ? renderSearchSurface() : null}
      </header>

      <button
        className={`drawer-backdrop ${drawerOpen ? "visible" : ""}`}
        type="button"
        aria-label="Закрыть меню"
        onClick={closeDrawer}
      />

      <aside className={`drawer ${drawerOpen ? "open" : ""}`} aria-label="Навигация">
        <nav className="drawer-nav">
          <NavLink
            className={({ isActive }) => `drawer-link ${isActive ? "active" : ""}`}
            to="/feeds"
            onClick={closeDrawer}
            end
          >
            <Rss size={18} aria-hidden />
            Мои ленты
          </NavLink>
          <NavLink
            className={({ isActive }) => `drawer-link ${isActive ? "active" : ""}`}
            to="/catalog"
            onClick={closeDrawer}
          >
            <Compass size={18} aria-hidden />
            Каталог
          </NavLink>
          <NavLink
            className={({ isActive }) => `drawer-link ${isActive ? "active" : ""}`}
            to="/saved"
            onClick={closeDrawer}
          >
            <Bookmark size={18} aria-hidden />
            Избранное
          </NavLink>
          <NavLink
            className={({ isActive }) => `drawer-link ${isActive ? "active" : ""}`}
            to="/sources"
            onClick={closeDrawer}
          >
            <Database size={18} aria-hidden />
            Источники
          </NavLink>

          <div className="drawer-section-title">Потоки</div>
          {feedsQuery.data?.length ? (
            feedsQuery.data.map((feed) => (
              <NavLink
                className={({ isActive }) =>
                  `drawer-link feed-link ${isActive ? "active" : ""}`
                }
                key={feed.id}
                to={`/feeds/${feed.id}`}
                onClick={closeDrawer}
              >
                <span
                  className="feed-dot"
                  style={{ backgroundColor: feed.theme_color || "#2563eb" }}
                />
                <span>{feed.name}</span>
              </NavLink>
            ))
          ) : (
            <span className="drawer-hint">Пока нет лент</span>
          )}
        </nav>

        <div className="drawer-user">
          <div className="user-summary">
            <UserCircle size={26} aria-hidden />
            <div>
              <strong>{user?.username ?? "Пользователь"}</strong>
              <span>{user?.email ?? "Аккаунт"}</span>
            </div>
          </div>
          <button className="text-button danger" type="button" onClick={handleLogout}>
            <LogOut size={17} aria-hidden />
            Выйти
          </button>
        </div>
      </aside>

      <main className="app-main">
        <Outlet />
      </main>

      {nextFeed && isFeedRoute && !searchOpen ? (
        <div className="feed-switch-zone" style={feedSwitchIndicatorStyle}>
          <div
            className={`feed-switch-indicator ${feedSwitchProgress > 0 ? "visible" : ""}`}
            aria-hidden={feedSwitchProgress <= 0}
          >
            <strong>{nextFeed.name}</strong>
            <span className="feed-switch-circle">
              <ArrowDown size={24} aria-hidden />
            </span>
          </div>
        </div>
      ) : null}

      <ArticleReaderModal item={readerItem} onClose={closeReader} />
    </div>
  );
}
