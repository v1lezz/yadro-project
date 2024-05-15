package handler

import (
	"fmt"
	"net/http"
)

func HandleError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf("error: %s", err)))
}
