import type { CatalogSource } from "../api/types";

const SOURCE_PREFERENCES_KEY = "content-digest-source-preferences";

export type ProviderPreference = {
  alias?: string;
  enabled?: boolean;
};

export type CatalogSourcePreference = {
  alias?: string;
  enabled?: boolean;
};

export type SourcePreferences = {
  providers: Record<string, ProviderPreference>;
  catalogSources: Record<string, CatalogSourcePreference>;
};

const emptyPreferences: SourcePreferences = {
  providers: {},
  catalogSources: {}
};

export function loadSourcePreferences(): SourcePreferences {
  if (typeof window === "undefined") {
    return emptyPreferences;
  }

  try {
    const raw = window.localStorage.getItem(SOURCE_PREFERENCES_KEY);
    if (!raw) {
      return emptyPreferences;
    }

    const parsed = JSON.parse(raw) as Partial<SourcePreferences>;
    return {
      providers: parsed.providers ?? {},
      catalogSources: parsed.catalogSources ?? {}
    };
  } catch {
    return emptyPreferences;
  }
}

export function saveSourcePreferences(preferences: SourcePreferences) {
  window.localStorage.setItem(SOURCE_PREFERENCES_KEY, JSON.stringify(preferences));
}

export function providerLabel(provider: string) {
  const labels: Record<string, string> = {
    habr: "Хабр",
    sports: "Sports.ru"
  };

  return labels[provider] ?? provider;
}

export function providerTitle(provider: string, preferences: SourcePreferences) {
  const alias = preferences.providers[provider]?.alias?.trim();
  return alias || providerLabel(provider);
}

export function catalogSourceTitle(source: CatalogSource, preferences: SourcePreferences) {
  const alias = preferences.catalogSources[source.id]?.alias?.trim();
  return alias || source.title;
}

export function providerEnabled(provider: string, preferences: SourcePreferences) {
  return preferences.providers[provider]?.enabled !== false;
}

export function catalogSourceEnabled(source: CatalogSource, preferences: SourcePreferences) {
  return providerEnabled(source.provider, preferences)
    && preferences.catalogSources[source.id]?.enabled !== false;
}

export function catalogFeedUrl(source: CatalogSource) {
  if (source.feed_url) {
    return source.feed_url;
  }

  if (source.provider !== "habr") {
    return source.page_url;
  }

  try {
    const url = new URL(source.page_url);
    const parts = url.pathname.split("/").filter(Boolean);
    if (parts[0] === "ru" && parts[1] !== "rss") {
      url.pathname = `/ru/rss/${parts.slice(1).join("/")}/`;
    }
    url.searchParams.set("fl", "ru");
    return url.toString();
  } catch {
    return source.page_url;
  }
}
