const apiErrorMessages: Record<string, string> = {
  "validation failed": "Проверьте заполненные поля: название должно быть не короче 2 символов, а RSS-ссылка должна быть корректным URL.",
  "invalid source url": "Проверьте RSS-ссылку: нужен корректный публичный http/https адрес.",
  "unsafe source url": "Этот адрес нельзя подключить: localhost и приватные сетевые адреса заблокированы для безопасности.",
  "rss feed could not be fetched": "Не удалось получить RSS-ленту. Проверьте, что ссылка ведет на доступный RSS/Atom-фид.",
  "source is disabled": "Источник выключен. Включите его и попробуйте обновить еще раз.",
  "only rss sources are supported": "Сейчас поддерживаются только RSS/Atom-источники.",
  "source preview is not configured": "Предпросмотр RSS временно недоступен на сервере.",
  "internal server error": "Сервер не смог обработать RSS-ленту. Проверьте ссылку или попробуйте позже."
};

export function errorMessage(error: unknown, fallback = "Проверьте API и попробуйте еще раз") {
  if (error instanceof Error && error.message) {
    return apiErrorMessages[error.message] ?? error.message;
  }

  return fallback;
}
