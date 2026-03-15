package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"tgfinance/cmd/tgbot/internal/auth"
	"tgfinance/internal/tasks"
)

// Tasks handles CRUD for task tracking.
type Tasks struct {
	auth  *auth.Validator
	store *tasks.Store
}

func NewTasks(a *auth.Validator, s *tasks.Store) *Tasks {
	return &Tasks{auth: a, store: s}
}

// List handles GET /api/tasks?status=todo|in_progress|done
func (h *Tasks) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	status := r.URL.Query().Get("status")
	list, err := h.store.List(userID, status)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []tasks.Task{}
	}
	writeJSON(w, list)
}

// Create handles POST /api/tasks
func (h *Tasks) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var t tasks.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(t.Title) == "" {
		writeErr(w, http.StatusBadRequest, "title is required")
		return
	}
	if t.Status == "" {
		t.Status = "todo"
	}
	if t.Priority == "" {
		t.Priority = "medium"
	}
	t.UserID = userID

	id, err := h.store.Add(t)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{"id": id})
}

// Update handles PATCH /api/tasks/{id}
func (h *Tasks) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := idFromPath(r.URL.Path, "/api/tasks/")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	var t tasks.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := h.store.Update(userID, id, t); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SetStatus handles PATCH /api/tasks/{id}/status
func (h *Tasks) SetStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// path: /api/tasks/{id}/status
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/tasks/"), "/")
	if len(parts) < 2 {
		writeErr(w, http.StatusBadRequest, "invalid path")
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Status != "todo" && body.Status != "in_progress" && body.Status != "done" {
		writeErr(w, http.StatusBadRequest, "status must be todo, in_progress or done")
		return
	}
	if err := h.store.SetStatus(userID, id, body.Status); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /api/tasks/{id}
func (h *Tasks) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := idFromPath(r.URL.Path, "/api/tasks/")
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

// Stats handles GET /api/tasks/stats
func (h *Tasks) Stats(w http.ResponseWriter, r *http.Request) {
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

func idFromPath(path, prefix string) (int64, error) {
	s := strings.TrimPrefix(path, prefix)
	s = strings.Split(s, "/")[0]
	return strconv.ParseInt(s, 10, 64)
}
