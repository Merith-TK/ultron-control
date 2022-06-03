package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)


func handleTurtleApi(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idInt,_ := strconv.Atoi(id)
	action := vars["action"]
	
	// check if id is in turtles
	var currentTurtle Turtle
	//found := false
	pos := 0
	for p, t := range turtles {
		if t.ID == idInt {
			currentTurtle = t
			//found = true
			pos = p
			break
		}
	}
	// if !found {
	// 	// if not found, return error
	// 	//w.Write([]byte("Error: Turtle with id " + id + " not found"))
	// 	//return
	// }

	
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
				json.NewEncoder(w).Encode(currentTurtle.Name)
			case "fuel":
				// return turtle fuel
				json.NewEncoder(w).Encode(currentTurtle.Fuel)
			case "misc":
				// return turtle misc
				json.NewEncoder(w).Encode(currentTurtle.MiscData)
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
		// r.Body should be a json string
		// decode json string into currentTurtle.CmdQueue
		if err := json.NewDecoder(r.Body).Decode(&currentTurtle.CmdQueue); err != nil {
			log.Println("[Error] Decoding json:", err)
			return
		}

		// add currentTurtle.CmdQueue to turtles[pos].CmdQueue
		turtles[pos].CmdQueue = append(turtles[pos].CmdQueue, currentTurtle.CmdQueue...)
	}
}

// handle turtle websocket
func turtleWs(w http.ResponseWriter, r *http.Request) {
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

		// find currentTurtle in turtles
		found := false
		pos := 0
		for p, t := range turtles {
			if t.ID == currentTurtle.ID {
				currentTurtle.CmdQueue = t.CmdQueue
				found = true
				pos = p
				break
			}
		}
		if !found {
			// add currentTurtle to turtles
			turtles = append(turtles, currentTurtle)
			log.Println("[Turtle] Added new turtle:", currentTurtle.ID, ":", currentTurtle.Name)
		} else {
			// update currentTurtle in turtles
			turtles[pos] = currentTurtle
		}
		if currentTurtle.CmdResult != "" {
			// log result
			log.Println("[Turtle]", currentTurtle.Name, ":", currentTurtle.CmdResult)
			// clear result
			currentTurtle.CmdResult = ""
			turtles[pos] = currentTurtle
		}
		// if cmdQueue is not empty, send cmdQueue to client
		if len(turtles[pos].CmdQueue) > 0 {
			// convert cmdQueue to json
			jsonCmdQueue, _ := json.Marshal(turtles[pos].CmdQueue)
			// send jsonCmdQueue to client and wait for response
			err := c.WriteMessage(mt, jsonCmdQueue)
			if err != nil {
				log.Println("write:", err)
				break
			}
			
			// clear cmdQueue
			turtles[pos].CmdQueue = []string{}
			currentTurtle.CmdQueue = []string{}
		}
		saveData()
	}
}
