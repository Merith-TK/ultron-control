package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var computers []Computer

type Computer struct {
	Name      string        `json:"name"`
	ID        int           `json:"id"`
	CmdResult []interface{} `json:"cmdResult"`
	CmdQueue  []string      `json:"cmdQueue"`
	MiscData  []interface{} `json:"miscData"`
}

// handle computer api
func handleComputerApi(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]
	idInt, _ := strconv.Atoi(id)
	//action := vars["action"]

	// check if id is in computers
	var currentComputer Computer
	//found := false
	pos := 0
	for p, t := range computers {
		if t.ID == idInt {
			currentComputer = t
			//found = true
			pos = p
			break
		}
	}

	// return computer data if request is GET
	if r.Method == "GET" {
		json.NewEncoder(w).Encode(computers)
	} else if r.Method == "POST" {
		// r.Body should be a json string
		// decode json string into currentComputer.CmdQueue
		if err := json.NewDecoder(r.Body).Decode(&currentComputer.CmdQueue); err != nil {
			log.Println("[Error] Decoding json:", err)
			return
		}

		// add currentComputer.CmdQueue to computers[pos].CmdQueue
		computers[pos].CmdQueue = append(computers[pos].CmdQueue, currentComputer.CmdQueue...)
	}
}

// handle computer websocket
func computerWs(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		// create empty CurrentComputer
		var currentComputer Computer

		//decode json message onto currentComputer
		json.Unmarshal(message, &currentComputer)

		// find currentComputer in computers
		found := false
		pos := 0
		for p, t := range computers {
			if t.ID == currentComputer.ID {
				currentComputer.CmdQueue = t.CmdQueue
				found = true
				pos = p
				break
			}
		}
		if !found {
			// add currentComputer to computers
			computers = append(computers, currentComputer)
			log.Println("[Computer] Added new computer:", currentComputer.ID, ":", currentComputer.Name)
		} else {
			// update currentComputer in computers
			computers[pos] = currentComputer
		}

		// check if cmdResult is empty
		if len(currentComputer.CmdResult) == 0 {
			// log result
			//log.Println("[Computer]", currentComputer.Name, ":", currentComputer.CmdResult)

			// clear currentComputer.CmdResult
			currentComputer.CmdResult = []interface{}{}
			computers[pos] = currentComputer
		}
		// if cmdQueue is not empty, send cmdQueue to client
		if len(computers[pos].CmdQueue) > 0 {
			// convert cmdQueue to json
			jsonCmdQueue, _ := json.Marshal(computers[pos].CmdQueue)
			// send jsonCmdQueue to client and wait for response
			err := c.WriteMessage(mt, jsonCmdQueue)
			if err != nil {
				log.Println("write:", err)
				break
			}

			// clear cmdQueue
			computers[pos].CmdQueue = []string{}
			currentComputer.CmdQueue = []string{}
		}
	}
}
