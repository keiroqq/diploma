import {
  useEffect,
  useId,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type MouseEvent
} from "react";
import { Bookmark, BookmarkCheck, ImageOff, Rss } from "lucide-react";

import type { Item } from "../api/types";
import { useUiStore } from "../store/ui";
import { formatPublishedAt } from "../utils/items";

type ArticleCardProps = {
  item: Item;
  isSaving?: boolean;
  onToggleSaved: (item: Item) => void;
};

export function ArticleCard({
  item,
  isSaving = false,
  onToggleSaved
}: ArticleCardProps) {
  const cardRef = useRef<HTMLElement>(null);
  const titleRef = useRef<HTMLAnchorElement>(null);
  const sourceButtonRef = useRef<HTMLButtonElement>(null);
  const sourceTooltipId = useId();
  const openReader = useUiStore((state) => state.openReader);
  const [titleLines, setTitleLines] = useState(1);
  const [copyLineBudget, setCopyLineBudget] = useState(6);
  const [detailsOpen, setDetailsOpen] = useState(false);
  const tags = useMemo(
    () => Array.from(new Set([...(item.categories ?? []), ...(item.tags ?? [])].filter(Boolean))),
    [item.categories, item.tags]
  );
  const publishedAt = formatPublishedAt(item.published_at);
  const excerptLines = item.excerpt ? Math.max(1, copyLineBudget - titleLines) : 0;
  const cardStyle = {
    "--article-excerpt-lines": excerptLines
  } as CSSProperties;

  function handleOpenReader(event: MouseEvent<HTMLAnchorElement>) {
    event.preventDefault();
    setDetailsOpen(false);
    openReader(item);
  }

  useLayoutEffect(() => {
    const card = cardRef.current;
    if (!card) {
      return;
    }

    const measure = () => {
      const styles = window.getComputedStyle(card);
      const parsedBudget = Number.parseInt(
        styles.getPropertyValue("--article-copy-line-budget"),
        10
      );
      const nextBudget = Number.isFinite(parsedBudget) ? parsedBudget : 6;
      setCopyLineBudget((current) => (current === nextBudget ? current : nextBudget));
    };

    measure();
    const resizeObserver = new ResizeObserver(measure);
    resizeObserver.observe(card);

    return () => resizeObserver.disconnect();
  }, []);

  useLayoutEffect(() => {
    const title = titleRef.current;
    if (!title) {
      return;
    }

    const measure = () => {
      const styles = window.getComputedStyle(title);
      const lineHeight = Number.parseFloat(styles.lineHeight);
      if (!lineHeight) {
        return;
      }

      const parsedClamp = Number.parseInt(
        styles.getPropertyValue("-webkit-line-clamp"),
        10
      );
      const maxTitleLines = Number.isFinite(parsedClamp) ? parsedClamp : 3;
      const nextLines = Math.min(
        maxTitleLines,
        Math.max(1, Math.round(title.getBoundingClientRect().height / lineHeight))
      );
      setTitleLines((current) => (current === nextLines ? current : nextLines));
    };

    measure();
    const resizeObserver = new ResizeObserver(measure);
    resizeObserver.observe(title);

    return () => resizeObserver.disconnect();
  }, [item.title]);

  useEffect(() => {
    if (!detailsOpen) {
      return;
    }

    const closeDetails = (event: Event) => {
      if (event.type === "pointerdown") {
        const target = event.target;
        if (target instanceof Node && sourceButtonRef.current?.contains(target)) {
          return;
        }
      }

      setDetailsOpen(false);
    };

    window.addEventListener("pointerdown", closeDetails, true);
    window.addEventListener("scroll", closeDetails, true);
    window.addEventListener("touchmove", closeDetails, true);
    window.addEventListener("wheel", closeDetails, true);
    window.addEventListener("keydown", closeDetails, true);

    return () => {
      window.removeEventListener("pointerdown", closeDetails, true);
      window.removeEventListener("scroll", closeDetails, true);
      window.removeEventListener("touchmove", closeDetails, true);
      window.removeEventListener("wheel", closeDetails, true);
      window.removeEventListener("keydown", closeDetails, true);
    };
  }, [detailsOpen]);

  return (
    <article ref={cardRef} className="article-card" style={cardStyle}>
      <a
        className="article-media"
        href={item.url}
        aria-label={item.title}
        onClick={handleOpenReader}
      >
        {item.image_url ? (
          <img src={item.image_url} alt="" loading="lazy" />
        ) : (
          <span className="article-placeholder">
            <ImageOff size={24} aria-hidden />
          </span>
        )}
      </a>

      <div className="article-content">
        <div className="article-copy">
          <a
            ref={titleRef}
            className="article-title"
            href={item.url}
            onClick={handleOpenReader}
          >
            {item.title}
          </a>
          {item.excerpt ? <p className="article-excerpt">{item.excerpt}</p> : null}
        </div>

        <div className="article-footer">
          <div className="article-meta">
            <div className="article-source">
              <button
                ref={sourceButtonRef}
                className="article-source-button"
                type="button"
                aria-expanded={detailsOpen}
                aria-describedby={detailsOpen ? sourceTooltipId : undefined}
                onClick={() => setDetailsOpen((open) => !open)}
              >
                <Rss size={13} aria-hidden />
                <span>{item.source_name || "Источник"}</span>
              </button>
              <div
                className={`article-source-popover ${detailsOpen ? "visible" : ""}`}
                id={sourceTooltipId}
                role="tooltip"
                aria-hidden={!detailsOpen}
              >
                <span className="article-source-author">
                  Автор: {item.author || "не указан"}
                </span>
                <div className="article-source-tags" aria-label="Тэги материала">
                  <span className="article-source-tags-label">Тэги:</span>
                  {tags.length ? (
                    tags.map((tag) => (
                      <span className="chip chip-muted" key={tag}>
                        {tag}
                      </span>
                    ))
                  ) : (
                    <span className="article-source-empty">нет</span>
                  )}
                </div>
              </div>
            </div>
            <div className="article-meta-actions">
              {publishedAt ? <span className="article-published-at">{publishedAt}</span> : null}
              <button
                className="icon-button save-button"
                type="button"
                title={item.is_saved ? "Убрать из избранного" : "Сохранить"}
                aria-label={item.is_saved ? "Убрать из избранного" : "Сохранить"}
                aria-pressed={item.is_saved}
                disabled={isSaving}
                onClick={() => onToggleSaved(item)}
              >
                {item.is_saved ? (
                  <BookmarkCheck size={19} aria-hidden />
                ) : (
                  <Bookmark size={19} aria-hidden />
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </article>
  );
}
