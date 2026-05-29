import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Check,
  ChevronDown,
  ChevronRight,
  Folder,
  FolderOpen,
  Loader2,
  Palette,
  Plus,
  Rss
} from "lucide-react";
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
import type { CatalogSource, Source } from "../api/types";
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

type CatalogFolder = {
  id: string;
  title: string;
  description: string;
  sources: CatalogSource[];
};

function sourceCard(
  source: CatalogSource,
  checked: boolean,
  onToggle: () => void
) {
  return (
    <label className={`source-card ${checked ? "selected" : ""}`} key={source.id}>
      <input
        type="checkbox"
        checked={checked}
        onChange={onToggle}
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
}

function customSourceCard(
  source: Source,
  checked: boolean,
  onToggle: () => void
) {
  return (
    <label className={`source-card ${checked ? "selected" : ""}`} key={source.id}>
      <input
        type="checkbox"
        checked={checked}
        onChange={onToggle}
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
}

export function CatalogPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const queryClient = useQueryClient();
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [feedName, setFeedName] = useState("Моя IT-лента");
  const [themeColor, setThemeColor] = useState("#2563eb");
  const [sourcePreferences] = useState(loadSourcePreferences);
  const [openedFolders, setOpenedFolders] = useState<Set<string>>(new Set());
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

  const catalogFolders = useMemo<CatalogFolder[]>(() => {
    return topics.map((topic) => ({
      id: topic.id,
      title: topic.title,
      description: topic.description,
      sources: topic.sources
    }));
  }, [topics]);

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

  function toggleFolder(folderId: string) {
    setOpenedFolders((previous) => {
      const next = new Set(previous);

      if (next.has(folderId)) {
        next.delete(folderId);
      } else {
        next.add(folderId);
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

      <div className="catalog-folder-list">
        {catalogFolders.map((folder) => {
          const opened = openedFolders.has(folder.id);
          const sourceCount = folder.sources.length;

          return (
            <section className="catalog-folder" key={folder.id}>
              <button
                className="catalog-folder-button"
                type="button"
                onClick={() => toggleFolder(folder.id)}
                aria-expanded={opened}
              >
                <span className="catalog-folder-title">
                  {opened ? <ChevronDown size={18} aria-hidden /> : <ChevronRight size={18} aria-hidden />}
                  {opened ? <FolderOpen size={20} aria-hidden /> : <Folder size={20} aria-hidden />}
                  <span>
                    <strong>{folder.title}</strong>
                    <small>{folder.description}</small>
                  </span>
                </span>
                <span className="catalog-folder-count">{sourceCount} источников</span>
              </button>

              {opened ? (
                <div className="catalog-folder-body">
                  <section className="topic-section">
                    <div className="source-grid">
                      {folder.sources.map((source) => {
                        const selectionKey = catalogSelectionKey(source.id);
                        return sourceCard(
                          source,
                          selected.has(selectionKey),
                          () => toggleSource(selectionKey)
                        );
                      })}
                    </div>
                  </section>
                </div>
              ) : null}
            </section>
          );
        })}

        <section className="catalog-folder" key="custom-sources">
          <button
            className="catalog-folder-button"
            type="button"
            onClick={() => toggleFolder("custom-sources")}
            aria-expanded={openedFolders.has("custom-sources")}
          >
            <span className="catalog-folder-title">
              {openedFolders.has("custom-sources") ? <ChevronDown size={18} aria-hidden /> : <ChevronRight size={18} aria-hidden />}
              {openedFolders.has("custom-sources") ? <FolderOpen size={20} aria-hidden /> : <Folder size={20} aria-hidden />}
              <span>
                <strong>Свои источники</strong>
                <small>Пользовательские RSS-источники с локальным хранением материалов.</small>
              </span>
            </span>
            <span className="catalog-folder-count">{customSources.length} источников</span>
          </button>

          {openedFolders.has("custom-sources") ? (
            <div className="catalog-folder-body">
              {customSources.length ? (
                <section className="topic-section">
                  <div className="source-grid">
                    {customSources.map((source) => {
                      const selectionKey = customSelectionKey(source.id);
                      return customSourceCard(
                        source,
                        selected.has(selectionKey),
                        () => toggleSource(selectionKey)
                      );
                    })}
                  </div>
                </section>
              ) : (
                <p className="catalog-folder-empty">
                  В этой папке пока нет источников. Добавить RSS можно на вкладке "Источники".
                </p>
              )}
            </div>
          ) : null}
        </section>
      </div>
    </section>
  );
}
