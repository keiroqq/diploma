import { useMemo } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { RefreshCw, Rss } from "lucide-react";
import { useParams, useSearchParams } from "react-router-dom";

import {
  getFeed,
  listCategories,
  listFeedItems,
  refreshFeed,
  saveItem,
  unsaveItem
} from "../api/client";
import type { Item } from "../api/types";
import { ArticleCard } from "../components/ArticleCard";
import { EmptyState } from "../components/EmptyState";
import { ErrorState } from "../components/ErrorState";
import { LoadingState } from "../components/LoadingState";
import { useUiStore } from "../store/ui";
import { errorMessage } from "../utils/errors";
import { filterItemsByQuery } from "../utils/items";

export function FeedPage() {
  const { feedId = "" } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const queryClient = useQueryClient();
  const searchQuery = useUiStore((state) => state.searchQuery);
  const category = searchParams.get("category") ?? "";

  const feedQuery = useQuery({
    queryKey: ["feed", feedId],
    queryFn: () => getFeed(feedId),
    enabled: Boolean(feedId)
  });

  const categoriesQuery = useQuery({
    queryKey: ["categories"],
    queryFn: listCategories
  });

  const todayItemsQuery = useQuery({
    queryKey: ["feedItems", feedId, "today", category],
    queryFn: () => listFeedItems(feedId, category || undefined, "today"),
    enabled: Boolean(feedId)
  });

  const todayItems = todayItemsQuery.data?.items ?? [];
  const archiveItemsQuery = useQuery({
    queryKey: ["feedItems", feedId, "archive", category],
    queryFn: () => listFeedItems(feedId, category || undefined, "archive"),
    enabled: Boolean(feedId) && todayItemsQuery.isSuccess && todayItems.length === 0
  });

  const refreshMutation = useMutation({
    mutationFn: () => refreshFeed(feedId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feedItems", feedId] });
      queryClient.invalidateQueries({ queryKey: ["saved"] });
    }
  });

  const toggleSavedMutation = useMutation({
    mutationFn: (item: Item) => (item.is_saved ? unsaveItem(item.id) : saveItem(item.id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feedItems", feedId] });
      queryClient.invalidateQueries({ queryKey: ["saved"] });
    }
  });

  const archiveItems = archiveItemsQuery.data?.items ?? [];
  const isArchiveFallback = todayItemsQuery.isSuccess && todayItems.length === 0;
  const items = isArchiveFallback ? archiveItems : todayItems;
  const visibleItems = useMemo(
    () => filterItemsByQuery(items, searchQuery),
    [items, searchQuery]
  );

  function setCategory(slug: string) {
    const next = new URLSearchParams(searchParams);

    if (slug) {
      next.set("category", slug);
    } else {
      next.delete("category");
    }

    setSearchParams(next, { replace: true });
  }

  if (feedQuery.isLoading || todayItemsQuery.isLoading || archiveItemsQuery.isLoading) {
    return <LoadingState label="Загружаем материалы" />;
  }

  if (feedQuery.isError) {
    return <ErrorState message={errorMessage(feedQuery.error)} />;
  }

  const feed = feedQuery.data;

  return (
    <section className="page-section">
      <div className="section-heading feed-heading">
        <div>
          <p className="eyebrow">{isArchiveFallback ? "Последние" : "Сегодня"}</p>
          <h1>{feed?.name ?? "Поток"}</h1>
          {feed?.description ? <p>{feed.description}</p> : null}
        </div>
        <button
          className="secondary-button"
          type="button"
          disabled={refreshMutation.isPending}
          onClick={() => refreshMutation.mutate()}
        >
          <RefreshCw size={17} aria-hidden className={refreshMutation.isPending ? "spin" : ""} />
          {refreshMutation.isPending ? "Обновляем" : "Обновить"}
        </button>
      </div>

      <div className="filter-rail" aria-label="Фильтр по категориям">
        <button
          className={`chip-button ${!category ? "active" : ""}`}
          type="button"
          onClick={() => setCategory("")}
        >
          Все
        </button>
        {(categoriesQuery.data ?? []).map((item) => (
          <button
            className={`chip-button ${category === item.slug ? "active" : ""}`}
            type="button"
            key={item.id}
            onClick={() => setCategory(item.slug)}
          >
            {item.name}
          </button>
        ))}
      </div>

      {refreshMutation.isError ? (
        <ErrorState title="Обновление не удалось" message={errorMessage(refreshMutation.error)} />
      ) : null}
      {todayItemsQuery.isError ? <ErrorState message={errorMessage(todayItemsQuery.error)} /> : null}
      {archiveItemsQuery.isError ? (
        <ErrorState message={errorMessage(archiveItemsQuery.error)} />
      ) : null}

      {!todayItemsQuery.isError && !archiveItemsQuery.isError && !items.length ? (
        <EmptyState
          icon={<Rss size={34} aria-hidden />}
          title="Материалов пока нет"
          description="Обновите поток: backend загрузит RSS и сохранит найденные статьи."
          action={
            <button
              className="primary-button"
              type="button"
              disabled={refreshMutation.isPending}
              onClick={() => refreshMutation.mutate()}
            >
              <RefreshCw size={18} aria-hidden className={refreshMutation.isPending ? "spin" : ""} />
              Обновить поток
            </button>
          }
        />
      ) : null}

      {items.length && !visibleItems.length ? (
        <EmptyState
          title="Ничего не найдено"
          description="Попробуйте другой запрос или снимите фильтр категории."
        />
      ) : null}

      {visibleItems.length ? (
        <div className="article-list">
          {visibleItems.map((item) => (
            <ArticleCard
              key={item.id}
              item={item}
              isSaving={
                toggleSavedMutation.isPending &&
                toggleSavedMutation.variables?.id === item.id
              }
              onToggleSaved={(nextItem) => toggleSavedMutation.mutate(nextItem)}
            />
          ))}
        </div>
      ) : null}
    </section>
  );
}
