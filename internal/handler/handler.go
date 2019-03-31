package handler

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func internalServerError(w http.ResponseWriter, err error) {
	log.Printf("error handling request: %v", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func int64Param(r *http.Request, name string) int64 {
	params := mux.Vars(r)
	value, _ := strconv.ParseInt(params[name], 10, 64)
	return value
}
