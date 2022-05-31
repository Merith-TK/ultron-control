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
	// Serve Turtle Files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(config.TurtleLua))))

	//create api for /api/turtle with argument for id
	r.HandleFunc("/api/v1/turtle", handleTurtleApi)
	r.HandleFunc("/api/v1/turtle/{id}", handleTurtleApi)
	r.HandleFunc("/api/v1/turtle/{id}/{action}", handleTurtleApi).Methods("GET")

	// todo: create api for /api/world
	//create api for /api/world
	r.HandleFunc("/api/v1/world/", handleWorldApi)
	
	// todo: make global api more than placeholder
	//handle global api on /api/v1
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
	arrayPos := 0
	found := false
	// create empty CurrentTurtle
	var currentTurtle Turtle
	
	for p, t := range turtles {
		if t.ID == idInt {
			currentTurtle = t
			arrayPos = p
			found = true
			break
		}
	}
	if !found {
		// create new turtle with empty data
		currentTurtle.ID = idInt
		turtles = append(turtles, currentTurtle)
		arrayPos = len(turtles) - 1
	}
	
	if r.Method == "GET" {
		// return turtle data on /api/turtle/{id}
		if id == "" {
			//return all turtle data
			json.NewEncoder(w).Encode(turtles)
		} else if id != "" {
			// make switch for action
			switch action {
			case "":
				// return turtle data
				json.NewEncoder(w).Encode(currentTurtle)
			case "name":
				// return turtle name
				w.Write([]byte(currentTurtle.Name))
			case "inventory":
				// return turtle inventory
				json.NewEncoder(w).Encode(currentTurtle.Inventory)
			case "selectedSlot":
				// return turtle selected slot
				json.NewEncoder(w).Encode(currentTurtle.SelectedSlot)
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
			
			// send POST data to currentTurtle
			if err := json.NewDecoder(r.Body).Decode(&currentTurtle) ; err != nil {
				w.Write([]byte("Error: " + err.Error()))
			} else if currentTurtle.ID == idInt {
				// save POST data to turtles
				turtles[arrayPos] = currentTurtle 
			} else {
				// check if id is already in use
				for _, t := range turtles {
					if t.ID == idInt {
						w.Write([]byte("Error: ID already in use"))
						break
					}
				}
			}
			saveData()
		}

	}
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