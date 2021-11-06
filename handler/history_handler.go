package handler

import (
	"encoding/json"
	"github.com/abdoub/location-history/store"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type HistoryHandler struct {
	log *log.Logger
	store store.Store
}

func NewHistoryHandler(log *log.Logger, store store.Store) *HistoryHandler {
	return &HistoryHandler{log: log, store: store}
}

func (h *HistoryHandler) Dispatch(pathPrefix string) (string, http.HandlerFunc) {
	return pathPrefix, func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, pathPrefix)
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodPost:
			h.post(id, w, r)
		case http.MethodGet:
			h.get(id, w, r)
		case http.MethodDelete:
			h.delete(id, w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func (h *HistoryHandler) post(id string, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var location store.Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	created,_, err := h.store.Append(id, location)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if created {
		w.WriteHeader(http.StatusCreated)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *HistoryHandler) get(id string, w http.ResponseWriter, r *http.Request) {
	max := r.URL.Query().Get("max")
	var limit int
	limit, _ = strconv.Atoi(max)
	locations, err := h.store.Get(id, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := struct {
		OrderID string           `json:"order_id"`
		History []store.Location `json:"history"`
	}{
		OrderID: id,
		History: locations,
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *HistoryHandler) delete(id string, w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}