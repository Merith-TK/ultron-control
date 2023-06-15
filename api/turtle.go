package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// func Name() string    { return `turtle` } //OP
// func Version() string { return `0.1.0` }
// func Desc() string    { return `The Base turtle control API` } //OP
func TurtleUsage() string {
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
	Sight struct {
		Up    string `json:"up"`
		Down  string `json:"down"`
		Front string `json:"front"`
	} `json:"sight"`
	CmdResult string        `json:"cmdResult"`
	CmdQueue  []string      `json:"cmdQueue"`
	MiscData  []interface{} `json:"miscData"`
}

func TurtleHandle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idInt, _ := strconv.Atoi(id)
	if id == "debug" {
		idInt = -1
	}
	action, action2 := vars["action"], vars["action2"]
	if id == "ws" {
		TurtleHandleWs(w, r)
		return
	}
	if id == "usage" {
		w.Write([]byte(TurtleUsage()))
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
		//	find the debug turtle in Turtles
		for p, t := range Turtles {
			if t.Name == "debug" {
				currentTurtle = t
				found = true
				pos = p
				break
			}
		}
		if !found {
			// create new empty turtle for debugging api
			currentTurtle.ID = -1
			currentTurtle.Name = "debug"
			currentTurtle.CmdQueue = []string{}
			currentTurtle.CmdResult = ""
			currentTurtle.Inventory = []interface{}{}
			currentTurtle.MiscData = []interface{}{}
			currentTurtle.Pos.R = 0
			currentTurtle.Pos.Rname = "north"
			currentTurtle.Pos.X = 0
			currentTurtle.Pos.Y = 0
			currentTurtle.Pos.Z = 0
			currentTurtle.Fuel.Current = 0
			currentTurtle.Fuel.Max = 0
			currentTurtle.Sight.Up = "minecraft:air"
			currentTurtle.Sight.Down = "minecraft:air"
			currentTurtle.Sight.Front = "minecraft:air"
			currentTurtle.SelectedSlot = 0
			Turtles = append(Turtles, currentTurtle)
		}
	}

	// http://localhost:3300/api/turtle/1

	if r.Method == "GET" {
		// return turtle data on /api/turtle/{id}
		if id == "" {
			// if Turtles is empty
			if len(Turtles) == 0 {
				// returnError no Turtles found as json with status code 503
				ReturnError(w, http.StatusServiceUnavailable, "No Turtles found")
				return
			} else {
				//return all turtle data
				ReturnData(w, Turtles)
			}

		} else if id != "" {
			if !found {
				ReturnError(w, http.StatusServiceUnavailable, "Turtle has not been added yet")
				return
			}
			// make switch for action
			switch action {
			case "":
				// return turtle data
				ReturnData(w, currentTurtle)
			case "name":
				// return turtle name
				ReturnData(w, currentTurtle.Name)
			case "fuel":
				if action2 == "" {
					// return turtle fuel
					ReturnData(w, currentTurtle.Fuel)
				} else if action2 == "current" {
					// return turtle fuel current
					ReturnData(w, currentTurtle.Fuel.Current)
				} else if action2 == "max" {
					// return turtle fuel max
					ReturnData(w, currentTurtle.Fuel.Max)
				}
			case "misc":
				// return turtle misc
				ReturnData(w, currentTurtle.MiscData)
			case "inventory":
				// return turtle inventory
				ReturnData(w, currentTurtle.Inventory)
			case "selectedSlot":
				// return turtle selected slot
				ReturnData(w, currentTurtle.SelectedSlot)
			case "sight":
				// return turtle sight
				if action2 == "" {
					ReturnData(w, currentTurtle.Sight)
				} else if action2 == "up" {
					ReturnData(w, currentTurtle.Sight.Up)
				} else if action2 == "down" {
					ReturnData(w, currentTurtle.Sight.Down)
				} else if action2 == "front" {
					ReturnData(w, currentTurtle.Sight.Front)
				}
			case "pos":
				// return turtle pos
				if action2 == "" {
					ReturnData(w, currentTurtle.Pos)
				} else if action2 == "x" {
					ReturnData(w, currentTurtle.Pos.X)
				} else if action2 == "y" {
					ReturnData(w, currentTurtle.Pos.Y)
				} else if action2 == "z" {
					ReturnData(w, currentTurtle.Pos.Z)
				} else if action2 == "r" {
					ReturnData(w, currentTurtle.Pos.R)
				} else if action2 == "rname" {
					ReturnData(w, currentTurtle.Pos.Rname)
				}
			case "cmdQueue":
				ReturnData(w, currentTurtle.CmdQueue)
			case "cmdResult":
				// return turtle cmdResult
				ReturnData(w, currentTurtle.CmdResult)
				// print turtle cmdResult
				log.Println("[Turtle]", currentTurtle.Name, ":", currentTurtle.CmdResult)
			default:
				ReturnError(w, http.StatusBadRequest, "Invalid action: "+action)
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
func TurtleHandleWs(w http.ResponseWriter, r *http.Request) {
	// message should come in as json
	c, err := Upgrader.Upgrade(w, r, nil)
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
			currentTurtle.CmdResult = ""
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
		if currentTurtle.CmdResult != "" {
			// log result
			log.Println("[Turtle]", currentTurtle.Name, ":", currentTurtle.CmdResult)
		}
		// if cmdQueue is not empty, send cmdQueue to client
		if len(Turtles[pos].CmdQueue) > 0 {
			currentCmd := Turtles[pos].CmdQueue[0]
			err := c.WriteMessage(mt, []byte(currentCmd))
			if err != nil {
				log.Println("write:", err)
				break
			} else {
				Turtles[pos].CmdQueue = Turtles[pos].CmdQueue[1:]
			}
		}
		// import currentTurtle.Sight into Turtles[pos].Sight
		Turtles[pos].Sight = currentTurtle.Sight
		// Comment: Dont even know why I did this, or if it is even needed. but the code works as is so I am not touching it

	}
}
