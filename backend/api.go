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


	//handle global api on /api/v1
	r.HandleFunc("/api/v1", handleGlobalApi)
	r.HandleFunc("/api/v1/", handleGlobalApi)

	//create api for /api/turtle with argument for id
	r.HandleFunc("/api/v1/turtle", handleTurtleApi)
	r.HandleFunc("/api/v1/turtle/", handleTurtleApi)
	r.HandleFunc("/api/v1/turtle/{id}", handleTurtleApi)
	r.HandleFunc("/api/v1/turtle/{id}/", handleTurtleApi)
	r.HandleFunc("/api/v1/turtle/{id}/{action}", handleTurtleApi)
	r.HandleFunc("/api/v1/turtle/{id}/{action}/{function}", handleTurtleApi)


	//create api for /api/world
	r.HandleFunc("/api/v1/world", handleWorldApi)
	r.HandleFunc("/api/v1/world/", handleWorldApi)


	// start webserver on port 3300
	http.ListenAndServe(":3300", r)
}

//handle global api
func handleGlobalApi(w http.ResponseWriter, r *http.Request) {
	// return list of available api
	w.Write([]byte("Available API: \n" +
		"/api/v1/turtle \n" +
		"/api/v1/turtle/{id} \n" +
		"/api/v1/turtle/{id}/{action} \n" +
		"/api/v1/turtle/{id}/{action}/{function} \n" +
		"/api/v1/world \n" +
		"/api/v1/world/{id} \n" +
		"/api/v1/world/{id}/{action} \n" +
		"/api/v1/world/{id}/{action}/{function} \n"))
}

//handle turtle api
func handleTurtleApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//parse data
		vars := mux.Vars(r)
		id := vars["id"]
		idInt,_ := strconv.Atoi(id)
		action := vars["action"]
		currentTurtle := turtleData[0]
		for _, t := range turtleData {
			if t.ID == idInt {
				currentTurtle = t
				break
			}
		}

		// return turtle data on /api/turtle/{id}
		if id == "" {
			//return all turtle data
			json.NewEncoder(w).Encode(turtleData)
		} else if id != "" && action != "" {
			// if action is pos, return turtle position and rotation
			if action == "pos" {
				json.NewEncoder(w).Encode(currentTurtle.Pos)
			}
		} else if id != "" && action == "" {
			// if no action is given, return turtle data
			json.NewEncoder(w).Encode(currentTurtle)
		}


	}
	/*
	if r.Method == "GET" {
		// parse data
		vars := mux.Vars(r)
		id := vars["id"]
		action := vars["action"]
		// check if id is int
		if idInt, err := strconv.Atoi(id); err == nil {
		if id != "" && action == "" {
			// iterate over apiData.Turtle and find turtle with id
			// if turtle is found, return turtle as json
			for _, t := range turtleData {
				if t.ID == idInt {
					json.NewEncoder(w).Encode(t)
					break
				}
			}
		} else if action != "" {

	} else if r.Method == "POST" {
		// do stuff
	}
	*/
}

//handle world api
func handleWorldApi(w http.ResponseWriter, r *http.Request) {
	world := worldData
	var resp []byte
	resp,_ = json.Marshal(world)
	w.Write(resp)

}