import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Check, Loader2, Palette, Plus, Rss } from "lucide-react";
import { useNavigate, useSearchParams } from "react-router-dom";

import {
  connectCatalogSources,
  createFeed,
  deleteFeed,
  listCatalogTopics,
  refreshFeed
} from "../api/client";
import { ErrorState } from "../components/ErrorState";
import { LoadingState } from "../components/LoadingState";
import { errorMessage } from "../utils/errors";

export function CatalogPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const queryClient = useQueryClient();
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [feedName, setFeedName] = useState("Моя IT-лента");
  const [themeColor, setThemeColor] = useState("#2563eb");
  const targetFeedId = searchParams.get("feed_id");

  const topicsQuery = useQuery({
    queryKey: ["catalogTopics"],
    queryFn: listCatalogTopics
  });

  const selectedSources = useMemo(() => {
    const topics = topicsQuery.data ?? [];
    const selectedIds = selected;

    return topics
      .flatMap((topic) => topic.sources)
      .filter((source) => selectedIds.has(source.id));
  }, [selected, topicsQuery.data]);

  const createFromCatalogMutation = useMutation({
    mutationFn: async () => {
      const sourceIds = selectedSources.map((source) => source.id);
      const sourceTitles = selectedSources.map((source) => source.title).join(", ");
      if (targetFeedId) {
        await connectCatalogSources(targetFeedId, sourceIds);
        await refreshFeed(targetFeedId).catch(() => undefined);
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
        await connectCatalogSources(feed.id, sourceIds);
      } catch (error) {
        await deleteFeed(feed.id).catch(() => undefined);
        throw error;
      }

      await refreshFeed(feed.id).catch(() => undefined);

      return feed;
    },
    onSuccess: async (feed) => {
      await queryClient.invalidateQueries({ queryKey: ["feeds"] });
      await queryClient.invalidateQueries({ queryKey: ["feedCategories", feed.id] });
      await queryClient.invalidateQueries({ queryKey: ["feedSources", feed.id] });
      navigate(`/feeds/${feed.id}`);
    }
  });

  function toggleSource(sourceId: string) {
    setSelected((previous) => {
      const next = new Set(previous);

      if (next.has(sourceId)) {
        next.delete(sourceId);
      } else {
        next.add(sourceId);
      }

      return next;
    });
  }

  if (topicsQuery.isLoading) {
    return <LoadingState label="Загружаем каталог" />;
  }

  if (topicsQuery.isError) {
    return <ErrorState message={errorMessage(topicsQuery.error)} />;
  }

  const topics = topicsQuery.data ?? [];

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
                const checked = selected.has(source.id);

                return (
                  <label className={`source-card ${checked ? "selected" : ""}`} key={source.id}>
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={() => toggleSource(source.id)}
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
      </div>
    </section>
  );
}
