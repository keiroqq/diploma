import type { Item } from "../api/types";

export function filterItemsByQuery(items: Item[], query: string) {
  const normalized = query.trim().toLowerCase();

  if (!normalized) {
    return items;
  }

  return items.filter((item) => {
    const searchable = [
      item.title,
      item.excerpt,
      item.source_name,
      item.author,
      ...item.tags,
      ...item.categories
    ]
      .filter(Boolean)
      .join(" ")
      .toLowerCase();

    return searchable.includes(normalized);
  });
}

export function formatPublishedAt(value: string) {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "";
  }

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
}
