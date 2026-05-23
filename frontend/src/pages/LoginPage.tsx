import { FormEvent, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { LogIn, Rss } from "lucide-react";

import { login } from "../api/client";
import { useAuthStore } from "../store/auth";
import { errorMessage } from "../utils/errors";

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const setAuth = useAuthStore((state) => state.setAuth);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const from =
    (location.state as { from?: { pathname?: string } } | null)?.from?.pathname ??
    "/feeds";

  const mutation = useMutation({
    mutationFn: login,
    onSuccess: (data) => {
      setAuth(data.token, data.user);
      navigate(from, { replace: true });
    }
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    mutation.mutate({ email, password });
  }

  return (
    <main className="auth-page">
      <section className="auth-card" aria-labelledby="login-title">
        <div className="auth-logo">
          <Rss size={24} aria-hidden />
        </div>
        <h1 id="login-title">Вход</h1>
        <p>Откройте свои потоки и продолжите чтение.</p>

        <form className="auth-form" onSubmit={handleSubmit}>
          <label>
            Email
            <input
              type="email"
              autoComplete="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              required
            />
          </label>
          <label>
            Пароль
            <input
              type="password"
              autoComplete="current-password"
              minLength={8}
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
            />
          </label>

          {mutation.isError ? (
            <div className="form-error" role="alert">
              {errorMessage(mutation.error, "Не удалось войти")}
            </div>
          ) : null}

          <button className="primary-button full-width" type="submit" disabled={mutation.isPending}>
            <LogIn size={18} aria-hidden />
            {mutation.isPending ? "Входим" : "Войти"}
          </button>
        </form>

        <div className="auth-switch">
          Нет аккаунта? <Link to="/register">Зарегистрироваться</Link>
        </div>
      </section>
    </main>
  );
}
