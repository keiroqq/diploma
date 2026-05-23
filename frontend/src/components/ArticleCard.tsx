import { Bookmark, BookmarkCheck, ExternalLink, ImageOff, Rss } from "lucide-react";

import type { Item } from "../api/types";
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
  const chips = [...item.categories, ...item.tags].filter(Boolean).slice(0, 5);
  const publishedAt = formatPublishedAt(item.published_at);

  return (
    <article className="article-card">
      <a
        className="article-media"
        href={item.url}
        target="_blank"
        rel="noreferrer"
        aria-label={item.title}
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
            className="article-title"
            href={item.url}
            target="_blank"
            rel="noreferrer"
          >
            {item.title}
            <ExternalLink size={14} aria-hidden />
          </a>
          {item.excerpt ? <p className="article-excerpt">{item.excerpt}</p> : null}
        </div>

        <div className="article-footer">
          <div className="article-meta">
            <span>
              <Rss size={13} aria-hidden />
              {item.source_name || "Источник"}
            </span>
            {item.author ? <span>{item.author}</span> : null}
            {publishedAt ? <span>{publishedAt}</span> : null}
          </div>

          <div className="article-bottom-row">
            <div className="chip-row" aria-label="Теги материала">
              {chips.map((chip) => (
                <span className="chip chip-muted" key={chip}>
                  {chip}
                </span>
              ))}
            </div>
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
    </article>
  );
}
