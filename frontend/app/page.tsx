"use client";

import { FormEvent, useEffect, useState } from "react";

type Mode = "login" | "register";

let sessionRequest: Promise<boolean> | null = null;
function restoreSession() {
  if (!sessionRequest) {
    sessionRequest = fetch("/api/auth/session", { cache: "no-store" })
      .then(async (response) => {
        if (!response.ok) return false;
        const session = await response.json();
        if (session.authenticated) return true;
        if (!session.canRefresh) return false;
        return (await fetch("/api/auth/refresh", { method: "POST" })).ok;
      })
      .catch(() => false)
      .finally(() => { sessionRequest = null; });
  }
  return sessionRequest;
}

function Mark() {
  return <span className="mark" aria-hidden="true"><svg viewBox="0 0 24 24" fill="none"><path d="M3 12h4l2-5 4 10 2-5h6" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" /></svg></span>;
}

export default function Home() {
  const [mode, setMode] = useState<Mode>("login");
  const [login, setLogin] = useState("");
  const [password, setPassword] = useState("");
  const [visible, setVisible] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [authenticated, setAuthenticated] = useState(false);
  const [checkingSession, setCheckingSession] = useState(true);

  useEffect(() => {
    let active = true;
    restoreSession().then((result) => {
      if (active) { setAuthenticated(result); setCheckingSession(false); }
    });
    return () => { active = false; };
  }, []);

  function switchMode(next: Mode) {
    setMode(next);
    setError("");
    setPassword("");
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    if (new TextEncoder().encode(login).length < 3 || new TextEncoder().encode(login).length > 64 || /\s/.test(login)) {
      setError("Логин должен содержать от 3 до 64 байт без пробелов.");
      return;
    }
    const passwordLength = new TextEncoder().encode(password).length;
    if (passwordLength < 8 || passwordLength > 72) {
      setError("Пароль должен содержать от 8 до 72 байт.");
      return;
    }
    setLoading(true);
    try {
      const response = await fetch(`/api/auth/${mode}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ login, password }),
      });
      if (response.ok) {
        setPassword("");
        setAuthenticated(true);
      } else if (response.status === 409) {
        setError("Этот логин уже занят. Попробуйте другой.");
      } else if (response.status === 401) {
        setError("Неверный логин или пароль.");
      } else if (response.status === 400) {
        setError("Проверьте логин и пароль и попробуйте снова.");
      } else {
        setError("Не удалось связаться с сервером. Попробуйте ещё раз.");
      }
    } catch {
      setError("Не удалось связаться с сервером. Попробуйте ещё раз.");
    } finally {
      setLoading(false);
    }
  }

  async function logout() {
    setLoading(true);
    try {
      const response = await fetch("/api/auth/logout", { method: "POST" });
      if (!response.ok) throw new Error();
      setAuthenticated(false);
      setPassword("");
      setError("");
    } catch {
      setError("Не удалось выйти. Попробуйте ещё раз.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="shell">
      <section className="story" aria-label="О сервисе Uptime Up">
        <div className="story-grid" aria-hidden="true" />
        <div className="brand brand-light"><Mark /><span>uptime<span className="brand-accent">up</span><span className="brand-dot">.</span></span></div>
        <div className="story-content">
          <div className="eyebrow"><span className="eyebrow-line" /> МОНИТОРИНГ БЕЗ ПЕРЕРЫВОВ</div>
          <h1>Ваши сервисы.<br /><span>Под присмотром.</span></h1>
          <p>Место, где спокойствие начинается с уверенности, что всё работает.</p>
          <div className="pulse-card" aria-hidden="true">
            <div className="pulse-heading"><span><i /> СТАТУС СИСТЕМЫ</span><span>LIVE MONITORING</span></div>
            <div className="pulse-graph"><svg viewBox="0 0 440 100" preserveAspectRatio="none"><defs><linearGradient id="fade" x1="0" x2="0" y1="0" y2="1"><stop offset="0" stopColor="#66e4bd" stopOpacity=".22" /><stop offset="1" stopColor="#66e4bd" stopOpacity="0" /></linearGradient></defs><path d="M0 75 L33 74 51 69 76 72 100 63 125 67 147 54 167 60 185 54 210 58 230 40 252 48 276 34 294 41 318 30 340 33 361 20 382 26 405 14 440 11 V100 H0Z" fill="url(#fade)" /><path d="M0 75 L33 74 51 69 76 72 100 63 125 67 147 54 167 60 185 54 210 58 230 40 252 48 276 34 294 41 318 30 340 33 361 20 382 26 405 14 440 11" fill="none" stroke="#75e7c5" strokeWidth="2" vectorEffect="non-scaling-stroke" /></svg></div>
            <div className="pulse-footer"><span>Всё под контролем</span><span>КРУГЛОСУТОЧНО</span></div>
          </div>
        </div>
        <div className="story-bottom"><span>СЛЕДИТЕ ЗА ВАЖНЫМ</span><span>UPTIME UP © {new Date().getFullYear()}</span></div>
      </section>
      <section className="auth-side">
        <div className="mobile-brand brand"><Mark /><span>uptime<span className="brand-accent">up</span><span className="brand-dot">.</span></span></div>
        <div className="auth-wrap">
          {checkingSession ? (
            <p className="subtitle" role="status">Проверяем сессию…</p>
          ) : authenticated ? (
            <div className="success" role="status">
              <span className="success-icon">✓</span>
              <div className="section-label">ВСЁ ГОТОВО</div>
              <h2>Вы вошли в аккаунт</h2>
              <p>{login ? `Рады видеть вас, ${login}.` : "Рады видеть вас снова."}</p>
              <button className="primary" type="button" onClick={logout} disabled={loading}>{loading ? "Выходим…" : "Выйти"}</button>
              {error && <p className="error" role="alert">{error}</p>}
            </div>
          ) : (
            <>
              <div className="section-label">ДОБРО ПОЖАЛОВАТЬ</div>
              <h2>{mode === "login" ? "С возвращением" : "Создать аккаунт"}<span className="heading-dot">.</span></h2>
              <p className="subtitle">{mode === "login" ? "Войдите, чтобы продолжить работу." : "Зарегистрируйтесь, чтобы начать работу."}</p>
              <div className="tabs" role="group" aria-label="Выбор действия">
                <button type="button" className={mode === "login" ? "active" : ""} aria-pressed={mode === "login"} onClick={() => switchMode("login")}>Вход</button>
                <button type="button" className={mode === "register" ? "active" : ""} aria-pressed={mode === "register"} onClick={() => switchMode("register")}>Регистрация</button>
              </div>
              <form onSubmit={submit}>
                <label htmlFor="login">Логин</label>
                <input id="login" name="username" type="text" autoComplete="username" placeholder="Введите ваш логин" value={login} onChange={(e) => setLogin(e.target.value)} required />
                <label htmlFor="password">Пароль</label>
                <div className="password-field"><input id="password" name="password" type={visible ? "text" : "password"} autoComplete={mode === "login" ? "current-password" : "new-password"} placeholder="Введите пароль" value={password} onChange={(e) => setPassword(e.target.value)} required /><button type="button" onClick={() => setVisible(!visible)} aria-label={visible ? "Скрыть пароль" : "Показать пароль"} aria-pressed={visible}>{visible ? "Скрыть" : "Показать"}</button></div>
                {mode === "register" && <p className="hint">От 8 до 72 байт. Используйте надёжный пароль.</p>}
                {error && <p className="error" role="alert">{error}</p>}
                <button className="primary" type="submit" disabled={loading}>{loading ? "Подождите…" : mode === "login" ? "Войти" : "Создать аккаунт"}</button>
              </form>
              <div className="switch-prompt">{mode === "login" ? "Ещё нет аккаунта?" : "Уже есть аккаунт?"} <button type="button" onClick={() => switchMode(mode === "login" ? "register" : "login")}>{mode === "login" ? "Зарегистрироваться" : "Войти"}</button></div>
            </>
          )}
        </div>
        <div className="auth-footer"><span>UPTIME UP</span><span>Создано для спокойной работы</span></div>
      </section>
    </main>
  );
}
