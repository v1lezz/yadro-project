package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"yadro-project/internal/core/services"
)

type ComicsHandler struct {
	svc   services.ComicsService
	mutex *sync.Mutex
}

func NewComicsHandler(svc services.ComicsService, mutex *sync.Mutex) *ComicsHandler {
	return &ComicsHandler{
		svc:   svc,
		mutex: mutex,
	}
}

var (
	errQueryIsEmpty = errors.New("query \"search\" is empty")
	errEncodeJSON   = errors.New("error encode json")
	errAccepted     = errors.New("update already started")
)

func (h *ComicsHandler) GetComics(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	search := queries.Get("search")
	if search == "" {
		HandleError(w, http.StatusBadRequest, errQueryIsEmpty)
		return
	}
	comics, err := h.svc.GetComics(r.Context(), search)
	if err != nil {
		if errors.Is(err, services.ErrContextDone) {
			HandleError(w, http.StatusServiceUnavailable, err)
		} else {
			HandleError(w, http.StatusInternalServerError, fmt.Errorf("error get comics from server: %w", err))
		}
		return
	}
	for i := 0; i < len(comics); i++ {
		comics[i].Keywords = nil
	}
	json.NewEncoder(w).Encode(comics)
}

func (h *ComicsHandler) UpdateComics(w http.ResponseWriter, r *http.Request) {
	if !h.mutex.TryLock() {
		HandleError(w, http.StatusAccepted, errAccepted)
		return
	}
	defer h.mutex.Unlock()
	meta, err := h.svc.UpdateComics(r.Context())
	if err != nil {
		if errors.Is(err, services.ErrContextDone) {
			HandleError(w, http.StatusServiceUnavailable, err)
			return
		}
		HandleError(w, http.StatusInternalServerError, fmt.Errorf("error update comics: %w", err))
		return
	}
	json.NewEncoder(w).Encode(meta)
}
