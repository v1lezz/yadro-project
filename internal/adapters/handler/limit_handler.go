package handler

import (
	"net/http"

	"go.uber.org/ratelimit"
)

type Semaphore struct {
	ch chan struct{}
}

func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.ch
}

func NewSemaphore(limit int) Semaphore {
	return Semaphore{
		ch: make(chan struct{}, limit),
	}
}

type LimitHandler struct {
	rl ratelimit.Limiter
	cl Semaphore
}

func NewLimitHandler(rateLimit, concurrencyLimit int) *LimitHandler {
	return &LimitHandler{
		rl: ratelimit.New(rateLimit),
		cl: NewSemaphore(concurrencyLimit),
	}
}

func (h *LimitHandler) LimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
