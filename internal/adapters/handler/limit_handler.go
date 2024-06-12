package handler

import (
	"net/http"
	"yadro-project/internal/core/services"
)

type LimitHandler struct {
	limitService services.LimitService
	authService  services.AuthService
}

func NewLimitHandler(limitService services.LimitService, authService services.AuthService) *LimitHandler {
	return &LimitHandler{
		limitService: limitService,
		authService:  authService,
	}
}

func (h *LimitHandler) LimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := getToken(r.Header.Get(authorizationHeader))
		if err != nil {
			HandleError(w, http.StatusUnauthorized, err)
			return
		}

		email, err := h.authService.GetEmailFromToken(token)
		if err != nil {
			HandleError(w, http.StatusUnauthorized, err)
			return
		}
		reservation, err := h.limitService.Limit(email)
		if err != nil {
			HandleError(w, http.StatusTooManyRequests, err)
			return
		}

		next.ServeHTTP(w, r)

		h.limitService.Done(reservation)
	})
}
