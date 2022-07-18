package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

var (
	// Version is the version of the plugin
	Version = "0.0.1"
	// Name is the name of the plugin
	PackageName = "hello"
	// Description is a short description of the plugin
	Description = "A simple plugin that says hello"
)

func Init(r *mux.Router) *mux.Router {
	println("Hello from plugins")
	r.HandleFunc("/api/hello", handleHello)
	return r
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from plugins"))
}
