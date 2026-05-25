package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/example/search-trends/internal/stoplist"
	"github.com/example/search-trends/internal/window"
)

type Handler struct {
	win *window.Window
	sl  *stoplist.StopList
}

func NewHandler(win *window.Window, sl *stoplist.StopList) *Handler {
	return &Handler{win: win, sl: sl}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/top", h.handleTop)
	mux.HandleFunc("/api/v1/stoplist", h.handleStoplist)
}

func (h *Handler) handleTop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	top := h.win.Top()
	if limit < len(top) {
		top = top[:limit]
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(top)
}

func (h *Handler) handleStoplist(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req struct {
			Word string `json:"word"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Word == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		h.sl.Add(req.Word)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "added"})
	case http.MethodDelete:
		word := r.URL.Query().Get("word")
		if word == "" {
			http.Error(w, "word parameter required", http.StatusBadRequest)
			return
		}
		h.sl.Remove(word)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "removed"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
