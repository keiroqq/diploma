import { useAuthStore } from "../store/auth";
import type {
  AuthResponse,
  Category,
  ConnectCatalogSourcesResponse,
  CreateFeedRequest,
  Feed,
  FeedItemsResponse,
  Item,
  LoginRequest,
  RefreshResult,
  RegisterRequest,
  SavedItemsResponse,
  Topic,
  UpdateFeedRequest,
  User
} from "./types";

const API_BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  status: number;
  details: unknown;

  constructor(status: number, message: string, details?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.details = details;
  }
}

async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = useAuthStore.getState().token;
  const headers = new Headers(init.headers);

  headers.set("Accept", "application/json");
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const raw = await response.text();
  const data = raw ? parseJSON(raw) : null;

  if (!response.ok) {
    if (response.status === 401) {
      useAuthStore.getState().logout();
    }

    throw new ApiError(
      response.status,
      readErrorMessage(data),
      readErrorDetails(data)
    );
  }

  return data as T;
}

function parseJSON(raw: string): unknown {
  try {
    return JSON.parse(raw);
  } catch {
    return { error: raw };
  }
}

function readErrorMessage(data: unknown) {
  if (data && typeof data === "object" && "error" in data) {
    const error = data.error;

    if (typeof error === "string") {
      return error;
    }
  }

  return "Ошибка запроса к API";
}

function readErrorDetails(data: unknown) {
  if (data && typeof data === "object" && "details" in data) {
    return data.details;
  }

  return undefined;
}

function withJSON(body: unknown): RequestInit {
  return {
    body: JSON.stringify(body),
    headers: {
      "Content-Type": "application/json"
    }
  };
}

export function getApiBaseUrl() {
  return API_BASE_URL;
}

export function login(payload: LoginRequest) {
  return apiRequest<AuthResponse>("/api/auth/login", {
    method: "POST",
    ...withJSON(payload)
  });
}

export function register(payload: RegisterRequest) {
  return apiRequest<AuthResponse>("/api/auth/register", {
    method: "POST",
    ...withJSON(payload)
  });
}

export function getMe() {
  return apiRequest<User>("/api/auth/me");
}

export function listFeeds() {
  return apiRequest<Feed[]>("/api/feeds");
}

export function getFeed(feedId: string) {
  return apiRequest<Feed>(`/api/feeds/${feedId}`);
}

export function createFeed(payload: CreateFeedRequest) {
  return apiRequest<Feed>("/api/feeds", {
    method: "POST",
    ...withJSON(payload)
  });
}

export function updateFeed(feedId: string, payload: UpdateFeedRequest) {
  return apiRequest<Feed>(`/api/feeds/${feedId}`, {
    method: "PUT",
    ...withJSON(payload)
  });
}

export function deleteFeed(feedId: string) {
  return apiRequest<void>(`/api/feeds/${feedId}`, {
    method: "DELETE"
  });
}

export function connectCatalogSources(feedId: string, sourceIds: string[]) {
  return apiRequest<ConnectCatalogSourcesResponse>(
    `/api/feeds/${feedId}/catalog-sources`,
    {
      method: "POST",
      ...withJSON({ source_ids: sourceIds })
    }
  );
}

export function refreshFeed(feedId: string) {
  return apiRequest<RefreshResult>(`/api/feeds/${feedId}/refresh`, {
    method: "POST"
  });
}

export function listCatalogTopics() {
  return apiRequest<Topic[]>("/api/catalog/topics");
}

export function listCategories() {
  return apiRequest<Category[]>("/api/categories");
}

export type ListFeedItemsParams = {
  mode?: "today" | "archive" | "all";
  categories?: string[];
  dateFrom?: string;
  dateTo?: string;
  limit?: number;
};

export function listFeedItems(feedId: string, options: ListFeedItemsParams = {}) {
  const params = new URLSearchParams({
    mode: options.mode ?? "today",
    limit: String(options.limit ?? 50)
  });

  if (options.categories?.length) {
    params.set("categories", options.categories.join(","));
  }
  if (options.dateFrom) {
    params.set("date_from", options.dateFrom);
  }
  if (options.dateTo) {
    params.set("date_to", options.dateTo);
  }

  return apiRequest<FeedItemsResponse>(
    `/api/feeds/${feedId}/items?${params.toString()}`
  );
}

export function saveItem(itemId: string) {
  return apiRequest<void>(`/api/items/${itemId}/save`, {
    method: "POST"
  });
}

export function unsaveItem(itemId: string) {
  return apiRequest<void>(`/api/items/${itemId}/save`, {
    method: "DELETE"
  });
}

export function listSavedItems() {
  return apiRequest<SavedItemsResponse>("/api/saved?limit=100");
}

export type SaveToggleVariables = Pick<Item, "id" | "is_saved">;
