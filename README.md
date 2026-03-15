# tgfinance

Telegram Mini App backend — трекер финансов и задач.

## Стек

- **Go 1.24** — HTTP-сервер (`net/http`)
- **SQLite** (`modernc.org/sqlite`) — хранилище без внешних зависимостей
- **Telegram Bot API** — webhook + Mini App

## Структура

```
cmd/tgbot/          — точка входа, роутер, хендлеры
  webapp/           — встроенный HTML Mini App (embed)
  internal/
    auth/           — валидация Telegram initData и internal key
    handler/        — transactions, stats, tasks, import, webhook
    router/         — сборка ServeMux
internal/
  finance/          — SQLite store для транзакций
  tasks/            — SQLite store для задач
deploy/
  tgbot.service     — systemd unit
docs/
  api-tasks.md      — документация Tasks API
```

## Переменные окружения

| Переменная        | По умолчанию              | Описание                            |
|-------------------|---------------------------|-------------------------------------|
| `TG_BOT_TOKEN`    | —                         | Токен бота от @BotFather (обязателен) |
| `TG_APP_URL`      | `https://zhandos.top/app` | Публичный URL Mini App              |
| `FINANCE_DB`      | `finance.db`              | Путь к SQLite-базе                  |
| `BOT_PORT`        | `4002`                    | Порт HTTP-сервера                   |
| `INTERNAL_API_KEY`| —                         | Ключ для server-to-server запросов  |
| `INTERNAL_USER_ID`| `0`                       | Telegram user ID владельца          |
| `USAGE_LOG`       | `claude-usage.jsonl`      | Путь к логу использования Claude    |

## Сборка и деплой

```bash
# Локальная сборка
make build

# Кросс-компиляция под Linux amd64
make linux

# Деплой на сервер
make deploy SERVER=root@host

# Тесты
make test
```

После деплоя бинарник кладётся в `/opt/aigate-bot`, сервис управляется через systemd:

```bash
systemctl restart aigate-bot
systemctl status aigate-bot
journalctl -u aigate-bot -f
```

Секреты хранятся в `/etc/aigate/secrets` (см. `deploy/tgbot.service`).

## API

Base URL: `https://zhandos.top`

### Аутентификация

| Метод            | Заголовок                     | Использование             |
|------------------|-------------------------------|---------------------------|
| Telegram Mini App | `X-Init-Data: <initData>`    | Браузер / Mini App        |
| Internal key      | `X-Internal-Key: <key>`      | Server-to-server          |

### Финансы

| Метод    | Путь                       | Описание                          |
|----------|----------------------------|-----------------------------------|
| `GET`    | `/api/stats`               | Статистика доходов и расходов     |
| `GET`    | `/api/transactions`        | Список транзакций                 |
| `POST`   | `/api/transactions`        | Добавить транзакцию               |
| `DELETE` | `/api/transactions/{id}`   | Удалить транзакцию                |
| `DELETE` | `/api/transactions`        | Очистить все транзакции           |
| `POST`   | `/api/import/claude`       | Импорт через Claude AI            |

### Задачи

| Метод    | Путь                       | Описание                          |
|----------|----------------------------|-----------------------------------|
| `GET`    | `/api/tasks`               | Список задач (фильтр по `status`) |
| `POST`   | `/api/tasks`               | Создать задачу                    |
| `PATCH`  | `/api/tasks/{id}`          | Обновить задачу                   |
| `PATCH`  | `/api/tasks/{id}/status`   | Обновить статус задачи            |
| `DELETE` | `/api/tasks/{id}`          | Удалить задачу                    |
| `GET`    | `/api/tasks/stats`         | Счётчики по статусам              |

Подробная документация Tasks API — в [`docs/api-tasks.md`](docs/api-tasks.md).

### Mini App

`GET /app` — отдаёт встроенный HTML интерфейс.

`POST /webhook` — Telegram webhook.
