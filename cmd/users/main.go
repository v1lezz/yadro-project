package main

import (
	"encoding/json"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	p1, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.MinCost)
	p2, _ := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.MinCost)
	p3, _ := bcrypt.GenerateFromPassword([]byte("87654321"), bcrypt.MinCost)
	b := map[string]interface{}{
		"admins": map[string]interface{}{
			"admin@amdin.com": string(p1),
		},
		"users": map[string]interface{}{
			"simple@simple.com":  string(p2),
			"simple2@simple.com": string(p3),
		},
	}
	f, err := os.Create("users.json")
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(f).Encode(b); err != nil {
		log.Fatal(err)
	}
}
