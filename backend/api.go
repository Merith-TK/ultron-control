package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)


func createApiServer() {
	// create webserver on port 3300
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//serve files from ./public
		http.ServeFile(w, r, "./public/index.html")
	})


	//create api for /api/turtle with argument for id
	r.HandleFunc("/api/v1/turtle", handleTurtleApi)
	r.HandleFunc("/api/v1/turtle/{id}", handleTurtleApi)
	r.HandleFunc("/api/v1/turtle/{id}/{action}", handleTurtleApi).Methods("GET")

	// todo: create api for /api/world
	//create api for /api/world
	r.HandleFunc("/api/v1/world", handleWorldApi)
	r.HandleFunc("/api/v1/world/", handleWorldApi)
	
	// todo: make global api more than placeholder
	//handle global api on /api/v1
	r.HandleFunc("/api/v1", handleGlobalApi)
	r.HandleFunc("/api/v1/", handleGlobalApi)


	//convert config.Port from int to string
	port := strconv.Itoa(config.Port)
	
	// start webserver on port 3300
	http.ListenAndServe(":"+port, r)
}

//handle turtle api
func handleTurtleApi(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idInt,_ := strconv.Atoi(id)
	action := vars["action"]
	currentTurtle := turtleData[0]
	arrayPos := 0
	for p, t := range turtleData {
		if t.ID == idInt {
			currentTurtle = t
			arrayPos = p
			break
		}
	}

	if r.Method == "GET" {
		// return turtle data on /api/turtle/{id}
		if id == "" {
			//return all turtle data
			json.NewEncoder(w).Encode(turtleData)
		} else if id != "" {
			// make switch for action
			switch action {
			case "":
				//return turtle data
				json.NewEncoder(w).Encode(currentTurtle)
			case "pos":
				//return turtle position
				json.NewEncoder(w).Encode(currentTurtle.Pos)
			default:
				w.Write([]byte("Error: Action not found"))
			}
		}
	} else if r.Method == "POST" {
		// check if id is empty
		if id == "" {
			// return error
			w.Write([]byte("Error: ID not found"))
		} else if id != "" {
			// check if currentTurtle.ID == id
			if currentTurtle.ID == idInt {
				// send POST data to currentTurtle
				json.NewDecoder(r.Body).Decode(&currentTurtle)
				// save POST data to turtleData
				turtleData[arrayPos] = currentTurtle
			}
			//save data
				saveData()
		} else {
			// return error
			w.Write([]byte("Error: ID Missmatch"))
			//TODO: handle for non-existant turtle
		}
	}
}


//handle world api
func handleWorldApi(w http.ResponseWriter, r *http.Request) {
	world := worldData
	var resp []byte
	resp,_ = json.Marshal(world)
	w.Write(resp)

}

//handle global api
func handleGlobalApi(w http.ResponseWriter, r *http.Request) {
	// return list of available api
	w.Write([]byte("Available API: \n" +
		"Not Filled in by design for now"))
}