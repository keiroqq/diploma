import { useEffect, useMemo, useRef } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Bookmark,
  Compass,
  LogOut,
  Menu,
  Rss,
  Search,
  UserCircle,
  X
} from "lucide-react";
import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";

import { getMe, listFeeds } from "../api/client";
import { useAuthStore } from "../store/auth";
import { useUiStore } from "../store/ui";

export function AppShell() {
  const location = useLocation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const token = useAuthStore((state) => state.token);
  const user = useAuthStore((state) => state.user);
  const setUser = useAuthStore((state) => state.setUser);
  const logout = useAuthStore((state) => state.logout);
  const drawerOpen = useUiStore((state) => state.drawerOpen);
  const searchOpen = useUiStore((state) => state.searchOpen);
  const searchQuery = useUiStore((state) => state.searchQuery);
  const toggleDrawer = useUiStore((state) => state.toggleDrawer);
  const setDrawerOpen = useUiStore((state) => state.setDrawerOpen);
  const setSearchOpen = useUiStore((state) => state.setSearchOpen);
  const setSearchQuery = useUiStore((state) => state.setSearchQuery);
  const closeSearch = useUiStore((state) => state.closeSearch);
  const searchInputRef = useRef<HTMLInputElement>(null);

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
  const currentFeed = feedsQuery.data?.find((feed) => feed.id === feedId);

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

  function handleLogout() {
    queryClient.clear();
    logout();
    navigate("/login", { replace: true });
  }

  function closeDrawer() {
    setDrawerOpen(false);
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <button
          className="icon-button"
          type="button"
          title="Открыть меню"
          aria-label="Открыть меню"
          onClick={toggleDrawer}
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
