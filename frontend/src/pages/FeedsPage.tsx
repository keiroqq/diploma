import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowRight, Compass, RefreshCw, Rss, Trash2 } from "lucide-react";
import { useNavigate } from "react-router-dom";

import { deleteFeed, listFeeds, refreshFeed } from "../api/client";
import { EmptyState } from "../components/EmptyState";
import { ErrorState } from "../components/ErrorState";
import { LoadingState } from "../components/LoadingState";
import { errorMessage } from "../utils/errors";

export function FeedsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const feedsQuery = useQuery({
    queryKey: ["feeds"],
    queryFn: listFeeds
  });

  const refreshMutation = useMutation({
    mutationFn: refreshFeed,
    onSuccess: (_, feedId) => {
      queryClient.invalidateQueries({ queryKey: ["feedItems", feedId] });
      queryClient.invalidateQueries({ queryKey: ["saved"] });
    }
  });

  const deleteMutation = useMutation({
    mutationFn: deleteFeed,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feeds"] });
    }
  });

  if (feedsQuery.isLoading) {
    return <LoadingState label="Загружаем ленты" />;
  }

  if (feedsQuery.isError) {
    return <ErrorState message={errorMessage(feedsQuery.error)} />;
  }

  const feeds = feedsQuery.data ?? [];

  if (!feeds.length) {
    return (
      <section className="page-section">
        <EmptyState
          icon={<Rss size={34} aria-hidden />}
          title="Лент пока нет"
          description="Выберите темы в каталоге, и приложение соберет первый поток из RSS-источников."
          action={
            <button className="primary-button" type="button" onClick={() => navigate("/catalog")}>
              <Compass size={18} aria-hidden />
              Перейти в каталог
            </button>
          }
        />
      </section>
    );
  }

  return (
    <section className="page-section">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Потоки</p>
          <h1>Мои ленты</h1>
        </div>
        <button className="secondary-button" type="button" onClick={() => navigate("/catalog")}>
          <Compass size={17} aria-hidden />
          Каталог
        </button>
      </div>

      {refreshMutation.isError ? (
        <ErrorState
          title="Не удалось обновить ленту"
          message={errorMessage(refreshMutation.error)}
        />
      ) : null}
      {deleteMutation.isError ? (
        <ErrorState
          title="Не удалось удалить ленту"
          message={errorMessage(deleteMutation.error)}
        />
      ) : null}

      <div className="feed-grid">
        {feeds.map((feed) => {
          const isRefreshing =
            refreshMutation.isPending && refreshMutation.variables === feed.id;

          return (
            <article className="feed-card" key={feed.id}>
              <div className="feed-card-main">
                <span
                  className="feed-card-icon"
                  style={{ backgroundColor: feed.theme_color || "#2563eb" }}
                >
                  <Rss size={19} aria-hidden />
                </span>
                <div>
                  <h2>{feed.name}</h2>
                  <p>{feed.description || "Персональный поток материалов"}</p>
                </div>
              </div>
              <div className="feed-card-actions">
                <button
                  className="secondary-button compact danger-action"
                  type="button"
                  title="Удалить"
                  aria-label={`Удалить ${feed.name}`}
                  disabled={deleteMutation.isPending}
                  onClick={() => {
                    if (window.confirm(`Удалить поток "${feed.name}"?`)) {
                      deleteMutation.mutate(feed.id);
                    }
                  }}
                >
                  <Trash2 size={18} aria-hidden />
                  Удалить
                </button>
                <button
                  className="icon-button"
                  type="button"
                  title="Обновить"
                  aria-label={`Обновить ${feed.name}`}
                  disabled={isRefreshing}
                  onClick={() => refreshMutation.mutate(feed.id)}
                >
                  <RefreshCw size={18} aria-hidden className={isRefreshing ? "spin" : ""} />
                </button>
                <button
                  className="secondary-button compact"
                  type="button"
                  onClick={() => navigate(`/feeds/${feed.id}`)}
                >
                  Открыть
                  <ArrowRight size={17} aria-hidden />
                </button>
              </div>
            </article>
          );
        })}
      </div>
    </section>
  );
}
