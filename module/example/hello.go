package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// for plugins that dont contain a command to
// load, all lines marked with //OP are optional
// if your plugin adds a c ommand, please fille these out
func Name() string    { return `hello` } //OP
func Version() string { return `0.0.1` }
func Usage() string   { return `example` }                //OP
func Desc() string    { return `description of example` } //OP 
func Init(m *mux.Router) {
	// initial setup for the plugin
	println("Hello from plugins")
	m.HandleFunc("/api/hello/ws", HandleWs)
	m.HandleFunc("/api/hello", Handle)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte("Hello from plugins"))
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