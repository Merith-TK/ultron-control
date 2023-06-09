package api

import (
	"net/http"
)

// handle global api
func handleGlobalApi(w http.ResponseWriter, r *http.Request) {
	// list of global api routes
	if r.Method == "GET" {
		// return global api routes
		ReturnError(w, http.StatusNotImplemented, "Global API: GET not implemented")
	} else if r.Method == "POST" {
		// return global api routes
		ReturnError(w, http.StatusNotImplemented, "Global API: POST not implemented")
	}
}
