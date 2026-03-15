package handler

import (
	"net/http"

	"tgfinance/cmd/tgbot/internal/auth"
	"tgfinance/internal/finance"
)

// Stats handles aggregated financial data.
type Stats struct {
	auth  *auth.Validator
	store *finance.Store
}

func NewStats(a *auth.Validator, s *finance.Store) *Stats {
	return &Stats{auth: a, store: s}
}

func (h *Stats) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, ok := h.auth.UserFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	stats, err := h.store.Stats(userID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, stats)
}
