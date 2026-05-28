import type { DateFilter } from "./filters";
import type { Item, PreviewItem, Source } from "../api/types";

const DB_NAME = "content-digest-local-cache";
const DB_VERSION = 1;
const ITEM_STORE = "items";
const MAX_ITEMS_PER_SOURCE = 500;

export type LocalItem = Item & {
  storage_mode: "local";
  feed_id: string;
  category_slugs: string[];
  search_text: string;
  cached_at: string;
};

export async function listLocalFeedItems(feedId: string) {
  const db = await openLocalDB();
  return getAllByIndex<LocalItem>(db, ITEM_STORE, "feed_id", feedId);
}

export async function searchLocalItems(query: string, feedId?: string) {
  const normalizedQuery = normalizeSearch(query);
  if (!normalizedQuery) {
    return [];
  }

  const db = await openLocalDB();
  const items = feedId
    ? await getAllByIndex<LocalItem>(db, ITEM_STORE, "feed_id", feedId)
    : await getAll<LocalItem>(db, ITEM_STORE);

  return items
    .filter((item) => item.search_text.includes(normalizedQuery))
    .sort(byPublishedDesc);
}

export async function cacheLocalSourceItems(feedId: string, source: Source, items: PreviewItem[]) {
  const db = await openLocalDB();
  const existingItems = await getAllByIndex<LocalItem>(db, ITEM_STORE, "source_id", source.id);
  const savedByID = new Map(existingItems.map((item) => [item.id, item.is_saved]));
  const cachedAt = new Date().toISOString();

  await writeTransaction(db, ITEM_STORE, "readwrite", (store) => {
    items.forEach((item) => {
      const localItem: LocalItem = {
        ...item,
        feed_id: feedId,
        source_id: source.id,
        source_name: source.name,
        storage_mode: "local",
        is_saved: savedByID.get(item.id) ?? false,
        category_slugs: item.category_slugs ?? [],
        search_text: normalizeSearch(item.search_text || searchText(item)),
        cached_at: cachedAt
      };

      store.put(localItem);
    });
  });

  await trimSourceItems(source.id);
}

export async function removeLocalItemsBySource(sourceId: string) {
  const db = await openLocalDB();
  const items = await getAllByIndex<LocalItem>(db, ITEM_STORE, "source_id", sourceId);
  await writeTransaction(db, ITEM_STORE, "readwrite", (store) => {
    items.forEach((item) => store.delete(item.id));
  });
}

export async function toggleLocalItemSaved(itemId: string) {
  const db = await openLocalDB();
  const item = await getByKey<LocalItem>(db, ITEM_STORE, itemId);
  if (!item) {
    return;
  }

  await writeTransaction(db, ITEM_STORE, "readwrite", (store) => {
    store.put({ ...item, is_saved: !item.is_saved });
  });
}

export function filterLocalItems(items: LocalItem[], dateFilter: DateFilter, selectedCategories: string[]) {
  const selectedCategorySet = new Set(selectedCategories);

  return items.filter((item) => {
    const publishedAt = new Date(item.published_at);
    if (Number.isNaN(publishedAt.getTime())) {
      return false;
    }

    if (!matchesDateFilter(publishedAt, dateFilter)) {
      return false;
    }

    if (!selectedCategorySet.size) {
      return true;
    }

    return item.category_slugs.some((slug) => selectedCategorySet.has(slug));
  });
}

function matchesDateFilter(publishedAt: Date, dateFilter: DateFilter) {
  if (dateFilter.dateFrom) {
    const from = startOfLocalDate(dateFilter.dateFrom);
    if (publishedAt < from) {
      return false;
    }
  }

  if (dateFilter.dateTo) {
    const to = addDays(startOfLocalDate(dateFilter.dateTo), 1);
    if (publishedAt >= to) {
      return false;
    }
  }

  if (!dateFilter.dateFrom && !dateFilter.dateTo && dateFilter.mode === "archive") {
    return publishedAt < startOfToday();
  }

  return true;
}

function startOfLocalDate(value: string) {
  const [year, month, day] = value.split("-").map(Number);
  return new Date(year, (month || 1) - 1, day || 1);
}

function startOfToday() {
  const now = new Date();
  return new Date(now.getFullYear(), now.getMonth(), now.getDate());
}

function addDays(value: Date, days: number) {
  const next = new Date(value);
  next.setDate(next.getDate() + days);
  return next;
}

async function trimSourceItems(sourceId: string) {
  const db = await openLocalDB();
  const items = await getAllByIndex<LocalItem>(db, ITEM_STORE, "source_id", sourceId);
  const staleItems = items.sort(byPublishedDesc).slice(MAX_ITEMS_PER_SOURCE);
  if (!staleItems.length) {
    return;
  }

  await writeTransaction(db, ITEM_STORE, "readwrite", (store) => {
    staleItems.forEach((item) => store.delete(item.id));
  });
}

function byPublishedDesc(left: Item, right: Item) {
  return new Date(right.published_at).getTime() - new Date(left.published_at).getTime();
}

function searchText(item: Item) {
  return [
    item.title,
    item.excerpt,
    item.author,
    item.source_name,
    ...(item.tags ?? []),
    ...(item.categories ?? [])
  ].join(" ");
}

function normalizeSearch(value: string) {
  return value.trim().toLowerCase();
}

function openLocalDB() {
  return new Promise<IDBDatabase>((resolve, reject) => {
    const request = window.indexedDB.open(DB_NAME, DB_VERSION);

    request.onupgradeneeded = () => {
      const db = request.result;
      if (!db.objectStoreNames.contains(ITEM_STORE)) {
        const store = db.createObjectStore(ITEM_STORE, { keyPath: "id" });
        store.createIndex("feed_id", "feed_id", { unique: false });
        store.createIndex("source_id", "source_id", { unique: false });
      }
    };

    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);
  });
}

function getAll<T>(db: IDBDatabase, storeName: string) {
  return new Promise<T[]>((resolve, reject) => {
    const request = db.transaction(storeName, "readonly").objectStore(storeName).getAll();

    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result as T[]);
  });
}

function getAllByIndex<T>(db: IDBDatabase, storeName: string, indexName: string, value: string) {
  return new Promise<T[]>((resolve, reject) => {
    const request = db
      .transaction(storeName, "readonly")
      .objectStore(storeName)
      .index(indexName)
      .getAll(value);

    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result as T[]);
  });
}

function getByKey<T>(db: IDBDatabase, storeName: string, key: string) {
  return new Promise<T | undefined>((resolve, reject) => {
    const request = db.transaction(storeName, "readonly").objectStore(storeName).get(key);

    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result as T | undefined);
  });
}

function writeTransaction(
  db: IDBDatabase,
  storeName: string,
  mode: IDBTransactionMode,
  writer: (store: IDBObjectStore) => void
) {
  return new Promise<void>((resolve, reject) => {
    const transaction = db.transaction(storeName, mode);
    writer(transaction.objectStore(storeName));
    transaction.onerror = () => reject(transaction.error);
    transaction.oncomplete = () => resolve();
  });
}
