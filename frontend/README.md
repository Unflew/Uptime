# Frontend

Next.js-приложение для входа и регистрации в Uptime Up. Запускайте команды из каталога `frontend/`.

```sh
npm ci
npm run dev
```

Откройте `http://localhost:3000`. Backend и PostgreSQL должны быть запущены; инструкции находятся в `../backend/README.md`. По умолчанию сервер Next.js обращается к backend по адресу `http://127.0.0.1:8080`. Для другого адреса задайте серверную переменную окружения `BACKEND_URL` перед запуском Next.js.

Токены хранятся в HTTP-only cookies. Если access-токен истёк, приложение пытается обновить сессию через refresh-токен при открытии страницы. Время жизни refresh-cookie по умолчанию составляет 30 дней; если на backend изменён `REFRESH_TTL`, задайте соответствующий срок в секундах через `REFRESH_COOKIE_MAX_AGE` на сервере Next.js. В production cookies требуют HTTPS. Выход удаляет cookies в браузере; backend пока не поддерживает отзыв сессии при выходе.

Проверка production-сборки: `npm run build`.
