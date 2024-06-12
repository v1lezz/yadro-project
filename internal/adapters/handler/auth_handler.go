package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"yadro-project/internal/core/domain"
	"yadro-project/internal/core/services"
)

const ( // headers
	authorizationHeader = "Authorization"
)

var ( //errors
	errAuthorizationHeaderIsEmpty = errors.New("authorization header is empty")
	errHeaderIsNotRequiredMask    = errors.New("authorization header must be required by mask \"Beaver <token>\"")
	errUserIsNotExist             = errors.New("user is not exist")
)

type AuthHandler struct {
	svc services.AuthService
}

func NewAuthHandler(svc services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	req := domain.LoginRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, http.StatusBadRequest, fmt.Errorf("error decode json: %w", err))
		return
	}

	t, err := h.svc.Login(req)

	if err != nil {
		if errors.Is(err, services.ErrBadCredentials) {
			HandleError(w, http.StatusUnauthorized, err)
			return
		}
		HandleError(w, http.StatusInternalServerError, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": t,
	})
}

func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := getToken(r.Header.Get(authorizationHeader))
		if err != nil {
			HandleError(w, http.StatusUnauthorized, err)
			return
		}

		flag, err := h.svc.CheckToken(token)
		if err != nil {
			if errors.Is(err, services.ErrTokenInvalid) {
				HandleError(w, http.StatusUnauthorized, err)
				return
			}
			HandleError(w, http.StatusInternalServerError, err)
			return
		}

		if !flag {
			HandleError(w, http.StatusUnauthorized, errUserIsNotExist)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getToken(authHead string) (string, error) {
	if authHead == "" {
		return "", errAuthorizationHeaderIsEmpty
	}

	headerParts := strings.Split(authHead, " ")
	if len(headerParts) != 2 || headerParts[0] != "Beaver" {
		return "", errHeaderIsNotRequiredMask
	}
	return headerParts[1], nil
}
