import { AlertCircle } from "lucide-react";

type ErrorStateProps = {
  title?: string;
  message: string;
};

export function ErrorState({
  title = "Что-то пошло не так",
  message
}: ErrorStateProps) {
  return (
    <div className="notice notice-error" role="alert">
      <AlertCircle size={18} aria-hidden />
      <div>
        <strong>{title}</strong>
        <span>{message}</span>
      </div>
    </div>
  );
}
