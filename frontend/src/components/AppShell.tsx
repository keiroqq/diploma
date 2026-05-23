import { useEffect, useMemo, useRef, useState, type CSSProperties } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Bookmark,
  CalendarDays,
  Check,
  ChevronDown,
  Compass,
  Filter,
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

import { getMe, listCategories, listFeeds } from "../api/client";
import { useAuthStore } from "../store/auth";
import { useUiStore } from "../store/ui";
import {
  categoryFilterLabel,
  getDateFilter,
  getSelectedCategorySlugs,
  localDateString,
  type DatePreset
} from "../utils/filters";

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
  const setDrawerOpen = useUiStore((state) => state.setDrawerOpen);
  const setSearchOpen = useUiStore((state) => state.setSearchOpen);
  const setSearchQuery = useUiStore((state) => state.setSearchQuery);
  const closeSearch = useUiStore((state) => state.closeSearch);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const dateFromInputRef = useRef<HTMLInputElement>(null);
  const dateToInputRef = useRef<HTMLInputElement>(null);
  const feedPillRefs = useRef<Record<string, HTMLAnchorElement | null>>({});
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [dateMenuOpen, setDateMenuOpen] = useState(false);
  const [categoryMenuOpen, setCategoryMenuOpen] = useState(false);
  const [draftDateFrom, setDraftDateFrom] = useState("");
  const [draftDateTo, setDraftDateTo] = useState("");
  const [draftCategories, setDraftCategories] = useState<string[]>([]);

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

  const categoriesQuery = useQuery({
    queryKey: ["categories"],
    queryFn: listCategories,
    enabled: Boolean(token)
  });

  useEffect(() => {
    if (meQuery.data) {
      setUser(meQuery.data);
    }
  }, [meQuery.data, setUser]);

  const feedId = location.pathname.match(/^\/feeds\/([^/]+)/)?.[1];
  const isFeedRoute = Boolean(feedId);
  const currentFeed = feedsQuery.data?.find((feed) => feed.id === feedId);
  const dateFilter = getDateFilter(searchParams);
  const selectedCategories = getSelectedCategorySlugs(searchParams);
  const categories = categoriesQuery.data ?? [];
  const categoryLabel = categoryFilterLabel(categories, selectedCategories);
  const headerOffset = isFeedRoute ? "126px" : "64px";
  const popoverOpen = dateMenuOpen || categoryMenuOpen;

  const pageTitle = useMemo(() => {
    if (location.pathname === "/catalog") {
      return "Каталог";
    }
    if (location.pathname === "/saved") {
      return "Избранное";
    }
    if (feedId) {
      return currentFeed?.name ?? "Поток";
    }
    return "Мои ленты";
  }, [currentFeed?.name, feedId, location.pathname]);

  const searchAvailable =
    location.pathname.startsWith("/feeds/") || location.pathname === "/saved";

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
    if (!isFeedRoute) {
      setFiltersOpen(false);
      setDateMenuOpen(false);
      setCategoryMenuOpen(false);
    }
  }, [isFeedRoute]);

  useEffect(() => {
    if (feedId) {
      feedPillRefs.current[feedId]?.scrollIntoView({
        behavior: "smooth",
        block: "nearest",
        inline: "center"
      });
    }
  }, [feedId, feedsQuery.data?.length]);

  function handleLogout() {
    queryClient.clear();
    logout();
    navigate("/login", { replace: true });
  }

  function closeDrawer() {
    setDrawerOpen(false);
  }

  function openDrawer() {
    setFiltersOpen(false);
    closeFilterPopovers();
    setDrawerOpen(true);
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

      if (preset === "today") {
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

  const shellStyle = {
    "--topbar-offset": headerOffset
  } as CSSProperties;

  return (
    <div className="app-shell" style={shellStyle}>
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
        className={`topbar ${isFeedRoute ? "with-feed-tools" : ""}`}
        onClickCapture={closeFiltersFromHeaderClick}
      >
        <div className="topbar-main">
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

          <div className="topbar-title">
            {searchOpen && searchAvailable ? (
              <label className="search-field">
                <Search size={18} aria-hidden />
                <input
                  ref={searchInputRef}
                  type="search"
                  value={searchQuery}
                  onChange={(event) => setSearchQuery(event.target.value)}
                  placeholder="Поиск по текущей выдаче"
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
                setSearchOpen(true);
              }
            }}
          >
            {searchOpen ? <X size={21} aria-hidden /> : <Search size={21} aria-hidden />}
          </button>
        </div>

        {isFeedRoute ? (
          <div className="feed-toolbar">
            <div className="feed-toolbar-row">
              <button
                className={`icon-button filter-toggle ${filtersOpen ? "active" : ""}`}
                type="button"
                title="Фильтры"
                aria-label="Фильтры"
                aria-expanded={filtersOpen}
                onClick={() => setFiltersOpen((open) => !open)}
              >
                <Filter size={19} aria-hidden />
              </button>
              <div className="feed-strip" aria-label="Быстрое переключение потоков">
                {(feedsQuery.data ?? []).map((feed) => (
                  <NavLink
                    className={({ isActive }) =>
                      `feed-pill ${isActive ? "active" : ""}`
                    }
                    key={feed.id}
                    ref={(node) => {
                      feedPillRefs.current[feed.id] = node;
                    }}
                    to={`/feeds/${feed.id}${location.search}`}
                  >
                    {feed.name}
                  </NavLink>
                ))}
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
                          {categories.map((category) => (
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
                          ))}
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
      </header>

      <button
        className={`drawer-backdrop ${drawerOpen ? "visible" : ""}`}
        type="button"
        aria-label="Закрыть меню"
        onClick={closeDrawer}
      />

      <aside className={`drawer ${drawerOpen ? "open" : ""}`} aria-label="Навигация">
        <div className="drawer-brand">
          <span className="brand-mark">
            <Rss size={22} aria-hidden />
          </span>
          <div>
            <strong>Content Digest</strong>
            <span>RSS-потоки</span>
          </div>
        </div>

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
    </div>
  );
}
