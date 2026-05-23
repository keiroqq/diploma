import { FormEvent, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Link, useNavigate } from "react-router-dom";
import { Rss, UserPlus } from "lucide-react";

import { register } from "../api/client";
import { useAuthStore } from "../store/auth";
import { errorMessage } from "../utils/errors";

export function RegisterPage() {
  const navigate = useNavigate();
  const setAuth = useAuthStore((state) => state.setAuth);
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const mutation = useMutation({
    mutationFn: register,
    onSuccess: (data) => {
      setAuth(data.token, data.user);
      navigate("/feeds", { replace: true });
    }
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    mutation.mutate({ username, email, password });
  }

  return (
    <main className="auth-page">
      <section className="auth-card" aria-labelledby="register-title">
        <div className="auth-logo">
          <Rss size={24} aria-hidden />
        </div>
        <h1 id="register-title">Регистрация</h1>
        <p>Создайте аккаунт, чтобы собрать персональные RSS-потоки.</p>

        <form className="auth-form" onSubmit={handleSubmit}>
          <label>
            Имя
            <input
              type="text"
              autoComplete="name"
              minLength={2}
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              required
            />
          </label>
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
              autoComplete="new-password"
              minLength={8}
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
            />
          </label>

          {mutation.isError ? (
            <div className="form-error" role="alert">
              {errorMessage(mutation.error, "Не удалось создать аккаунт")}
            </div>
          ) : null}

          <button className="primary-button full-width" type="submit" disabled={mutation.isPending}>
            <UserPlus size={18} aria-hidden />
            {mutation.isPending ? "Создаем" : "Создать аккаунт"}
          </button>
        </form>

        <div className="auth-switch">
          Уже есть аккаунт? <Link to="/login">Войти</Link>
        </div>
      </section>
    </main>
  );
}
