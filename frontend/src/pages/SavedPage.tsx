import { useMemo } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Bookmark } from "lucide-react";

import { listSavedItems, saveItem, unsaveItem } from "../api/client";
import type { Item } from "../api/types";
import { ArticleCard } from "../components/ArticleCard";
import { EmptyState } from "../components/EmptyState";
import { ErrorState } from "../components/ErrorState";
import { LoadingState } from "../components/LoadingState";
import { useUiStore } from "../store/ui";
import { errorMessage } from "../utils/errors";
import { filterItemsByQuery } from "../utils/items";

export function SavedPage() {
  const queryClient = useQueryClient();
  const searchQuery = useUiStore((state) => state.searchQuery);

  const savedQuery = useQuery({
    queryKey: ["saved"],
    queryFn: listSavedItems
  });

  const toggleSavedMutation = useMutation({
    mutationFn: (item: Item) => (item.is_saved ? unsaveItem(item.id) : saveItem(item.id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["saved"] });
      queryClient.invalidateQueries({ queryKey: ["feedItems"] });
    }
  });

  const items = savedQuery.data?.items ?? [];
  const visibleItems = useMemo(
    () => filterItemsByQuery(items, searchQuery),
    [items, searchQuery]
  );

  if (savedQuery.isLoading) {
    return <LoadingState label="Загружаем избранное" />;
  }

  if (savedQuery.isError) {
    return <ErrorState message={errorMessage(savedQuery.error)} />;
  }

  if (!items.length) {
    return (
      <section className="page-section">
        <EmptyState
          icon={<Bookmark size={34} aria-hidden />}
          title="Избранное пусто"
          description="Сохраняйте материалы из потоков, чтобы быстро возвращаться к ним позже."
        />
      </section>
    );
  }

  return (
    <section className="page-section">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Сохраненное</p>
          <h1>Избранное</h1>
        </div>
      </div>

      {!visibleItems.length ? (
        <EmptyState title="Ничего не найдено" description="Попробуйте изменить запрос поиска." />
      ) : (
        <div className="article-list">
          {visibleItems.map((item) => (
            <ArticleCard
              key={item.id}
              item={{ ...item, is_saved: true }}
              isSaving={
                toggleSavedMutation.isPending &&
                toggleSavedMutation.variables?.id === item.id
              }
              onToggleSaved={(nextItem) => toggleSavedMutation.mutate(nextItem)}
            />
          ))}
        </div>
      )}
    </section>
  );
}
