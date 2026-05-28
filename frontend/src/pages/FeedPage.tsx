import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Pencil, Plus, RefreshCw, Rss, Trash2 } from "lucide-react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";

import {
  deleteFeed,
  getFeed,
  listFeedSources,
  listFeedItems,
  refreshSource,
  removeFeedSource,
  refreshFeed,
  saveItem,
  unsaveItem
} from "../api/client";
import type { Item } from "../api/types";
import { ArticleCard } from "../components/ArticleCard";
import { EmptyState } from "../components/EmptyState";
import { ErrorState } from "../components/ErrorState";
import { FeedEditDialog } from "../components/FeedEditDialog";
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
  const [editingOpen, setEditingOpen] = useState(false);
  const [sourcesOpen, setSourcesOpen] = useState(false);
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

  const feedSourcesQuery = useQuery({
    queryKey: ["feedSources", feedId],
    queryFn: () => listFeedSources(feedId),
    enabled: Boolean(feedId) && sourcesOpen
  });

  const refreshMutation = useMutation({
    mutationFn: () => refreshFeed(feedId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feedItems", feedId] });
      queryClient.invalidateQueries({ queryKey: ["feedCategories", feedId] });
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

  const refreshSourceMutation = useMutation({
    mutationFn: refreshSource,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feedItems", feedId] });
      queryClient.invalidateQueries({ queryKey: ["feedCategories", feedId] });
      queryClient.invalidateQueries({ queryKey: ["feedSources", feedId] });
      queryClient.invalidateQueries({ queryKey: ["saved"] });
    }
  });

  const removeSourceMutation = useMutation({
    mutationFn: (sourceId: string) => removeFeedSource(feedId, sourceId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feedSources", feedId] });
      queryClient.invalidateQueries({ queryKey: ["feedItems", feedId] });
      queryClient.invalidateQueries({ queryKey: ["feedCategories", feedId] });
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
      <FeedEditDialog
        feed={editingOpen && feed ? feed : null}
        onClose={() => setEditingOpen(false)}
      />

      <div className="section-heading feed-heading">
        <div className="feed-heading-content">
          <div className="feed-heading-meta">
            <p className="eyebrow">{dateFilter.label}</p>
            <div className="feed-heading-actions">
              <button
                className={`icon-button ${sourcesOpen ? "active" : ""}`}
                type="button"
                title="Источники"
                aria-label={`Источники ${feedName}`}
                aria-expanded={sourcesOpen}
                onClick={() => setSourcesOpen((open) => !open)}
              >
                <Rss size={17} aria-hidden />
              </button>
              <button
                className="icon-button"
                type="button"
                title="Изменить"
                aria-label={`Изменить ${feedName}`}
                onClick={() => setEditingOpen(true)}
              >
                <Pencil size={17} aria-hidden />
              </button>
              <button
                className="icon-button danger-button"
                type="button"
                title="Удалить"
                aria-label={`Удалить ${feedName}`}
                disabled={deleteMutation.isPending}
                onClick={() => {
                  if (window.confirm(`Удалить поток "${feedName}"?`)) {
                    deleteMutation.mutate();
                  }
                }}
              >
                <Trash2 size={17} aria-hidden />
              </button>
              <button
                className="icon-button"
                type="button"
                title="Обновить"
                aria-label={`Обновить ${feedName}`}
                disabled={refreshMutation.isPending}
                onClick={() => refreshMutation.mutate()}
              >
                <RefreshCw size={17} aria-hidden className={refreshMutation.isPending ? "spin" : ""} />
              </button>
            </div>
          </div>

          <h1>{feedName}</h1>
          {feed?.description ? <p>{feed.description}</p> : null}
        </div>
      </div>

      {refreshMutation.isError ? (
        <ErrorState title="Обновление не удалось" message={errorMessage(refreshMutation.error)} />
      ) : null}
      {deleteMutation.isError ? (
        <ErrorState title="Удаление не удалось" message={errorMessage(deleteMutation.error)} />
      ) : null}
      {refreshSourceMutation.isError ? (
        <ErrorState title="Источник не обновлен" message={errorMessage(refreshSourceMutation.error)} />
      ) : null}
      {removeSourceMutation.isError ? (
        <ErrorState title="Источник не отключен" message={errorMessage(removeSourceMutation.error)} />
      ) : null}
      {itemsQuery.isError ? <ErrorState message={errorMessage(itemsQuery.error)} /> : null}

      {sourcesOpen ? (
        <section className="feed-sources-panel" aria-label="Источники ленты">
          <div className="feed-sources-heading">
            <div>
              <h2>Источники</h2>
              <p>RSS-источники, подключенные к этому потоку.</p>
            </div>
            <button
              className="icon-button"
              type="button"
              title="Добавить источник"
              aria-label="Добавить источник"
              onClick={() => navigate(`/catalog?feed_id=${feedId}`)}
            >
              <Plus size={18} aria-hidden />
            </button>
          </div>

          {feedSourcesQuery.isLoading ? (
            <p className="feed-sources-status">Загружаем источники...</p>
          ) : null}
          {feedSourcesQuery.isError ? (
            <p className="feed-sources-status error">{errorMessage(feedSourcesQuery.error)}</p>
          ) : null}
          {!feedSourcesQuery.isLoading && !feedSourcesQuery.isError && !feedSourcesQuery.data?.length ? (
            <p className="feed-sources-status">В ленте пока нет источников.</p>
          ) : null}

          {feedSourcesQuery.data?.length ? (
            <div className="feed-source-list">
              {feedSourcesQuery.data.map((link) => {
                const source = link.source;
                const sourceName = source?.name ?? "Источник";
                const sourceURL = source?.feed_url ?? "";
                const storageMode = source?.storage_mode === "local" ? "Локально" : "Сервер";

                return (
                  <article className="feed-source-row" key={link.id}>
                    <div className="feed-source-main">
                      <strong>{sourceName}</strong>
                      {sourceURL ? <span>{sourceURL}</span> : null}
                      <div className="feed-source-meta">
                        <span>{storageMode}</span>
                        {source?.status ? <span>{source.status}</span> : null}
                      </div>
                    </div>
                    <div className="feed-source-actions">
                      <button
                        className="icon-button"
                        type="button"
                        title="Обновить источник"
                        aria-label={`Обновить ${sourceName}`}
                        disabled={
                          refreshSourceMutation.isPending ||
                          source?.storage_mode === "local"
                        }
                        onClick={() => refreshSourceMutation.mutate(link.source_id)}
                      >
                        <RefreshCw
                          size={17}
                          aria-hidden
                          className={
                            refreshSourceMutation.isPending &&
                            refreshSourceMutation.variables === link.source_id
                              ? "spin"
                              : ""
                          }
                        />
                      </button>
                      <button
                        className="icon-button danger-button"
                        type="button"
                        title="Отключить источник"
                        aria-label={`Отключить ${sourceName}`}
                        disabled={removeSourceMutation.isPending}
                        onClick={() => {
                          if (window.confirm(`Отключить источник "${sourceName}" от потока?`)) {
                            removeSourceMutation.mutate(link.source_id);
                          }
                        }}
                      >
                        <Trash2 size={17} aria-hidden />
                      </button>
                    </div>
                  </article>
                );
              })}
            </div>
          ) : null}
        </section>
      ) : null}

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
