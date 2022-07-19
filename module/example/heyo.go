package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// for plugins that dont contain a command to
// load, all lines marked with //OP are optional
// if your plugin adds a c ommand, please fille these out
func Name() string    { return `heyo` } //OP 
func Version() string { return `0.0.1` }
func Desc() string    { return `description of example` } //OP 
func Usage() string   { return `usage of example` }

func Exposes() []string {
	var exposes []string
	exposes = append(exposes, "api/heyo")
	exposes = append(exposes, "api/heyo/ws")
	return exposes 
}
func Init(m *mux.Router) { 
	// initial setup for the plugin
	println("heyo from plugins")
	m.HandleFunc("/api/heyo/ws", HandleWs)
	m.HandleFunc("/api/heyo", Handle)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte("heyo from plugins"))
	} else if r.Method == "POST" {
		// print post data
		fmt.Println(r.Form)

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func HandleWs(w http.ResponseWriter, r *http.Request) {
	// unused but you use this for your websockets, which are exposed at `PackageNamews`
}

func HandleFiles(w http.ResponseWriter, r *http.Request) {
	// unused but you use this for your files, which are exposed at `PackageNamefiles`
}