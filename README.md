# Uptime Up

## Локальный запуск

Нужны Go 1.22 или новее, Node.js, npm, Docker с Compose и утилита `setsid`. Сначала установите зависимости frontend:

```sh
cd frontend && npm ci
```

Из корня проекта запустите:

```sh
./scripts/dev.sh
```

Скрипт запускает PostgreSQL в Docker, ожидает его готовности, применяет миграции и запускает backend и frontend одновременно. Откройте <http://localhost:3000>. Ctrl+C останавливает оба сервера; PostgreSQL и его данные сохраняются. Чтобы остановить базу, выполните `docker compose stop` из каталога `backend/`.

Для локального запуска используются стандартные параметры PostgreSQL из `backend/docker-compose.yml` и тестовый секрет JWT. Можно задать `DATABASE_URL`, `JWT_SECRET`, `POSTGRES_PORT`, `HTTP_ADDR` и `BACKEND_URL` через переменные окружения. При изменении `POSTGRES_USER`, `POSTGRES_PASSWORD` или `POSTGRES_DB` необходимо также явно указать соответствующий `DATABASE_URL`. При изменении адреса backend задайте соответствующий `BACKEND_URL`. Стандартные учётные данные и секрет предназначены только для разработки.

Подробнее о настройке и API: [backend](backend/README.md), [frontend](frontend/README.md).
