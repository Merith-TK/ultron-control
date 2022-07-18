package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

var (
	// Version is the version of the plugin
	Version = "0.0.1"
	// Name is the name of the plugin
	PackageName = "hello"
	// Description is a short description of the plugin
	Description = "A simple plugin that says hello"
)

func Init(r *mux.Router, w http.ResponseWriter) *mux.Router {
	println("Hello from plugins")
	r.HandleFunc("/api/hello", Handle)
	return r
}

func Handle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from plugins"))
}
