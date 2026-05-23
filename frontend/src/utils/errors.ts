export function errorMessage(error: unknown, fallback = "Проверьте API и попробуйте еще раз") {
  if (error instanceof Error && error.message) {
    return error.message;
  }

  return fallback;
}
