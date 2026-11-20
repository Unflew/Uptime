"use client";

import { useEffect, useState } from "react";

type Profile = { login: string; created_at: string };

export default function Dashboard({ onLogout, loading, error }: { onLogout: () => Promise<void>; loading: boolean; error: string }) {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [profileError, setProfileError] = useState("");
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let active = true;
    async function load() {
      try {
        let response = await fetch("/api/auth/profile", { cache: "no-store" });
        if (response.status === 401) {
          const refresh = await fetch("/api/auth/refresh", { method: "POST" });
          if (refresh.ok) response = await fetch("/api/auth/profile", { cache: "no-store" });
          else if (refresh.status === 401) {
            if (active) await onLogout();
            return;
          } else throw new Error();
        }
        if (!response.ok) throw new Error();
        const data = await response.json();
        if (active) { setProfile(data); setProfileError(""); }
      } catch {
        if (active) setProfileError("Не удалось загрузить профиль. Попробуйте ещё раз.");
      }
    }
    load();
    return () => { active = false; };
  // Reload only when explicitly requested; the parent callback changes on render.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [attempt]);

  return (
    <main className="dashboard">
      <aside className="dashboard-sidebar">
        <div className="dashboard-brand">⌁ <span>uptime<span>up</span>.</span></div>
        <div className="dashboard-nav-label">РАБОЧЕЕ ПРОСТРАНСТВО</div>
        <div className="dashboard-nav" aria-current="page">◫ <span>Профиль</span></div>
        <div className="dashboard-sidebar-bottom">UPTIME UP © {new Date().getFullYear()}</div>
      </aside>
      <section className="dashboard-main">
        <header className="dashboard-header"><span>Профиль</span><button type="button" onClick={onLogout} disabled={loading}>{loading ? "Выходим…" : "Выйти"}</button></header>
        <div className="dashboard-content">
          <div className="dashboard-eyebrow">ВАШ АККАУНТ</div>
          <h1>Профиль<span>.</span></h1>
          <p className="dashboard-intro">Информация о вашем аккаунте.</p>
          <div className="profile-card">
            <div className="profile-card-heading"><div><h2>Данные профиля</h2><p>Ваши данные для входа в Uptime Up</p></div><div className="profile-avatar" aria-hidden="true">{profile?.login.slice(0, 1).toUpperCase() ?? "•"}</div></div>
            {profile ? <div className="profile-detail"><span>Логин</span><strong>{profile.login}</strong></div> : profileError ? <div role="alert" className="profile-error">{profileError} <button type="button" onClick={() => { setProfileError(""); setAttempt((n) => n + 1); }}>Повторить</button></div> : <p role="status">Загружаем профиль…</p>}
          </div>
          {error && <p className="profile-error" role="alert">{error}</p>}
        </div>
      </section>
    </main>
  );
}
