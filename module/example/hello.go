package main

import (
	"net/http"
)

// for plugins that dont contain a command to
// load, all lines marked with //OP are optional
// if your plugin adds a c ommand, please fille these out
func Name() string    { return `hello` } //OP
func Version() string { return `0.0.1` }
func Usage() string   { return `example` }                //OP
func Desc() string    { return `description of example` } //OP
func Init() {
	// initial setup for the plugin
	println("Hello from plugins")
}

func Handle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from plugins"))
}

func HandleWs(w http.ResponseWriter, r *http.Request) {
	// unused but you use this for your websockets, which are exposed at `PackageNamews`
}
