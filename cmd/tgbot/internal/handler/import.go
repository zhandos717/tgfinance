package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"tgfinance/cmd/tgbot/internal/auth"
	"tgfinance/internal/finance"
)

// Import handles importing usage data from external sources.
type Import struct {
	auth     *auth.Validator
	store    *finance.Store
	usageLog string
}

func NewImport(a *auth.Validator, s *finance.Store, usageLog string) *Import {
	return &Import{auth: a, store: s, usageLog: usageLog}
}

type usageRecord struct {
	TS    string `json:"ts"`
	Model string `json:"model"`
	In    int    `json:"in"`
	Out   int    `json:"out"`
}

// Claude handles POST /api/import/claude
func (h *Import) Claude(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID, ok := h.auth.UserFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	f, err := os.Open(h.usageLog)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "cannot open usage log: "+err.Error())
		return
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	var imported, skipped int
	for dec.More() {
		var rec usageRecord
		if err := dec.Decode(&rec); err != nil {
			continue
		}
		inCost, outCost := modelCostUSD(rec.Model)
		cost := round2dp(float64(rec.In)*inCost + float64(rec.Out)*outCost)
		if cost < 0.00001 {
			continue
		}
		t := finance.Transaction{
			UserID:      userID,
			Type:        "expense",
			Amount:      cost,
			Currency:    "USD",
			Category:    "Claude API",
			Description: fmt.Sprintf("%s — %d вх + %d исх токенов", rec.Model, rec.In, rec.Out),
			Source:      "claude",
			ImportKey:   fmt.Sprintf("claude:%s:%d:%d", rec.TS, rec.In, rec.Out),
		}
		id, err := h.store.Add(t)
		if err != nil {
			log.Printf("import claude: %v", err)
		} else if id == 0 {
			skipped++
		} else {
			imported++
		}
	}
	writeJSON(w, map[string]any{"imported": imported, "skipped": skipped})
}

// modelCostUSD returns (input $/token, output $/token) for the given model.
func modelCostUSD(model string) (float64, float64) {
	switch {
	case strings.Contains(model, "opus"):
		return 15.0 / 1e6, 75.0 / 1e6
	case strings.Contains(model, "haiku"):
		return 0.80 / 1e6, 4.0 / 1e6
	default: // sonnet
		return 3.0 / 1e6, 15.0 / 1e6
	}
}

func round2dp(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
