import { Navigate, Route, Routes } from "react-router-dom";

import { AppShell } from "./components/AppShell";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import { CatalogPage } from "./pages/CatalogPage";
import { FeedPage } from "./pages/FeedPage";
import { FeedsPage } from "./pages/FeedsPage";
import { SavedPage } from "./pages/SavedPage";
import { SourcesPage } from "./pages/SourcesPage";
import { useAuthStore } from "./store/auth";

function GuestRoute({ children }: { children: JSX.Element }) {
  const token = useAuthStore((state) => state.token);

  if (token) {
    return <Navigate to="/feeds" replace />;
  }

  return children;
}

export function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/feeds" replace />} />
      <Route
        path="/login"
        element={
          <GuestRoute>
            <LoginPage />
          </GuestRoute>
        }
      />
      <Route
        path="/register"
        element={
          <GuestRoute>
            <RegisterPage />
          </GuestRoute>
        }
      />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppShell />}>
          <Route path="/feeds" element={<FeedsPage />} />
          <Route path="/feeds/:feedId" element={<FeedPage />} />
          <Route path="/catalog" element={<CatalogPage />} />
          <Route path="/saved" element={<SavedPage />} />
          <Route path="/sources" element={<SourcesPage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/feeds" replace />} />
    </Routes>
  );
}
