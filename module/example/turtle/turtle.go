package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"ultron/api"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func Name() string    { return `turtle` } //OP
func Version() string { return `0.1.0` }
func Desc() string    { return `The Base turtle control API` } //OP
func Usage() string {
	return `
/api/turtle
	GET: Returns data of all turtles
/api/turtle/<ID>
	TIP: use ID "debug" to see the structure of the json
	GET: Returns data of single turtle
	POST: Send command to turtle
		EX: JSON ["print('Hello from Ultron')"] will print to turtle display
/api/turtle/ws
	This is the websocket for turtles, please do not attempt to use
`
}

func Init(m *mux.Router) {
	//create api for /api/turtle with argument for id
	m.HandleFunc("/api/turtle/ws", HandleWs)
	m.HandleFunc("/api/turtle", Handle)
	m.HandleFunc("/api/turtle/{id}", Handle)
	m.HandleFunc("/api/turtle/{id}/{action}", Handle)
}

var Turtles []Turtle

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

func Handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idInt, _ := strconv.Atoi(id)
	action := vars["action"]

	if id == "ws" {
		HandleWs(w, r)
		return
	}
	if id == "usage" {
		w.Write([]byte(Usage()))
		return
	}

	// check if id is in Turtles
	var currentTurtle Turtle
	found := false
	pos := 0
	for p, t := range Turtles {
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
		Turtles = append(Turtles, currentTurtle)
	}

	// http://localhost:3300/api/turtle/1
	if r.Method == "GET" {
		// return turtle data on /api/turtle/{id}
		if id == "" {
			// if Turtles is empty
			if len(Turtles) == 0 {
				// returnError no Turtles found as json with status code 503
				api.ReturnError(w, http.StatusServiceUnavailable, "No Turtles found")
				return
			} else {
				//return all turtle data
				//api.ReturnData(w, Turtles)
				api.ReturnData(w, Turtles)
			}

		} else if id != "" {
			if !found {
				api.ReturnError(w, http.StatusServiceUnavailable, "Turtle has not been added yet")
				return
			}
			// make switch for action
			switch action {
			case "":
				// return turtle data
				api.ReturnData(w, currentTurtle)
			case "name":
				// return turtle name
				api.ReturnData(w, currentTurtle.Name)
			case "fuel":
				// return turtle fuel
				api.ReturnData(w, currentTurtle.Fuel)
			case "misc":
				// return turtle misc
				api.ReturnData(w, currentTurtle.MiscData)
			case "inventory":
				// return turtle inventory
				api.ReturnData(w, currentTurtle.Inventory)
			case "selectedSlot":
				// return turtle selected slot
				api.ReturnData(w, currentTurtle.SelectedSlot)
			case "pos":
				//return turtle position
				api.ReturnData(w, currentTurtle.Pos)
			case "cmdQueue":
				// return turtle cmdQueue
				api.ReturnData(w, currentTurtle.CmdQueue)
			case "cmdResult":
				// return turtle cmdResult
				api.ReturnData(w, currentTurtle.CmdResult)
				// print turtle cmdResult
				log.Println("[Turtle]", currentTurtle.Name, ":", currentTurtle.CmdResult)
			default:
				api.ReturnError(w, http.StatusBadRequest, "Invalid action: "+action)
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
		log.Println("[Command Queue]: [", Turtles[pos].ID, "]", currentTurtle.CmdQueue)
		// add currentTurtle.CmdQueue to Turtles[pos].CmdQueue
		Turtles[pos].CmdQueue = append(Turtles[pos].CmdQueue, currentTurtle.CmdQueue...)
	}
}

// handle turtle websocket
func HandleWs(w http.ResponseWriter, r *http.Request) {
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

		// find currentTurtle in Turtles
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
			Turtles = append(Turtles, currentTurtle)
		} else {
			for p, t := range Turtles {
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
			// add currentTurtle to Turtles
			Turtles = append(Turtles, currentTurtle)
			log.Println("[Turtle] Added new turtle:", currentTurtle.ID, ":", currentTurtle.Name)
		} else {
			// update currentTurtle in Turtles
			Turtles[pos] = currentTurtle
		}
		// check if currentTurtle.CmdResult is the same as Turtles[pos].CmdResult
		if len(currentTurtle.CmdResult) != len(Turtles[pos].CmdResult) {
			// log result
			log.Println("[Turtle]", currentTurtle.Name, ":", currentTurtle.CmdResult)
		}
		// if cmdQueue is not empty, send cmdQueue to client
		if len(Turtles[pos].CmdQueue) > 0 {
			// convert cmdQueue to json
			jsonCmdQueue, jsonErr := json.Marshal(Turtles[pos].CmdQueue)
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

			Turtles[pos].CmdQueue = []string{}
			currentTurtle.CmdQueue = []string{}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}
