package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"

	"github.com/gorilla/mux"
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
	r.HandleFunc("/api/world/", handleWorldApi)
	
	// todo: make global api more than placeholder
	//handle global api on /api/v1
	r.HandleFunc("/api/", handleGlobalApi)

	r.HandleFunc("/turtlews", turtleWs)
	r.HandleFunc("/pocketws", pocketWs)
	r.HandleFunc("/computerws", computerWs)

	// start webserver on config.Port
	port := strconv.Itoa(config.Port)
	http.ListenAndServe(":"+port, r)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}


//handle world api
func handleWorldApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// return world data
		json.NewEncoder(w).Encode(worldDataBlocks)
	} else if r.Method == "POST" {
		var data WorldDataBlock
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.Write([]byte("Error: " + err.Error()))
		} else {
			worldDataBlocks = append(worldDataBlocks, data)
			saveData()
		}
	}
}

//handle global api
func handleGlobalApi(w http.ResponseWriter, r *http.Request) {
	// return list of available api
	w.Write([]byte("Available API: \n" +
		"Not Filled in by design for now"))
}