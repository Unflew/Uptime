# Backend

Требуются Go 1.22 или новее и Docker Compose. Команды выполняются из каталога `backend/`.

## Запуск

```sh
docker compose up -d --wait
export DATABASE_URL='postgres://uptime_up:uptime_up@localhost:5432/uptime_up?sslmode=disable'
export JWT_SECRET='replace-this-with-a-random-secret-at-least-32-bytes'
go run ./cmd/migrate
go run ./cmd/server
```

Пароль и пользователь PostgreSQL для локальной разработки задаются переменными `POSTGRES_PASSWORD`, `POSTGRES_USER`, имя базы — `POSTGRES_DB`, порт — `POSTGRES_PORT`. Изменение этих переменных не изменяет существующего пользователя или базу в уже созданном томе PostgreSQL. Указанные значения по умолчанию предназначены только для локальной разработки. Для других сред используйте безопасные учётные данные, защищённое соединение и случайный секрет JWT.

HTTP слушает `:8080`; адрес задаётся переменной `HTTP_ADDR`. Сроки задаются переменными `ACCESS_TTL` и `REFRESH_TTL` в формате Go duration (по умолчанию `15m` и `720h`). Миграции в `migrations/` применяются один раз командой выше; запускайте её из каталога `backend/` перед запуском сервера.

## API

Все запросы и ответы используют JSON. Логин регистрозависим, содержит от 3 до 64 байт без пробельных символов по краям и ASCII-пробелов внутри; пароль содержит от 8 до 72 байт. Ограничение пароля связано с bcrypt.

- `POST /auth/register`: `{"login":"alice","password":"password123"}` — ответ 201.
- `POST /auth/login`: то же тело — ответ 200.
- `POST /auth/refresh`: `{"refresh_token":"..."}` — ответ 200 с новой парой; прежний refresh-токен больше не действует.

Успешный ответ: `{"access_token":"...","refresh_token":"...","token_type":"Bearer","expires_in":900}`. Ошибки возвращаются как `{"error":"..."}`. Занятый логин возвращает 409, неверные учётные данные или refresh-токен — 401, некорректный запрос — 400. Access-токен передаётся защищённым обработчикам в заголовке `Authorization: Bearer <access_token>`; middleware доступен через `RequireAuth`.

## Проверка

```sh
go test ./...
go build ./...
```

Для интеграционного теста создайте отдельную тестовую базу, примените к ней миграции через `DATABASE_URL`, затем запустите тесты с `TEST_DATABASE_URL`, указывающим на эту базу. Тест создаёт и удаляет собственную учётную запись.
