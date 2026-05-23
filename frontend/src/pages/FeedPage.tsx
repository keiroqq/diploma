import { useMemo } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { RefreshCw, Rss, Trash2 } from "lucide-react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";

import {
  deleteFeed,
  getFeed,
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
import { getDateFilter, getSelectedCategorySlugs } from "../utils/filters";
import { filterItemsByQuery } from "../utils/items";

export function FeedPage() {
  const { feedId = "" } = useParams();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const searchQuery = useUiStore((state) => state.searchQuery);
  const dateFilter = getDateFilter(searchParams);
  const selectedCategories = getSelectedCategorySlugs(searchParams);

  const feedQuery = useQuery({
    queryKey: ["feed", feedId],
    queryFn: () => getFeed(feedId),
    enabled: Boolean(feedId)
  });

  const itemsQuery = useQuery({
    queryKey: [
      "feedItems",
      feedId,
      dateFilter.mode,
      dateFilter.dateFrom,
      dateFilter.dateTo,
      selectedCategories
    ],
    queryFn: () =>
      listFeedItems(feedId, {
        mode: dateFilter.mode,
        dateFrom: dateFilter.dateFrom,
        dateTo: dateFilter.dateTo,
        categories: selectedCategories
      }),
    enabled: Boolean(feedId)
  });

  const refreshMutation = useMutation({
    mutationFn: () => refreshFeed(feedId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feedItems", feedId] });
      queryClient.invalidateQueries({ queryKey: ["saved"] });
    }
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteFeed(feedId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feeds"] });
      navigate("/feeds", { replace: true });
    }
  });

  const toggleSavedMutation = useMutation({
    mutationFn: (item: Item) => (item.is_saved ? unsaveItem(item.id) : saveItem(item.id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feedItems", feedId] });
      queryClient.invalidateQueries({ queryKey: ["saved"] });
    }
  });

  const items = itemsQuery.data?.items ?? [];
  const visibleItems = useMemo(
    () => filterItemsByQuery(items, searchQuery),
    [items, searchQuery]
  );

  if (feedQuery.isLoading || itemsQuery.isLoading) {
    return <LoadingState label="Загружаем материалы" />;
  }

  if (feedQuery.isError) {
    return <ErrorState message={errorMessage(feedQuery.error)} />;
  }

  const feed = feedQuery.data;
  const feedName = feed?.name ?? "Поток";

  return (
    <section className="page-section">
      <div className="section-heading feed-heading">
        <div>
          <p className="eyebrow">{dateFilter.label}</p>
          <h1>{feedName}</h1>
          {feed?.description ? <p>{feed.description}</p> : null}
        </div>
        <div className="feed-heading-actions">
          <button
            className="secondary-button danger-action"
            type="button"
            disabled={deleteMutation.isPending}
            onClick={() => {
              if (window.confirm(`Удалить поток "${feedName}"?`)) {
                deleteMutation.mutate();
              }
            }}
          >
            <Trash2 size={17} aria-hidden />
            {deleteMutation.isPending ? "Удаляем" : "Удалить"}
          </button>
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
      </div>

      {refreshMutation.isError ? (
        <ErrorState title="Обновление не удалось" message={errorMessage(refreshMutation.error)} />
      ) : null}
      {deleteMutation.isError ? (
        <ErrorState title="Удаление не удалось" message={errorMessage(deleteMutation.error)} />
      ) : null}
      {itemsQuery.isError ? <ErrorState message={errorMessage(itemsQuery.error)} /> : null}

      {!itemsQuery.isError && !items.length ? (
        <EmptyState
          icon={<Rss size={34} aria-hidden />}
          title="Материалов пока нет"
          description="Для выбранных фильтров материалов нет. Можно изменить дату, темы или обновить поток."
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
