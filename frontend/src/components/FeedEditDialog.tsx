import { useEffect, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check, Loader2, Pencil, X } from "lucide-react";

import { updateFeed } from "../api/client";
import type { Feed } from "../api/types";
import { errorMessage } from "../utils/errors";

type FeedEditDialogProps = {
  feed: Feed | null;
  onClose: () => void;
};

const colorOptions = [
  "#2563eb",
  "#0f766e",
  "#16a34a",
  "#ca8a04",
  "#dc2626",
  "#7c3aed"
];

export function FeedEditDialog({ feed, onClose }: FeedEditDialogProps) {
  const queryClient = useQueryClient();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [themeColor, setThemeColor] = useState("#2563eb");

  const updateMutation = useMutation({
    mutationFn: () => {
      if (!feed) {
        throw new Error("feed is not selected");
      }

      return updateFeed(feed.id, {
        name: name.trim(),
        description: description.trim(),
        icon: feed.icon || "rss",
        theme_color: themeColor || "#2563eb",
        layout_type: feed.layout_type || "cards",
        is_default: feed.is_default
      });
    },
    onSuccess: (updatedFeed) => {
      queryClient.setQueryData<Feed[]>(["feeds"], (current) =>
        current?.map((record) => (record.id === updatedFeed.id ? updatedFeed : record))
      );
      queryClient.setQueryData(["feed", updatedFeed.id], updatedFeed);
      queryClient.invalidateQueries({ queryKey: ["feeds"] });
      queryClient.invalidateQueries({ queryKey: ["feed", updatedFeed.id] });
      onClose();
    }
  });

  useEffect(() => {
    if (!feed) {
      return;
    }

    setName(feed.name);
    setDescription(feed.description);
    setThemeColor(feed.theme_color || "#2563eb");
    updateMutation.reset();
  }, [feed]);

  if (!feed) {
    return null;
  }

  const canSubmit = name.trim().length >= 2 && !updateMutation.isPending;

  return (
    <div
      className="modal-layer"
      role="presentation"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget && !updateMutation.isPending) {
          onClose();
        }
      }}
    >
      <section
        className="modal-panel feed-edit-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="feed-edit-title"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <div className="modal-heading">
          <div>
            <p className="eyebrow">Поток</p>
            <h2 id="feed-edit-title">Редактирование</h2>
          </div>
          <button
            className="icon-button"
            type="button"
            title="Закрыть"
            aria-label="Закрыть"
            disabled={updateMutation.isPending}
            onClick={onClose}
          >
            <X size={20} aria-hidden />
          </button>
        </div>

        <form
          className="edit-feed-form"
          onSubmit={(event) => {
            event.preventDefault();
            if (canSubmit) {
              updateMutation.mutate();
            }
          }}
        >
          <label>
            Название
            <input
              type="text"
              value={name}
              minLength={2}
              maxLength={120}
              required
              onChange={(event) => setName(event.target.value)}
            />
          </label>

          <label>
            Описание
            <textarea
              value={description}
              maxLength={1000}
              rows={4}
              onChange={(event) => setDescription(event.target.value)}
            />
          </label>

          <label>
            Цвет
            <span className="color-field">
              <input
                type="color"
                value={themeColor}
                onChange={(event) => setThemeColor(event.target.value)}
              />
              <span>{themeColor}</span>
            </span>
          </label>

          <div className="color-swatch-list" aria-label="Быстрый выбор цвета">
            {colorOptions.map((color) => {
              const selected = color.toLowerCase() === themeColor.toLowerCase();

              return (
                <button
                  className={`color-swatch ${selected ? "active" : ""}`}
                  key={color}
                  type="button"
                  style={{ backgroundColor: color }}
                  title={color}
                  aria-label={`Цвет ${color}`}
                  onClick={() => setThemeColor(color)}
                >
                  {selected ? <Check size={16} aria-hidden /> : null}
                </button>
              );
            })}
          </div>

          {updateMutation.isError ? (
            <div className="form-error" role="alert">
              {errorMessage(updateMutation.error)}
            </div>
          ) : null}

          <div className="form-actions">
            <button
              className="secondary-button"
              type="button"
              disabled={updateMutation.isPending}
              onClick={onClose}
            >
              Отмена
            </button>
            <button className="primary-button" type="submit" disabled={!canSubmit}>
              {updateMutation.isPending ? (
                <Loader2 size={18} aria-hidden className="spin" />
              ) : (
                <Pencil size={18} aria-hidden />
              )}
              {updateMutation.isPending ? "Сохраняем" : "Сохранить"}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}
