// Package tasks provides SQLite-backed storage for task tracking.
package tasks

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

// Task represents a single to-do item.
type Task struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`   // "todo" | "in_progress" | "done"
	Priority    string    `json:"priority"` // "low" | "medium" | "high"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store manages tasks in SQLite.
type Store struct {
	db *sql.DB
}

// New opens (or creates) the SQLite database at path.
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	return s, s.migrate()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id     INTEGER NOT NULL,
			title       TEXT    NOT NULL,
			description TEXT    NOT NULL DEFAULT '',
			status      TEXT    NOT NULL DEFAULT 'todo',
			priority    TEXT    NOT NULL DEFAULT 'medium',
			created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_tasks_user
			ON tasks(user_id, status, created_at DESC);
	`)
	return err
}

// Add inserts a new task and returns its ID.
func (s *Store) Add(t Task) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO tasks (user_id, title, description, status, priority)
		VALUES (?, ?, ?, ?, ?)`,
		t.UserID, t.Title, t.Description, t.Status, t.Priority,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// List returns tasks for a user, optionally filtered by status.
func (s *Store) List(userID int64, status string) ([]Task, error) {
	var rows *sql.Rows
	var err error
	if status != "" {
		rows, err = s.db.Query(`
			SELECT id, user_id, title, description, status, priority, created_at, updated_at
			FROM tasks WHERE user_id = ? AND status = ?
			ORDER BY
				CASE priority WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END,
				created_at DESC`, userID, status)
	} else {
		rows, err = s.db.Query(`
			SELECT id, user_id, title, description, status, priority, created_at, updated_at
			FROM tasks WHERE user_id = ?
			ORDER BY
				CASE status WHEN 'in_progress' THEN 0 WHEN 'todo' THEN 1 ELSE 2 END,
				CASE priority WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END,
				created_at DESC`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Update patches title, description, status, and priority of a task.
func (s *Store) Update(userID, id int64, t Task) error {
	_, err := s.db.Exec(`
		UPDATE tasks SET title=?, description=?, status=?, priority=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=? AND user_id=?`,
		t.Title, t.Description, t.Status, t.Priority, id, userID,
	)
	return err
}

// SetStatus is a shortcut to update only the status field.
func (s *Store) SetStatus(userID, id int64, status string) error {
	_, err := s.db.Exec(`
		UPDATE tasks SET status=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=? AND user_id=?`,
		status, id, userID,
	)
	return err
}

// Delete removes a task belonging to userID.
func (s *Store) Delete(userID, id int64) error {
	_, err := s.db.Exec(`DELETE FROM tasks WHERE id=? AND user_id=?`, id, userID)
	return err
}

// Stats returns counts per status for a user.
func (s *Store) Stats(userID int64) (map[string]int, error) {
	rows, err := s.db.Query(`
		SELECT status, COUNT(*) FROM tasks WHERE user_id=? GROUP BY status`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{"todo": 0, "in_progress": 0, "done": 0}
	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		out[status] = count
	}
	return out, rows.Err()
}
