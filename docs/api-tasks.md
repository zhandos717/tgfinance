# Tasks API

Base URL: `https://zhandos.top`

## Authentication

Two methods are supported:

| Method | Header | Use case |
|--------|--------|----------|
| Telegram Mini App | `X-Init-Data: <initData>` | Browser / Mini App |
| Internal key | `X-Internal-Key: <key>` | Server-to-server (ZeroClaw) |

---

## Endpoints

### GET /api/tasks

Returns task list for the authenticated user.

**Query params:**

| Param | Values | Description |
|-------|--------|-------------|
| `status` | `todo` \| `in_progress` \| `done` | Filter by status. Omit for all. |

**Response `200`:**
```json
[
  {
    "id": 1,
    "user_id": 480568670,
    "title": "Задеплоить проект Каспи QR",
    "description": "",
    "status": "todo",
    "priority": "high",
    "created_at": "2026-03-15T13:51:00Z",
    "updated_at": "2026-03-15T13:51:00Z"
  }
]
```

**Example:**
```bash
curl https://zhandos.top/api/tasks \
  -H "X-Internal-Key: <key>"

curl "https://zhandos.top/api/tasks?status=todo" \
  -H "X-Internal-Key: <key>"
```

---

### POST /api/tasks

Creates a new task.

**Body:**
```json
{
  "title": "Название задачи",
  "description": "Подробности (необязательно)",
  "status": "todo",
  "priority": "medium"
}
```

| Field | Type | Required | Values |
|-------|------|----------|--------|
| `title` | string | ✅ | any |
| `description` | string | — | any |
| `status` | string | — | `todo` (default) |
| `priority` | string | — | `low` \| `medium` (default) \| `high` |

**Response `201`:**
```json
{ "id": 2 }
```

**Example:**
```bash
curl -X POST https://zhandos.top/api/tasks \
  -H "Content-Type: application/json" \
  -H "X-Internal-Key: <key>" \
  -d '{"title":"Обновить nginx","priority":"high"}'
```

---

### PATCH /api/tasks/{id}/status

Updates task status.

**Body:**
```json
{ "status": "done" }
```

| Value | Meaning |
|-------|---------|
| `todo` | Не начато |
| `in_progress` | В процессе |
| `done` | Выполнено |

**Response:** `204 No Content`

**Example:**
```bash
curl -X PATCH https://zhandos.top/api/tasks/1/status \
  -H "Content-Type: application/json" \
  -H "X-Internal-Key: <key>" \
  -d '{"status":"done"}'
```

---

### PATCH /api/tasks/{id}

Updates task fields (title, description, status, priority).

**Body:** same fields as POST (all optional)

**Response:** `204 No Content`

**Example:**
```bash
curl -X PATCH https://zhandos.top/api/tasks/1 \
  -H "Content-Type: application/json" \
  -H "X-Internal-Key: <key>" \
  -d '{"title":"Новое название","priority":"low"}'
```

---

### DELETE /api/tasks/{id}

Deletes a task.

**Response:** `204 No Content`

**Example:**
```bash
curl -X DELETE https://zhandos.top/api/tasks/1 \
  -H "X-Internal-Key: <key>"
```

---

### GET /api/tasks/stats

Returns task counts grouped by status.

**Response `200`:**
```json
{
  "todo": 3,
  "in_progress": 1,
  "done": 12
}
```

**Example:**
```bash
curl https://zhandos.top/api/tasks/stats \
  -H "X-Internal-Key: <key>"
```

---

## Status flow

```
todo → in_progress → done
 ↑________________________|
```

## Priority levels

| Value | Meaning |
|-------|---------|
| `high` | Срочно / важно |
| `medium` | Обычная задача |
| `low` | Не срочно |

## Error responses

```json
{ "error": "unauthorized" }        // 401 — неверный ключ
{ "error": "title is required" }   // 400 — валидация
{ "error": "invalid id" }          // 400 — неверный ID
```
