package services

import (
	"errors"
	"fmt"
	"time"
	"yadro-project/internal/core/domain"
	"yadro-project/internal/core/ports"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrBadCredentials = errors.New("email or password is incorrect")
	ErrTokenInvalid   = errors.New("token invalid")
)

type AuthService struct {
	repo         ports.AuthRepository
	tokenMaxTime uint
	jwtSecretKey []byte
}

func NewAuthService(repo ports.AuthRepository, tokenMaxTime uint) *AuthService {
	return &AuthService{
		repo:         repo,
		tokenMaxTime: tokenMaxTime,
		jwtSecretKey: []byte("sdgsfdghtrmogmfsdgskfdgnlsf"),
	}
}

func (svc *AuthService) Login(request domain.LoginRequest) (string, error) {
	flag, err := svc.repo.CheckUser(request)

	if err != nil {
		return "", fmt.Errorf("error check user: %w", err)
	}

	if !flag {
		return "", ErrBadCredentials
	}

	payload := jwt.MapClaims{
		"sub": request.Email,
		"exp": time.Now().Add(time.Minute * time.Duration(svc.tokenMaxTime)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := token.SignedString(svc.jwtSecretKey)
	if err != nil {
		return "", fmt.Errorf("error signing jwt secret key: %w", err)
	}

	return t, nil
}

func (svc *AuthService) CheckToken(sToken string) (bool, error) {
	t, err := jwt.Parse(sToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return nil, nil
	})

	if err != nil {
		return false, fmt.Errorf("error parse token: %w", err)
	}

	if !t.Valid {
		return false, ErrTokenInvalid
	}

	sub, err := t.Claims.GetSubject()
	if err != nil {
		return false, fmt.Errorf("error get sub from token: %w", err)
	}
	return svc.repo.CheckUserByEmail(sub)
}

func (svc *AuthService)