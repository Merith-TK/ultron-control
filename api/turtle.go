package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var turtles []Turtle

type Turtle struct {
	Name         string        `json:"name"`
	ID           int           `json:"id"`
	Inventory    []interface{} `json:"inventory"`
	SelectedSlot int           `json:"selectedSlot"`
	Pos          struct {
		Y     int    `json:"y"`
		X     int    `json:"x"`
		Z     int    `json:"z"`
		R     int    `json:"r"`
		Rname string `json:"rname"`
	} `json:"pos"`
	Fuel struct {
		Current int `json:"current"`
		Max     int `json:"max"`
	} `json:"fuel"`

	CmdResult []interface{} `json:"cmdResult"`
	CmdQueue  []string      `json:"cmdQueue"`
	MiscData  []interface{} `json:"miscData"`
}

func handleTurtleApi(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idInt, _ := strconv.Atoi(id)
	action := vars["action"]

	// check if id is in turtles
	var currentTurtle Turtle
	found := false
	pos := 0
	for p, t := range turtles {
		if t.ID == idInt {
			currentTurtle = t
			found = true
			pos = p
			break
		}
	}
	if id == "debug" {
		// create new empty turtle for debugging api
		currentTurtle.ID = -1
		currentTurtle.Name = "debug"
		currentTurtle.CmdQueue = []string{}
		currentTurtle.CmdResult = nil
		found = true
		turtles = append(turtles, currentTurtle)
	}

	// http://localhost:3300/api/turtle/1

	if r.Method == "GET" {
		// return turtle data on /api/turtle/{id}
		if id == "" {
			// if turtles is empty
			if len(turtles) == 0 {
				// returnError no turtles found as json with status code 503
				returnError(w, http.StatusServiceUnavailable, "No turtles found")
				return
			} else {
				//return all turtle data
				json.NewEncoder(w).Encode(turtles)
			}

		} else if id != "" {
			if !found {
				returnError(w, http.StatusServiceUnavailable, "Turtle has not been added yet")
				return
			}
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
			case "cmdResult":
				// return turtle cmdResult
				json.NewEncoder(w).Encode(currentTurtle.CmdResult)
				// print turtle cmdResult
				log.Println("[Turtle]", currentTurtle.Name, ":", currentTurtle.CmdResult)
			default:
				returnError(w, http.StatusBadRequest, "Invalid action: "+action)
			}
		}
	} else if r.Method == "POST" {
		// r.Body should be a json string
		// decode json string into currentTurtle.CmdQueue
		if err := json.NewDecoder(r.Body).Decode(&currentTurtle.CmdQueue); err != nil {
			log.Println("[Error] Decoding json:", err)
			w.Write([]byte("Error: Decoding json " + err.Error()))
			return
		}

		// log command queue to console
		log.Println("[Command Queue]: [", turtles[pos].ID, "]", currentTurtle.CmdQueue)
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
		// if ID is -1, it is a debug turtle
		if currentTurtle.ID == -1 {
			// create new empty turtle for debugging api
			currentTurtle.ID = -1
			currentTurtle.Name = "debug"
			currentTurtle.CmdQueue = []string{}
			currentTurtle.CmdResult = nil
			found = true
			turtles = append(turtles, currentTurtle)
		} else {
			for p, t := range turtles {
				if t.ID == currentTurtle.ID {
					currentTurtle.CmdQueue = t.CmdQueue
					t.CmdResult = currentTurtle.CmdResult
					found = true
					pos = p
					break
				}
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
		// check if currentTurtle.CmdResult is the same as turtles[pos].CmdResult
		if len(currentTurtle.CmdResult) != len(turtles[pos].CmdResult) {
			// log result
			log.Println("[Turtle]", currentTurtle.Name, ":", currentTurtle.CmdResult)
		}
		// if cmdQueue is not empty, send cmdQueue to client
		if len(turtles[pos].CmdQueue) > 0 {
			// convert cmdQueue to json
			jsonCmdQueue, jsonErr := json.Marshal(turtles[pos].CmdQueue)
			if jsonErr != nil {
				log.Println("[Error] Marshalling json:", jsonErr)
				// return error to client
				c.WriteMessage(mt, []byte("Error: Marshalling json"))
			}
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
	}
}
