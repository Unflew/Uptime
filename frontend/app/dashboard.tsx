"use client";

import { FormEvent, useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";

type Profile = { login: string; display_name: string; created_at: string };

export default function Dashboard({ onLogout, loading, error }: { onLogout: () => Promise<void>; loading: boolean; error: string }) {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [profileError, setProfileError] = useState("");
  const [attempt, setAttempt] = useState(0);
  const [name, setName] = useState("");
  const [saving, setSaving] = useState(false);
  const [saveMessage, setSaveMessage] = useState("");
  const pathname = usePathname();
  const isProfile = pathname === "/profile";
  const displayName = profile?.display_name || profile?.login;

  async function saveProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = name.trim();
    if (!trimmed || Array.from(trimmed).length > 100) { setSaveMessage("Имя должно содержать от 1 до 100 символов."); return; }
    setSaving(true);
    setSaveMessage("");
    try {
      let response = await fetch("/api/auth/profile", { method: "PATCH", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ display_name: trimmed }) });
      if (response.status === 401) {
        const refresh = await fetch("/api/auth/refresh", { method: "POST" });
        if (refresh.ok) response = await fetch("/api/auth/profile", { method: "PATCH", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ display_name: trimmed }) });
        else if (refresh.status === 401) { await onLogout(); return; }
      }
      if (!response.ok) throw new Error();
      const data = await response.json();
      setProfile(data);
      setName(data.display_name);
      setSaveMessage("Имя сохранено.");
    } catch { setSaveMessage("Не удалось сохранить имя. Попробуйте ещё раз."); }
    finally { setSaving(false); }
  }

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
        if (active) { setProfile(data); setName(data.display_name || data.login); setProfileError(""); }
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
        <Link className="dashboard-nav" href="/">◫ <span>Дашборд</span></Link>
        <div className="dashboard-sidebar-bottom">UPTIME UP © {new Date().getFullYear()}</div>
      </aside>
      <section className="dashboard-main">
        <header className="dashboard-header">
          <span>{isProfile ? "Профиль" : "Дашборд"}</span>
          <details className="account-menu">
            <summary aria-label="Меню профиля"><span className="account-menu-avatar" aria-hidden="true">{displayName?.slice(0, 1).toUpperCase() ?? "•"}</span><span className="account-menu-login">{displayName ?? "Аккаунт"}</span><span aria-hidden="true">⌄</span></summary>
            <div className="account-menu-panel">
              <Link href="/profile" className="account-menu-profile"><span>Профиль</span><strong>{displayName ?? "Загрузка…"}</strong></Link>
              <button type="button" onClick={onLogout} disabled={loading}>{loading ? "Выходим…" : "Выйти"}</button>
            </div>
          </details>
        </header>
        <div className="dashboard-content">
          {isProfile && <>
            <div className="dashboard-eyebrow">ВАШ АККАУНТ</div>
            <h1>Профиль<span>.</span></h1>
            <p className="dashboard-intro">Управляйте данными своего профиля.</p>
            {profile && <form className="profile-card profile-form" onSubmit={saveProfile}>
              <h2>Данные профиля</h2>
              <label htmlFor="profile-name">Имя профиля</label>
              <input id="profile-name" value={name} onChange={(event) => { setName(event.target.value); setSaveMessage(""); }} maxLength={200} required />
              <p>Логин для входа: <strong>{profile.login}</strong></p>
              <button type="submit" disabled={saving}>{saving ? "Сохраняем…" : "Сохранить"}</button>
              {saveMessage && <p role="status">{saveMessage}</p>}
            </form>}
          </>}
          {profileError && <div role="alert" className="profile-error">{profileError} <button type="button" onClick={() => { setProfileError(""); setAttempt((n) => n + 1); }}>Повторить</button></div>}
          {error && <p className="profile-error" role="alert">{error}</p>}
        </div>
      </section>
    </main>
  );
}
