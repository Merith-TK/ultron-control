// package main

// import (
// 	"fmt"
// 	"net/http"
// 	"ultron/api"

// 	"github.com/gorilla/mux"
// )

// // Name is the name of the module, this is used for the module name in the API
// // api/<name>
// // IF YOU ARE NOT THE ORIGINAL CREATOR OF THE MODULE, PLEASE DO NOT CHANGE THIS
// // AS IT WILL BREAK THINGS UNLESS YOU MODIFY ALL ASSOCIATED LUA FILES
// func Name() string    { return `example` } //OP
// func Version() string { return `0.0.1` }

// // Desc is the description of the module, this is used for the module description when loading
// func Desc() string { return `description of example` } //OP
// // Usage is used to tell the users/developers how this particiular plugin works
// func Usage() string {
// 	return `
// /api/example
// 	GET: Returns response from the example Hello Module
// /api/example/ws
// 	This is the websocket for the example Hello Module, please do not attempt to use
// `
// }
// func Init(m *mux.Router) {
// 	// initial setup for the plugin
// 	println("Hello from plugins")
// 	m.HandleFunc("/api/example/ws", HandleWs)
// 	m.HandleFunc("/api/example", Handle)
// 	m.HandleFunc("/api/example/files/{file}", HandleFiles)
// }

// func Handle(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == "GET" {
// 		api.ReturnData(w, map[string]string{"hello": "This is an example", "notice": "This is a notice"})

// 	} else if r.Method == "POST" {
// 		// print post data
// 		fmt.Println(r.Form)

// 	} else {
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 	}
// }

// func HandleWs(w http.ResponseWriter, r *http.Request) {
// 	// unused but you use this for your websockets, which are exposed at `PackageNamews`
// }

// func HandleFiles(w http.ResponseWriter, r *http.Request) {
// 	// return files from the module
// 	vars := mux.Vars(r)
// 	id := vars["file"]
// 	if id == "hello.lua" {
// 		w.Write([]byte("print('Hello from plugins')"))
// 	} else {
// 		w.WriteHeader(http.StatusNotFound)
// 	}
// }
