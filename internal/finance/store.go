// Package finance provides SQLite-backed storage for income/expense tracking.
package finance

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

// Transaction is a single financial record.
type Transaction struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Type        string    `json:"type"` // "income" | "expense"
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Source      string    `json:"source"`    // "manual" | "claude"
	ImportKey   string    `json:"import_key,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Stats is aggregated financial data for a user.
type Stats struct {
	TotalIncome   float64 `json:"total_income"`
	TotalExpenses float64 `json:"total_expenses"`
	Balance       float64 `json:"balance"`
	MonthIncome   float64 `json:"month_income"`
	MonthExpenses float64 `json:"month_expenses"`
}

// Store manages transactions in SQLite.
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
		CREATE TABLE IF NOT EXISTS transactions (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id     INTEGER NOT NULL,
			type        TEXT    NOT NULL,
			amount      REAL    NOT NULL,
			currency    TEXT    NOT NULL DEFAULT 'USD',
			category    TEXT    NOT NULL DEFAULT '',
			description TEXT    NOT NULL DEFAULT '',
			source      TEXT    NOT NULL DEFAULT 'manual',
			import_key  TEXT    NOT NULL DEFAULT '',
			created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_txn_user
			ON transactions(user_id, created_at DESC);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_import_key
			ON transactions(import_key) WHERE import_key != '';
	`)
	return err
}

// Add inserts a transaction, ignoring duplicates by import_key.
// Returns the new row ID (0 if ignored as duplicate).
func (s *Store) Add(t Transaction) (int64, error) {
	res, err := s.db.Exec(`
		INSERT OR IGNORE INTO transactions
			(user_id, type, amount, currency, category, description, source, import_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.UserID, t.Type, t.Amount, t.Currency,
		t.Category, t.Description, t.Source, t.ImportKey,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// List returns the last limit transactions for a user, newest first.
func (s *Store) List(userID int64, limit int) ([]Transaction, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, type, amount, currency, category, description, source, created_at
		FROM transactions
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Type, &t.Amount, &t.Currency,
			&t.Category, &t.Description, &t.Source, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Delete removes a transaction that belongs to userID.
func (s *Store) Delete(userID, id int64) error {
	_, err := s.db.Exec(
		`DELETE FROM transactions WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

// Clear deletes all transactions for a user, optionally filtered by type ("income", "expense", or "" for all).
func (s *Store) Clear(userID int64, txType string) error {
	if txType == "income" || txType == "expense" {
		_, err := s.db.Exec(`DELETE FROM transactions WHERE user_id = ? AND type = ?`, userID, txType)
		return err
	}
	_, err := s.db.Exec(`DELETE FROM transactions WHERE user_id = ?`, userID)
	return err
}

// Stats returns all-time and current-month totals for a user.
func (s *Store) Stats(userID int64) (Stats, error) {
	var st Stats
	if err := s.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN type='income'  THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END), 0)
		FROM transactions WHERE user_id = ?`, userID,
	).Scan(&st.TotalIncome, &st.TotalExpenses); err != nil {
		return st, err
	}
	st.Balance = st.TotalIncome - st.TotalExpenses

	err := s.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN type='income'  THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END), 0)
		FROM transactions
		WHERE user_id = ?
		  AND strftime('%Y-%m', created_at) = strftime('%Y-%m', 'now')`,
		userID,
	).Scan(&st.MonthIncome, &st.MonthExpenses)
	return st, err
}
