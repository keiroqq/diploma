import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type PointerEvent
} from "react";
import { useQuery } from "@tanstack/react-query";
import { ExternalLink, Loader2, Rss, X } from "lucide-react";

import { getItem } from "../api/client";
import type { Item } from "../api/types";
import { errorMessage } from "../utils/errors";
import { formatPublishedAt } from "../utils/items";

type ArticleReaderModalProps = {
  item: Item | null;
  onClose: () => void;
};

function textParagraphs(value: string) {
  return value
    .split(/\n+/)
    .map((part) => part.trim())
    .filter(Boolean);
}

export function ArticleReaderModal({ item, onClose }: ArticleReaderModalProps) {
  const isLocalItem = item?.storage_mode === "local" || item?.id.startsWith("local:");
  const [collapsed, setCollapsed] = useState(false);
  const dragStartY = useRef<number | null>(null);
  const readerQuery = useQuery({
    queryKey: ["readerItem", item?.id],
    queryFn: () => getItem(item?.id ?? ""),
    enabled: Boolean(item?.id) && !isLocalItem
  });

  const readerItem = readerQuery.data ?? item;
  const localParagraphs = useMemo(
    () => textParagraphs(item?.excerpt ?? ""),
    [item?.excerpt]
  );
  const publishedAt = readerItem ? formatPublishedAt(readerItem.published_at) : "";
  const readerHTML = readerQuery.data?.reader_html ?? "";
  const showFallback = isLocalItem || readerQuery.isError || (!readerQuery.isLoading && !readerHTML);

  useEffect(() => {
    setCollapsed(false);
  }, [item?.id]);

  useEffect(() => {
    if (!item) {
      return;
    }

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        if (collapsed) {
          onClose();
        } else {
          setCollapsed(true);
        }
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [collapsed, item, onClose]);

  function handlePointerDown(event: PointerEvent<HTMLElement>) {
    dragStartY.current = event.clientY;
    event.currentTarget.setPointerCapture(event.pointerId);
  }

  function handlePointerMove(event: PointerEvent<HTMLElement>) {
    if (dragStartY.current === null) {
      return;
    }

    const distance = event.clientY - dragStartY.current;
    if (distance > 72) {
      setCollapsed(true);
    }
    if (distance < -42) {
      setCollapsed(false);
    }
  }

  function handlePointerUp(event: PointerEvent<HTMLElement>) {
    if (dragStartY.current === null) {
      return;
    }

    const distance = event.clientY - dragStartY.current;
    if (distance > 56) {
      setCollapsed(true);
    } else if (distance < -28) {
      setCollapsed(false);
    }
    dragStartY.current = null;
    event.currentTarget.releasePointerCapture(event.pointerId);
  }

  function stopCloseButtonPointer(event: PointerEvent<HTMLButtonElement>) {
    event.stopPropagation();
  }

  if (!item || !readerItem) {
    return null;
  }

  return (
    <div className={`reader-layer ${collapsed ? "collapsed" : ""}`} role="presentation">
      <article
        className={`reader-panel ${collapsed ? "collapsed" : ""}`}
        role="dialog"
        aria-modal="true"
        aria-labelledby="reader-title"
      >
        <div
          className="reader-collapsed-bar"
          role="button"
          tabIndex={0}
          aria-label="Развернуть статью"
          onClick={() => setCollapsed(false)}
          onPointerDown={handlePointerDown}
          onPointerMove={handlePointerMove}
          onPointerUp={handlePointerUp}
          onKeyDown={(event) => {
            if (event.key === "Enter" || event.key === " ") {
              event.preventDefault();
              setCollapsed(false);
            }
          }}
        >
          <div className="reader-collapsed-text">
            <h2>{readerItem.title}</h2>
            {readerItem.author ? <p>{readerItem.author}</p> : null}
          </div>
          <button
            className="icon-button reader-close"
            type="button"
            aria-label="Закрыть читалку"
            onPointerDown={stopCloseButtonPointer}
            onPointerMove={stopCloseButtonPointer}
            onPointerUp={stopCloseButtonPointer}
            onClick={(event) => {
              event.stopPropagation();
              onClose();
            }}
          >
            <X size={20} aria-hidden />
          </button>
        </div>

        <div
          className="reader-sheet-handle"
          onPointerDown={handlePointerDown}
          onPointerMove={handlePointerMove}
          onPointerUp={handlePointerUp}
        >
          <span aria-hidden />
          <button
            className="icon-button reader-close"
            type="button"
            aria-label="Закрыть читалку"
            onPointerDown={stopCloseButtonPointer}
            onPointerMove={stopCloseButtonPointer}
            onPointerUp={stopCloseButtonPointer}
            onClick={(event) => {
              event.stopPropagation();
              onClose();
            }}
          >
            <X size={20} aria-hidden />
          </button>
        </div>

        <div className="reader-scroll">
          <header className="reader-header">
            <div className="reader-heading">
              <h2 id="reader-title">{readerItem.title}</h2>
              {readerItem.author ? <p>{readerItem.author}</p> : null}
            </div>
            <div className="reader-source-line">
              <Rss size={14} aria-hidden />
              <span>{readerItem.source_name || "Источник"}</span>
              {publishedAt ? <span>{publishedAt}</span> : null}
            </div>
          </header>

          {readerItem.image_url ? (
            <figure className="reader-image-frame">
              <img className="reader-image" src={readerItem.image_url} alt="" />
            </figure>
          ) : null}

          <div className="reader-content">
            {readerQuery.isLoading ? (
              <div className="reader-loading">
                <Loader2 size={22} aria-hidden className="spin" />
                <span>Загружаем текст</span>
              </div>
            ) : null}

            {readerQuery.isError ? (
              <p className="reader-note">{errorMessage(readerQuery.error)}</p>
            ) : null}

            {!isLocalItem && readerQuery.data && !readerQuery.data.has_full_content ? (
              <p className="reader-note">Полный текст недоступен, открыт сохраненный анонс.</p>
            ) : null}

            {showFallback ? (
              <div className="reader-body">
                {localParagraphs.length ? (
                  localParagraphs.map((paragraph, index) => (
                    <p key={`${item.id}-${index}`}>{paragraph}</p>
                  ))
                ) : (
                  <p>Текст материала пока недоступен.</p>
                )}
              </div>
            ) : null}

            {readerHTML ? (
              <div
                className="reader-body"
                dangerouslySetInnerHTML={{ __html: readerHTML }}
              />
            ) : null}
          </div>
        </div>

        <a
          className="reader-source-fab"
          href={readerItem.url}
          target="_blank"
          rel="noreferrer"
          aria-label="Открыть источник"
          title="Открыть источник"
        >
          <ExternalLink size={26} aria-hidden />
        </a>
      </article>
    </div>
  );
}
