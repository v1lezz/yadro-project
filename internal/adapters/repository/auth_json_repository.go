package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"yadro-project/internal/core/ports"
)

var errAuthFileIsNotExist = errors.New("auth file is not exist")

//type Users struct {
//
//}

type AuthJSONRepository struct {
	Users  map[string]string `json:"users"`
	Admins map[string]string `json:"admins"`
}

func NewAuthJSONRepository(filePath string) (*AuthJSONRepository, error) {
	isExist, err := FileIsExist(filePath)
	if err != nil {
		return nil, fmt.Errorf("error check exist file: %w", err)
	}

	if !isExist {
		return nil, errAuthFileIsNotExist
	}

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("error open file \"%s\": %w", filePath, err)
	}

	authRepo := &AuthJSONRepository{}

	if err = json.NewDecoder(file).Decode(authRepo); err != nil {
		return nil, fmt.Errorf("error decode json from \"%s\": %w", filePath, err)
	}

	return authRepo, nil
}

//func (r *AuthJSONRepository) CheckUser(request domain.LoginRequest) (bool, error) {
//	if password, ok := r.Users[request.Email]; ok && password == request.Password {
//		return true, nil
//	}
//
//	if password, ok := r.Admins[request.Email]; ok && password == request.Password {
//		return true, nil
//	}
//
//	return false, nil
//}

func (r *AuthJSONRepository) GetPasswordByEmail(email string) (string, error) {
	if password, ok := r.Users[email]; ok {
		return password, nil
	}

	if password, ok := r.Admins[email]; ok {
		return password, nil
	}

	return "", ports.ErrIsNotExist
}

func (r *AuthJSONRepository) CheckUserByEmail(email string) (bool, error) {
	if _, ok := r.Users[email]; ok {
		return true, nil
	}

	if _, ok := r.Admins[email]; ok {
		return true, nil
	}

	return false, nil
}

func (r *AuthJSONRepository) CheckAdminByEmail(email string) (bool, error) {
	if _, ok := r.Admins[email]; ok {
		return true, nil
	}

	return false, nil
}
