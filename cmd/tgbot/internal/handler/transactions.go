package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"tgfinance/cmd/tgbot/internal/auth"
	"tgfinance/internal/finance"
)

// Transactions handles CRUD for financial records.
type Transactions struct {
	auth  *auth.Validator
	store *finance.Store
}

func NewTransactions(a *auth.Validator, s *finance.Store) *Transactions {
	return &Transactions{auth: a, store: s}
}

// List handles GET /api/transactions
func (h *Transactions) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	txns, err := h.store.List(userID, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if txns == nil {
		txns = []finance.Transaction{}
	}
	writeJSON(w, txns)
}

// Create handles POST /api/transactions
func (h *Transactions) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var t finance.Transaction
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if t.Amount <= 0 {
		writeErr(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	if t.Type != "income" && t.Type != "expense" {
		writeErr(w, http.StatusBadRequest, "type must be income or expense")
		return
	}
	if t.Currency == "" {
		t.Currency = "USD"
	}
	t.UserID = userID
	t.Source = "manual"

	id, err := h.store.Add(t)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{"id": id})
}

// Delete handles DELETE /api/transactions/{id}
func (h *Transactions) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/transactions/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.store.Delete(userID, id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Clear handles DELETE /api/transactions?type=income|expense|all
func (h *Transactions) Clear(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	txType := r.URL.Query().Get("type") // "income", "expense", or "" = all
	if err := h.store.Clear(userID, txType); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
