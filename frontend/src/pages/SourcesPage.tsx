import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ChevronDown,
  ChevronRight,
  ExternalLink,
  Folder,
  FolderOpen,
  Loader2,
  Plus,
  Power,
  PowerOff,
  Rss
} from "lucide-react";

import {
  createSource,
  listCatalogTopics,
  listSources,
  updateSource
} from "../api/client";
import type { CatalogSource, Source, UpdateSourceRequest } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { ErrorState } from "../components/ErrorState";
import { LoadingState } from "../components/LoadingState";
import { errorMessage } from "../utils/errors";
import {
  catalogFeedUrl,
  catalogSourceEnabled,
  loadSourcePreferences,
  providerEnabled,
  providerLabel,
  providerTitle,
  saveSourcePreferences,
  type SourcePreferences
} from "../utils/sourcePreferences";

type ProviderGroup = {
  provider: string;
  sources: CatalogSource[];
};

type UpdateCustomSourceVariables = {
  source: Source;
  patch: Partial<Pick<Source, "name" | "status">>;
};

function sourceUpdatePayload(source: Source, patch: UpdateCustomSourceVariables["patch"]): UpdateSourceRequest {
  return {
    name: patch.name ?? source.name,
    type: "rss",
    url: source.url || source.feed_url,
    feed_url: source.feed_url || source.url,
    description: source.description || "",
    language: source.language || "ru",
    is_public: source.is_public,
    storage_mode: source.storage_mode,
    status: patch.status ?? source.status
  };
}

export function SourcesPage() {
  const queryClient = useQueryClient();
  const [preferences, setPreferences] = useState(loadSourcePreferences);
  const [openedProviders, setOpenedProviders] = useState<Set<string>>(new Set(["habr"]));
  const [customOpen, setCustomOpen] = useState(true);
  const [customDrafts, setCustomDrafts] = useState<Record<string, string>>({});
  const [newSourceName, setNewSourceName] = useState("");
  const [newSourceUrl, setNewSourceUrl] = useState("");

  const topicsQuery = useQuery({
    queryKey: ["catalogTopics"],
    queryFn: listCatalogTopics
  });

  const sourcesQuery = useQuery({
    queryKey: ["sources"],
    queryFn: listSources
  });

  const providerGroups = useMemo<ProviderGroup[]>(() => {
    const byProvider = new Map<string, CatalogSource[]>();

    for (const topic of topicsQuery.data ?? []) {
      for (const source of topic.sources) {
        const sources = byProvider.get(source.provider) ?? [];
        sources.push(source);
        byProvider.set(source.provider, sources);
      }
    }

    return [...byProvider.entries()].map(([provider, sources]) => ({ provider, sources }));
  }, [topicsQuery.data]);

  const customSources = useMemo(
    () => (sourcesQuery.data ?? [])
      .filter((source) => source.storage_mode === "local" && !source.is_public)
      .sort((left, right) => left.name.localeCompare(right.name, "ru")),
    [sourcesQuery.data]
  );

  const createCustomSourceMutation = useMutation({
    mutationFn: () => {
      const name = newSourceName.trim();
      const feedURL = newSourceUrl.trim();

      return createSource({
        name,
        type: "rss",
        url: feedURL,
        feed_url: feedURL,
        description: "Пользовательский RSS-источник",
        language: "ru",
        is_public: false,
        storage_mode: "local"
      });
    },
    onSuccess: async () => {
      setNewSourceName("");
      setNewSourceUrl("");
      await queryClient.invalidateQueries({ queryKey: ["sources"] });
    }
  });

  const updateCustomSourceMutation = useMutation({
    mutationFn: ({ source, patch }: UpdateCustomSourceVariables) =>
      updateSource(source.id, sourceUpdatePayload(source, patch)),
    onSuccess: async (_updated, variables) => {
      setCustomDrafts((previous) => {
        const next = { ...previous };
        delete next[variables.source.id];
        return next;
      });
      await queryClient.invalidateQueries({ queryKey: ["sources"] });
    }
  });

  function persistPreferences(next: SourcePreferences) {
    setPreferences(next);
    saveSourcePreferences(next);
  }

  function updateProviderPreference(provider: string, patch: SourcePreferences["providers"][string]) {
    persistPreferences({
      ...preferences,
      providers: {
        ...preferences.providers,
        [provider]: {
          ...preferences.providers[provider],
          ...patch
        }
      }
    });
  }

  function updateCatalogSourcePreference(
    sourceID: string,
    patch: SourcePreferences["catalogSources"][string]
  ) {
    persistPreferences({
      ...preferences,
      catalogSources: {
        ...preferences.catalogSources,
        [sourceID]: {
          ...preferences.catalogSources[sourceID],
          ...patch
        }
      }
    });
  }

  function toggleProvider(provider: string) {
    setOpenedProviders((previous) => {
      const next = new Set(previous);
      if (next.has(provider)) {
        next.delete(provider);
      } else {
        next.add(provider);
      }
      return next;
    });
  }

  function saveCustomSourceName(source: Source, name: string) {
    const nextName = name.trim();
    if (nextName.length < 2 || nextName === source.name || updateCustomSourceMutation.isPending) {
      return;
    }

    updateCustomSourceMutation.mutate({
      source,
      patch: { name: nextName }
    });
  }

  if (topicsQuery.isLoading || sourcesQuery.isLoading) {
    return <LoadingState label="Загружаем источники" />;
  }

  if (topicsQuery.isError) {
    return <ErrorState message={errorMessage(topicsQuery.error)} />;
  }

  if (sourcesQuery.isError) {
    return <ErrorState message={errorMessage(sourcesQuery.error)} />;
  }

  return (
    <section className="page-section sources-page">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Источники</p>
          <h1>RSS-источники</h1>
        </div>
      </div>

      <div className="source-folder-list">
        {providerGroups.map((group) => {
          const opened = openedProviders.has(group.provider);
          const enabled = providerEnabled(group.provider, preferences);

          return (
            <article className={`source-folder ${enabled ? "" : "disabled"}`} key={group.provider}>
              <div className="source-folder-header">
                <button
                  className="source-folder-title"
                  type="button"
                  onClick={() => toggleProvider(group.provider)}
                  aria-expanded={opened}
                >
                  {opened ? <ChevronDown size={18} aria-hidden /> : <ChevronRight size={18} aria-hidden />}
                  {opened ? <FolderOpen size={20} aria-hidden /> : <Folder size={20} aria-hidden />}
                  <span>{providerTitle(group.provider, preferences)}</span>
                </button>
                <label className="source-name-field">
                  Локальное название
                  <input
                    type="text"
                    value={preferences.providers[group.provider]?.alias ?? ""}
                    placeholder={providerLabel(group.provider)}
                    onChange={(event) =>
                      updateProviderPreference(group.provider, { alias: event.target.value })
                    }
                  />
                </label>
                <button
                  className={`icon-button ${enabled ? "" : "muted"}`}
                  type="button"
                  title={enabled ? "Выключить источник" : "Включить источник"}
                  aria-label={enabled ? "Выключить источник" : "Включить источник"}
                  onClick={() => updateProviderPreference(group.provider, { enabled: !enabled })}
                >
                  {enabled ? <Power size={17} aria-hidden /> : <PowerOff size={17} aria-hidden />}
                </button>
              </div>

              {opened ? (
                <div className="source-folder-body">
                  {group.sources.map((source) => {
                    const sourceEnabled = catalogSourceEnabled(source, preferences);
                    const directEnabled = preferences.catalogSources[source.id]?.enabled !== false;

                    return (
                      <div
                        className={`managed-source-row ${sourceEnabled ? "" : "disabled"}`}
                        key={source.id}
                      >
                        <label className="source-name-field managed-source-name">
                          Название
                          <input
                            type="text"
                            value={preferences.catalogSources[source.id]?.alias ?? source.title}
                            placeholder={source.title}
                            onChange={(event) =>
                              updateCatalogSourcePreference(source.id, { alias: event.target.value })
                            }
                          />
                        </label>
                        <a
                          className="managed-source-link"
                          href={catalogFeedUrl(source)}
                          target="_blank"
                          rel="noreferrer"
                        >
                          <Rss size={16} aria-hidden />
                          <span>{catalogFeedUrl(source)}</span>
                          <ExternalLink size={13} aria-hidden />
                        </a>
                        <button
                          className={`icon-button ${sourceEnabled ? "" : "muted"}`}
                          type="button"
                          title={directEnabled ? "Скрыть из каталога" : "Показать в каталоге"}
                          aria-label={directEnabled ? "Скрыть из каталога" : "Показать в каталоге"}
                          onClick={() =>
                            updateCatalogSourcePreference(source.id, { enabled: !directEnabled })
                          }
                        >
                          {directEnabled ? <Power size={17} aria-hidden /> : <PowerOff size={17} aria-hidden />}
                        </button>
                      </div>
                    );
                  })}
                </div>
              ) : null}
            </article>
          );
        })}

        <article className="source-folder">
          <div className="source-folder-header">
            <button
              className="source-folder-title"
              type="button"
              onClick={() => setCustomOpen((value) => !value)}
              aria-expanded={customOpen}
            >
              {customOpen ? <ChevronDown size={18} aria-hidden /> : <ChevronRight size={18} aria-hidden />}
              {customOpen ? <FolderOpen size={20} aria-hidden /> : <Folder size={20} aria-hidden />}
              <span>Свои источники</span>
            </button>
          </div>

          {customOpen ? (
            <div className="source-folder-body">
              <form
                className="custom-source-builder"
                onSubmit={(event) => {
                  event.preventDefault();
                  createCustomSourceMutation.mutate();
                }}
              >
                <label>
                  Название
                  <input
                    type="text"
                    value={newSourceName}
                    maxLength={160}
                    placeholder="Мой RSS"
                    onChange={(event) => setNewSourceName(event.target.value)}
                  />
                </label>
                <label>
                  RSS-ссылка
                  <input
                    type="url"
                    value={newSourceUrl}
                    placeholder="https://example.com/feed.xml"
                    onChange={(event) => setNewSourceUrl(event.target.value)}
                  />
                </label>
                <button
                  className="primary-button custom-source-submit"
                  type="submit"
                  disabled={
                    createCustomSourceMutation.isPending
                    || newSourceName.trim().length < 2
                    || !newSourceUrl.trim()
                  }
                >
                  {createCustomSourceMutation.isPending ? (
                    <Loader2 size={18} aria-hidden className="spin" />
                  ) : (
                    <Plus size={18} aria-hidden />
                  )}
                </button>
              </form>

              {createCustomSourceMutation.isError ? (
                <ErrorState
                  title="Источник не добавлен"
                  message={errorMessage(createCustomSourceMutation.error)}
                />
              ) : null}

              {customSources.length ? (
                customSources.map((source) => {
                  const draftName = customDrafts[source.id] ?? source.name;
                  const enabled = source.status !== "disabled";
                  const changed = draftName.trim() !== source.name;
                  const updating = updateCustomSourceMutation.isPending;

                  return (
                    <div
                      className={`managed-source-row custom ${enabled ? "" : "disabled"}`}
                      key={source.id}
                    >
                      <label className="source-name-field managed-source-name">
                        Название
                        <input
                          type="text"
                          value={draftName}
                          maxLength={160}
                          onBlur={() => saveCustomSourceName(source, draftName)}
                          onKeyDown={(event) => {
                            if (event.key === "Enter") {
                              event.preventDefault();
                              saveCustomSourceName(source, draftName);
                              event.currentTarget.blur();
                            }
                          }}
                          onChange={(event) =>
                            setCustomDrafts((previous) => ({
                              ...previous,
                              [source.id]: event.target.value
                            }))
                          }
                        />
                      </label>
                      <a
                        className="managed-source-link"
                        href={source.feed_url}
                        target="_blank"
                        rel="noreferrer"
                      >
                        <Rss size={16} aria-hidden />
                        <span>{source.feed_url}</span>
                        <ExternalLink size={13} aria-hidden />
                      </a>
                      <button
                        className={`icon-button ${enabled ? "" : "muted"}`}
                        type="button"
                        title={enabled ? "Выключить источник" : "Включить источник"}
                        aria-label={enabled ? "Выключить источник" : "Включить источник"}
                        disabled={updating}
                        onClick={() =>
                          updateCustomSourceMutation.mutate({
                            source,
                            patch: {
                              name:
                                changed && draftName.trim().length >= 2
                                  ? draftName.trim()
                                  : source.name,
                              status: enabled ? "disabled" : "active"
                            }
                          })
                        }
                      >
                        {enabled ? <Power size={17} aria-hidden /> : <PowerOff size={17} aria-hidden />}
                      </button>
                    </div>
                  );
                })
              ) : (
                <EmptyState
                  icon={<Rss size={34} aria-hidden />}
                  title="Своих источников пока нет"
                  description="Добавьте RSS-ссылку, и она появится в конце каталога."
                />
              )}

              {updateCustomSourceMutation.isError ? (
                <ErrorState
                  title="Источник не обновлен"
                  message={errorMessage(updateCustomSourceMutation.error)}
                />
              ) : null}
            </div>
          ) : null}
        </article>
      </div>
    </section>
  );
}
