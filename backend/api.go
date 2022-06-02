package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"

	"github.com/gorilla/mux"
) 



func createApiServer() {
	// // create webserver on port 3300
	//go func() {
	r := mux.NewRouter()

	// Serve Turtle Files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(config.LuaFiles))))

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

	r.HandleFunc("/turtlews", handleWs)

	// start webserver on config.Port
	port := strconv.Itoa(config.Port)
	http.ListenAndServe(":"+port, r)
}

var upgrader = websocket.Upgrader{}
// handle websocket
func handleWs(w http.ResponseWriter, r *http.Request) {
	// message should come in as json
	
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		// DOC: unused value is header from client
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		// create empty CurrentTurtle
		var currentTurtle Turtle
		
		//decode json message onto currentTurtle
		json.Unmarshal(message, &currentTurtle)

		// check if turtle is already in list
		found := false
		pos := 0
		for p, t := range turtles {
			if t.ID == currentTurtle.ID {
				found = true
				turtles[p] = currentTurtle
				pos = p
				break
			}
		}
		if !found {
			turtles = append(turtles, currentTurtle)
		}
	
		// print recieved message from current turtle
		//log.Println("recieved:", currentTurtle.Name)
		
		
		
		// convert turtles[pos].cmdQueue to json
		jsonCmdQueue, err := json.Marshal(turtles[pos].CmdQueue)
		if err != nil {
			log.Println("marshal:", err)
			break
		}


		// send cmdQueue to client
		//log.Println("sending:", turtles[pos].CmdQueue)
		err = c.WriteMessage(mt, []byte(jsonCmdQueue))
		if err != nil {
			log.Println("write:", err)
			break
		}

		log.Println("[WebSocket]", turtles[pos].Name, ":", turtles[pos].CmdQueue)




		// clear cmdQueue
		//turtles[pos].CmdQueue = nil
		// save data
		saveData()
	}
}





//handle GET turtle api
func handleTurtleApi(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idInt,_ := strconv.Atoi(id)
	action := vars["action"]
	found := false
	pos := 0
	// create empty CurrentTurtle
	var currentTurtle Turtle
	
	for p, t := range turtles {
		if t.ID == idInt {
			currentTurtle = t
			found = true
			pos = p
			break
		}
	}
	if !found {
		currentTurtle.ID = idInt
		turtles = append(turtles, currentTurtle)
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
			case "cmdQueue":
				// return turtle cmdQueue
				json.NewEncoder(w).Encode(currentTurtle.CmdQueue)
			default:
				w.Write([]byte("Error: Action not found"))
			}
		}
	} else if r.Method == "POST" {
		// print r.Body
		if err := json.NewDecoder(r.Body).Decode(&currentTurtle.CmdQueue); err != nil {
			log.Println(err)
		} else {
			// print currentTurtle.CmdQueue
			//log.Println("Added To Que ", currentTurtle.CmdQueue)
			// append currentTurtle.CmdQueue to turtles[pos].CmdQueue
			turtles[pos].CmdQueue = append(turtles[pos].CmdQueue, currentTurtle.CmdQueue...)
			// print turtles[pos].CmdQueue
			//log.Println("Current Que ", turtles[arraypos].CmdQueue)
			

			// state currentTurtle.id
			log.Println("[Post]", turtles[pos].Name, ":", turtles[pos].CmdQueue)
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