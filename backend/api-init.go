package main

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func createApiServer() {
	// // create webserver on port 3300
	//go func() {
	r := mux.NewRouter()

	// handle /
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// return all api routes
		w.Write([]byte("Welcome to the Ultron API!"))
	})

	// Serve Turtle Files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(config.LuaFiles))))

	// handle /api/computer
	r.HandleFunc("/api/computer", handleComputerApi).Methods("GET", "POST")
	r.HandleFunc("/api/computer/{id}", handleComputerApi).Methods("GET", "POST")

	//create api for /api/turtle with argument for id
	r.HandleFunc("/api/turtle", handleTurtleApi)
	r.HandleFunc("/api/turtle/{id}", handleTurtleApi)
	r.HandleFunc("/api/turtle/{id}/{action}", handleTurtleApi)

	// todo: create api for /api/world
	//create api for /api/world
	r.HandleFunc("/api/world", handleWorldApi)

	// todo: make global api more than placeholder
	//handle global api on /api/v1
	r.HandleFunc("/api", handleGlobalApi)

	r.HandleFunc("/turtlews", turtleWs)
	r.HandleFunc("/pocketws", api.pocketWs)
	r.HandleFunc("/computerws", computerWs)

	// if page not found, return server error
	r.NotFoundHandler = http.HandlerFunc(handleServerError)

	// start webserver on config.Port
	port := strconv.Itoa(config.Port)
	http.ListenAndServe(":"+port, r)
}

// make function to handle server errors
func handleServerError(w http.ResponseWriter, r *http.Request) {
	returnError(w, http.StatusNotImplemented, "Server Error: Check for trailing / in url")
}

func returnError(w http.ResponseWriter, code int, message string) {
	w.Write([]byte("{ \"error\": { \"code\":" + strconv.Itoa(code) + ", \"message\": \"" + message + "\" } }"))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}
