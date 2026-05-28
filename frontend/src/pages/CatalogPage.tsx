import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Check, Loader2, Palette, Plus, Rss } from "lucide-react";
import { useNavigate, useSearchParams } from "react-router-dom";

import {
  addFeedSource,
  connectCatalogSources,
  createFeed,
  deleteFeed,
  listCatalogTopics,
  listSources,
  previewSourceItems,
  refreshFeed
} from "../api/client";
import type { Source } from "../api/types";
import { ErrorState } from "../components/ErrorState";
import { LoadingState } from "../components/LoadingState";
import { errorMessage } from "../utils/errors";
import { cacheLocalSourceItems } from "../utils/localItems";
import {
  catalogSourceEnabled,
  catalogSourceTitle,
  loadSourcePreferences
} from "../utils/sourcePreferences";

function catalogSelectionKey(sourceID: string) {
  return `catalog:${sourceID}`;
}

function customSelectionKey(sourceID: string) {
  return `source:${sourceID}`;
}

export function CatalogPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const queryClient = useQueryClient();
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [feedName, setFeedName] = useState("Моя IT-лента");
  const [themeColor, setThemeColor] = useState("#2563eb");
  const [sourcePreferences] = useState(loadSourcePreferences);
  const targetFeedId = searchParams.get("feed_id");

  const topicsQuery = useQuery({
    queryKey: ["catalogTopics"],
    queryFn: listCatalogTopics
  });

  const sourcesQuery = useQuery({
    queryKey: ["sources"],
    queryFn: listSources
  });

  const topics = useMemo(() => {
    return (topicsQuery.data ?? [])
      .map((topic) => ({
        ...topic,
        sources: topic.sources
          .filter((source) => catalogSourceEnabled(source, sourcePreferences))
          .map((source) => ({
            ...source,
            title: catalogSourceTitle(source, sourcePreferences)
          }))
      }))
      .filter((topic) => topic.sources.length > 0);
  }, [sourcePreferences, topicsQuery.data]);

  const customSources = useMemo(() => {
    return (sourcesQuery.data ?? [])
      .filter(
        (source) =>
          source.storage_mode === "local"
          && !source.is_public
          && source.status !== "disabled"
      )
      .sort((left, right) => left.name.localeCompare(right.name, "ru"));
  }, [sourcesQuery.data]);

  const selectedCatalogSources = useMemo(() => {
    const topics = topicsQuery.data ?? [];
    const selectedIds = selected;

    return topics
      .flatMap((topic) => topic.sources)
      .filter((source) => catalogSourceEnabled(source, sourcePreferences))
      .map((source) => ({
        ...source,
        title: catalogSourceTitle(source, sourcePreferences)
      }))
      .filter((source) => selectedIds.has(catalogSelectionKey(source.id)));
  }, [selected, sourcePreferences, topicsQuery.data]);

  const selectedCustomSources = useMemo(() => {
    const selectedIds = selected;

    return customSources.filter((source) => selectedIds.has(customSelectionKey(source.id)));
  }, [customSources, selected]);

  const createFromCatalogMutation = useMutation({
    mutationFn: async () => {
      const sourceIds = selectedCatalogSources.map((source) => source.id);
      const sourceTitles = [
        ...selectedCatalogSources.map((source) => source.title),
        ...selectedCustomSources.map((source) => source.name)
      ].join(", ");

      async function addCustomSources(feedId: string, sources: Source[]) {
        for (const [index, source] of sources.entries()) {
          await addFeedSource(feedId, source.id, sources.length - index);
          const preview = await previewSourceItems(source.id);
          await cacheLocalSourceItems(feedId, source, preview.items);
        }
      }

      if (targetFeedId) {
        if (sourceIds.length) {
          await connectCatalogSources(targetFeedId, sourceIds);
          await refreshFeed(targetFeedId).catch(() => undefined);
        }
        if (selectedCustomSources.length) {
          await addCustomSources(targetFeedId, selectedCustomSources);
        }
        return { id: targetFeedId };
      }

      const feed = await createFeed({
        name: feedName.trim() || "Новая лента",
        description: sourceTitles ? `Источники: ${sourceTitles}` : "Лента из каталога",
        icon: "rss",
        theme_color: themeColor,
        layout_type: "cards"
      });

      try {
        if (sourceIds.length) {
          await connectCatalogSources(feed.id, sourceIds);
        }
        if (selectedCustomSources.length) {
          await addCustomSources(feed.id, selectedCustomSources);
        }
      } catch (error) {
        await deleteFeed(feed.id).catch(() => undefined);
        throw error;
      }

      if (sourceIds.length) {
        await refreshFeed(feed.id).catch(() => undefined);
      }

      return feed;
    },
    onSuccess: async (feed) => {
      await queryClient.invalidateQueries({ queryKey: ["feeds"] });
      await queryClient.invalidateQueries({ queryKey: ["feedCategories", feed.id] });
      await queryClient.invalidateQueries({ queryKey: ["feedSources", feed.id] });
      navigate(`/feeds/${feed.id}`);
    }
  });

  function toggleSource(selectionKey: string) {
    setSelected((previous) => {
      const next = new Set(previous);

      if (next.has(selectionKey)) {
        next.delete(selectionKey);
      } else {
        next.add(selectionKey);
      }

      return next;
    });
  }

  if (topicsQuery.isLoading || sourcesQuery.isLoading) {
    return <LoadingState label="Загружаем каталог" />;
  }

  if (topicsQuery.isError) {
    return <ErrorState message={errorMessage(topicsQuery.error)} />;
  }

  if (sourcesQuery.isError) {
    return <ErrorState message={errorMessage(sourcesQuery.error)} />;
  }

  return (
    <section className="page-section catalog-page">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Каталог</p>
          <h1>{targetFeedId ? "Добавить источники" : "Темы и источники"}</h1>
        </div>
      </div>

      <div className={`catalog-builder ${targetFeedId ? "existing-feed" : ""}`}>
        {!targetFeedId ? (
          <label>
            Название потока
            <input
              type="text"
              value={feedName}
              onChange={(event) => setFeedName(event.target.value)}
              maxLength={120}
            />
          </label>
        ) : null}
        <div className="builder-actions">
          {!targetFeedId ? (
            <label
              className="catalog-color-button"
              title="Выбрать цвет"
              aria-label="Выбрать цвет потока"
            >
              <input
                type="color"
                aria-label="Выбрать цвет потока"
                value={themeColor}
                onChange={(event) => setThemeColor(event.target.value)}
              />
              <span
                className="catalog-color-swatch"
                style={{ backgroundColor: themeColor }}
                aria-hidden
              />
              <Palette size={17} aria-hidden />
            </label>
          ) : null}
          <button
            className="primary-button catalog-create-button"
            type="button"
            title={
              createFromCatalogMutation.isPending
                ? targetFeedId
                  ? "Добавляем источники"
                  : "Создаем поток"
                : targetFeedId
                  ? "Добавить источники"
                  : "Создать поток"
            }
            aria-label={
              createFromCatalogMutation.isPending
                ? targetFeedId
                  ? "Добавляем источники"
                  : "Создаем поток"
                : targetFeedId
                  ? "Добавить источники"
                  : "Создать поток"
            }
            disabled={!selected.size || createFromCatalogMutation.isPending}
            onClick={() => createFromCatalogMutation.mutate()}
          >
            {createFromCatalogMutation.isPending ? (
              <Loader2 size={18} aria-hidden className="spin" />
            ) : (
              <Plus size={18} aria-hidden />
            )}
          </button>
        </div>
        <span className="builder-selected-count">{selected.size} выбрано</span>
      </div>

      {createFromCatalogMutation.isError ? (
        <ErrorState
          title={targetFeedId ? "Источники не добавлены" : "Поток не создан"}
          message={errorMessage(createFromCatalogMutation.error)}
        />
      ) : null}

      <div className="topic-list">
        {topics.map((topic) => (
          <section className="topic-section" key={topic.id}>
            <div className="topic-heading">
              <h2>{topic.title}</h2>
              <p>{topic.description}</p>
            </div>
            <div className="source-grid">
              {topic.sources.map((source) => {
                const selectionKey = catalogSelectionKey(source.id);
                const checked = selected.has(selectionKey);

                return (
                  <label className={`source-card ${checked ? "selected" : ""}`} key={source.id}>
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={() => toggleSource(selectionKey)}
                    />
                    <span className="source-check" aria-hidden>
                      {checked ? <Check size={16} /> : <Rss size={16} />}
                    </span>
                    <span className="source-card-body">
                      <strong>{source.title}</strong>
                      <span>{source.description}</span>
                      <span className="chip-row">
                        {source.tags.slice(0, 4).map((tag) => (
                          <span className="chip chip-muted" key={tag}>
                            {tag}
                          </span>
                        ))}
                      </span>
                    </span>
                  </label>
                );
              })}
            </div>
          </section>
        ))}

        {customSources.length ? (
          <section className="topic-section" key="custom-sources">
            <div className="topic-heading">
              <h2>Свои источники</h2>
              <p>Пользовательские RSS-источники с локальным хранением материалов.</p>
            </div>
            <div className="source-grid">
              {customSources.map((source) => {
                const selectionKey = customSelectionKey(source.id);
                const checked = selected.has(selectionKey);

                return (
                  <label className={`source-card ${checked ? "selected" : ""}`} key={source.id}>
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={() => toggleSource(selectionKey)}
                    />
                    <span className="source-check" aria-hidden>
                      {checked ? <Check size={16} /> : <Rss size={16} />}
                    </span>
                    <span className="source-card-body">
                      <strong>{source.name}</strong>
                      <span>{source.description || source.feed_url}</span>
                      <span className="chip-row">
                        <span className="chip chip-muted">rss</span>
                        <span className="chip chip-muted">local</span>
                      </span>
                    </span>
                  </label>
                );
              })}
            </div>
          </section>
        ) : null}
      </div>
    </section>
  );
}
