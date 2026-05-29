import type { Category } from "../api/types";

export type DatePreset = "today" | "7d" | "30d" | "all" | "custom";

export type DateFilter = {
  preset: DatePreset;
  label: string;
  mode: "today" | "archive" | "all";
  dateFrom?: string;
  dateTo?: string;
};

export function getSelectedCategorySlugs(searchParams: URLSearchParams) {
  const rawValues = [
    ...searchParams.getAll("categories"),
    ...searchParams.getAll("category")
  ];
  const selected = new Set<string>();

  rawValues.forEach((rawValue) => {
    rawValue.split(",").forEach((part) => {
      const normalized = part.trim();

      if (normalized) {
        selected.add(normalized);
      }
    });
  });

  return Array.from(selected);
}

export function getDateFilter(searchParams: URLSearchParams): DateFilter {
  const preset = (searchParams.get("date") ?? "all") as DatePreset;
  const dateFrom = searchParams.get("date_from") ?? "";
  const dateTo = searchParams.get("date_to") ?? "";

  if (preset === "all") {
    return {
      preset,
      label: "Все даты",
      mode: "all"
    };
  }

  if (preset === "7d") {
    const today = localDateString(new Date());
    const dateFrom = localDateString(addDays(new Date(), -6));

    return {
      preset,
      label: dateRangeLabel(dateFrom, today),
      mode: "all",
      dateFrom,
      dateTo: today
    };
  }

  if (preset === "30d") {
    const today = localDateString(new Date());
    const dateFrom = localDateString(addDays(new Date(), -29));

    return {
      preset,
      label: dateRangeLabel(dateFrom, today),
      mode: "all",
      dateFrom,
      dateTo: today
    };
  }

  if (preset === "custom" && (dateFrom || dateTo)) {
    return {
      preset,
      label: dateRangeLabel(dateFrom, dateTo),
      mode: "all",
      dateFrom: dateFrom || undefined,
      dateTo: dateTo || undefined
    };
  }

  const today = localDateString(new Date());

  return {
    preset: "today",
    label: "Сегодня",
    mode: "today",
    dateFrom: today,
    dateTo: today
  };
}

export function categoryFilterLabel(categories: Category[], selectedSlugs: string[]) {
  if (!selectedSlugs.length) {
    return "Все темы";
  }

  const nameBySlug = new Map(categories.map((category) => [category.slug, category.name]));

  return selectedSlugs.map((slug) => nameBySlug.get(slug) ?? slug).join(", ");
}

export function localDateString(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;
}

function addDays(date: Date, days: number) {
  const next = new Date(date);
  next.setDate(next.getDate() + days);
  return next;
}

function dateRangeLabel(dateFrom: string, dateTo: string) {
  if (dateFrom && dateTo && dateFrom === dateTo) {
    return formatDateOnly(dateFrom);
  }
  if (dateFrom && dateTo) {
    return `${formatDateOnly(dateFrom)} - ${formatDateOnly(dateTo)}`;
  }
  if (dateFrom) {
    return `С ${formatDateOnly(dateFrom)}`;
  }
  return `До ${formatDateOnly(dateTo)}`;
}

function formatDateOnly(value: string) {
  const parsed = new Date(`${value}T00:00:00`);

  if (Number.isNaN(parsed.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "short"
  }).format(parsed);
}
