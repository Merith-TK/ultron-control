package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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
	POST: Send command to turtle (synchronous by default)
		EX: JSON ["print('Hello from Ultron')"] will execute command and return result
		
	Execution Modes:
		- Default (sync): Real-time execution with immediate results
		- Async: Queue-based execution (use X-Execution-Mode: async header)
		
	Headers:
		- Content-Type: application/json or text/plain
		- X-Execution-Mode: sync (default) or async
		
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
		Up    interface{} `json:"up"`
		Down  interface{} `json:"down"`
		Front interface{} `json:"front"`
	} `json:"sight"`
	CmdResult interface{} `json:"cmdResult"`
	CmdQueue  []string    `json:"cmdQueue"`
	Misc      interface{} `json:"misc"`
	HeartBeat int         `json:"heartbeat"`
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
			currentTurtle.CmdResult = []interface{}{}
			currentTurtle.Inventory = []interface{}{}
			currentTurtle.Misc = []interface{}{}
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
				ReturnData(w, currentTurtle.Name)
			case "id":
				ReturnData(w, currentTurtle.ID)
			case "inventory":
				if action2 == "" {
					ReturnData(w, currentTurtle.Inventory)
				} else {
					act2, convertError := strconv.Atoi(action2)
					if convertError != nil || act2 < 0 || act2 > 16 {
						ReturnError(w, http.StatusBadRequest, "Invalid choice "+action2+", please use int 0-15 for slots 1-16")
					} else {
						ReturnData(w, currentTurtle.Inventory[act2])
					}
				}
			case "selectedSlot":
				ReturnData(w, currentTurtle.SelectedSlot)
			case "pos":
				switch action2 {
				case "":
					ReturnData(w, currentTurtle.Pos)
				case "x":
					ReturnData(w, currentTurtle.Pos.X)
				case "y":
					ReturnData(w, currentTurtle.Pos.Y)
				case "z":
					ReturnData(w, currentTurtle.Pos.Z)
				case "r":
					ReturnData(w, currentTurtle.Pos.R)
				case "rname":
					ReturnData(w, currentTurtle.Pos.Rname)
				default:
					ReturnError(w, http.StatusBadRequest, "Invalid choice "+action2+", please use x,y,z, r or rname")
				}
			case "fuel":
				switch action2 {
				case "":
					ReturnData(w, currentTurtle.Fuel)
				case "current":
					ReturnData(w, currentTurtle.Fuel.Current)
				case "max":
					ReturnData(w, currentTurtle.Fuel.Max)
				default:
					ReturnError(w, http.StatusBadRequest, "Invalid choice "+action2+", please use x,y,z, r or rname")
				}
			case "sight":
				switch action2 {
				case "":
					ReturnData(w, currentTurtle.Sight)
				case "down":
					ReturnData(w, currentTurtle.Sight.Down)
				case "front":
					ReturnData(w, currentTurtle.Sight.Front)
				case "up":
					ReturnData(w, currentTurtle.Sight.Up)
				}
			case "cmdResult":
				// TODO: Fetch data from Pos/Action2
				ReturnData(w, currentTurtle.CmdResult)
			case "cmdQueue":
				ReturnData(w, currentTurtle.CmdQueue)
			case "misc":
				// TODO: Fetch data from Pos/Action2
				ReturnData(w, currentTurtle.Misc)
			case "heartbeat":
				ReturnData(w, currentTurtle.HeartBeat)
			default:
				ReturnError(w, http.StatusBadRequest, "Invalid action: "+action)
			}
		}
	} else if r.Method == "POST" {
		// Check execution mode header - default to sync
		executionMode := r.Header.Get("X-Execution-Mode")
		if executionMode == "" {
			executionMode = "sync" // Default to synchronous execution
		}
		
		// Read command from request body
		contentType := r.Header.Get("Content-Type")
		postBody := r.Body
		var commands []string
		
		if contentType == "application/json" {
			if err := json.NewDecoder(postBody).Decode(&commands); err != nil {
				log.Println("[Error] Decoding json:", err)
				ReturnError(w, http.StatusBadRequest, "Error decoding json: "+err.Error())
				return
			}
		} else if contentType == "text/plain" {
			if postBody != nil {
				postbytes, err := ioutil.ReadAll(postBody)
				if err != nil {
					log.Println("[Error] Reading request body:", err)
					ReturnError(w, http.StatusBadRequest, "Error reading request body: "+err.Error())
					return
				}
				commands = append(commands, string(postbytes))
			}
		}
		
		if len(commands) == 0 {
			ReturnError(w, http.StatusBadRequest, "No commands provided")
			return
		}
		
		// Handle synchronous execution mode (now default)
		if executionMode == "sync" {
			if len(commands) > 1 {
				ReturnError(w, http.StatusBadRequest, "Synchronous mode only supports single command")
				return
			}
			
			command := commands[0]
			log.Println("[Sync Command (Default)]: [", idInt, "]", command)
			
			// Execute command synchronously with 30 second timeout
			result, err := executeCommandSync(idInt, command, 30*time.Second)
			if err != nil {
				if cmdErr, ok := err.(*CommandError); ok {
					ReturnError(w, http.StatusServiceUnavailable, cmdErr.Message)
				} else {
					ReturnError(w, http.StatusInternalServerError, err.Error())
				}
				return
			}
			
			// Return the result
			ReturnData(w, map[string]interface{}{
				"success": true,
				"result":  result,
				"mode":    "synchronous",
			})
			return
		}
		
		// Handle asynchronous execution mode (legacy behavior)
		if executionMode == "async" {
			if !found {
				ReturnError(w, http.StatusServiceUnavailable, "Turtle has not been added yet")
				return
			}
			
			// Get per-turtle command mutex to ensure FIFO execution with sync commands
			commandMutex := getTurtleCommandMutex(idInt)
			commandMutex.mutex.Lock()
			defer commandMutex.mutex.Unlock()
			
			log.Println("[Async Command Queue (Serialized)]: [", Turtles[pos].ID, "]", commands)
			// Add commands to turtle's queue (existing behavior)
			Turtles[pos].CmdQueue = append(Turtles[pos].CmdQueue, commands...)
			
			// Return success for async mode
			ReturnData(w, map[string]interface{}{
				"success": true,
				"message": "Commands queued successfully",
				"mode":    "asynchronous", 
				"count":   len(commands),
			})
			return
		}
		
		// Invalid execution mode
		ReturnError(w, http.StatusBadRequest, "Invalid execution mode. Use 'sync' (default) or 'async'")
	}
}

// handle turtle websocket
func TurtleHandleWs(w http.ResponseWriter, r *http.Request) {
	c, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	
	var turtleID int = -1 // Initialize to -1 to indicate unregistered
	log.Printf("[WebSocket] New connection established from %s", r.RemoteAddr)
	
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("[WebSocket] Read error (turtle %d): %v", turtleID, err)
			break
		}

		log.Printf("[WebSocket] Received message from turtle %d: %s", turtleID, string(message)[:min(100, len(message))])

		// Try to parse as structured message first
		var messageObj struct {
			Type     string      `json:"type"`
			Data     interface{} `json:"data"`
			TurtleID int         `json:"turtle_id"`
		}
		
		// Try to parse as command response
		var cmdResponse CommandResponse
		
		if json.Unmarshal(message, &cmdResponse) == nil && cmdResponse.RequestID != "" {
			// Handle command response
			log.Printf("[Command Response] RequestID: %s, Success: %v", cmdResponse.RequestID, cmdResponse.Success)
			handleCommandResponse(cmdResponse)
			continue
		}
		
		if json.Unmarshal(message, &messageObj) == nil && messageObj.Type != "" {
			// Handle structured messages
			log.Printf("[Structured Message] Type: %s from turtle %d", messageObj.Type, messageObj.TurtleID)
			switch messageObj.Type {
			case "command_result":
				log.Printf("[Structured Command Result] from turtle %d", messageObj.TurtleID)
				// This could be used for async command results in the future
			case "turtle_update":
				log.Printf("[Structured Turtle Update] from turtle %d", messageObj.TurtleID)
				// Handle structured turtle updates (include ID 0)
				if messageObj.TurtleID >= 0 {
					turtleID = messageObj.TurtleID
					registerTurtleConnection(turtleID, c)
					log.Printf("[Connection] Registered turtle %d via structured message", turtleID)
				}
			default:
				log.Printf("[Unknown Message Type] %s from turtle %d", messageObj.Type, messageObj.TurtleID)
			}
			continue
		}

		// Handle legacy turtle data (existing behavior)
		log.Printf("[Legacy Data] Processing turtle data")
		handleLegacyTurtleData(message, &turtleID, c)
	}
	
	// Clean up connection on disconnect (include ID 0)
	if turtleID >= 0 {
		log.Printf("[Disconnect] Turtle %d disconnected", turtleID)
		unregisterTurtleConnection(turtleID)
	} else {
		log.Printf("[Disconnect] Unregistered connection disconnected")
	}
}

// Helper function for safe string truncation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func handleLegacyTurtleData(message []byte, turtleID *int, c *websocket.Conn) {
	// Create empty CurrentTurtle
	var currentTurtle Turtle

	//decode json message onto currentTurtle
	jsonErr := json.Unmarshal(message, &currentTurtle)
	if jsonErr != nil {
		// ignore error that cannot be fixed currently
		if !strings.Contains(jsonErr.Error(), "Turtle.cmdQueue of type") {
			log.Println("[Error] Decoding TurtleJson:", jsonErr)
		}
	}

	// find currentTurtle in Turtles
	found := false
	pos := 0
	// if ID is -1, it is a debug turtle
	if currentTurtle.ID == -1 {
		// create new empty turtle for debugging api
		currentTurtle.ID = -1
		currentTurtle.Name = "debug"
		currentTurtle.CmdQueue = []string{}
		currentTurtle.CmdResult = []interface{}{}
		found = true
		Turtles = append(Turtles, currentTurtle)
	} else {
		for p, t := range Turtles {
			if t.ID == currentTurtle.ID {
				if len(currentTurtle.CmdQueue) == 0 {
					currentTurtle.CmdQueue = []string{}
				}
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
		pos = len(Turtles) - 1
		log.Println("[Turtle] Added new turtle:", currentTurtle.ID, ":", currentTurtle.Name)
	} else {
		// update currentTurtle in Turtles
		currentTurtle.CmdQueue = Turtles[pos].CmdQueue
		Turtles[pos] = currentTurtle
	}
	
	// Register/update turtle connection (include ID 0)
	if currentTurtle.ID >= 0 {
		*turtleID = currentTurtle.ID
		registerTurtleConnection(currentTurtle.ID, c)
	}
	
	// if cmdQueue is not empty, send cmdQueue to client
	if len(Turtles[pos].CmdQueue) > 0 {
		currentCmd := Turtles[pos].CmdQueue[0]
		err := c.WriteMessage(websocket.TextMessage, []byte(currentCmd))
		if err != nil {
			log.Println("write:", err)
		} else {
			Turtles[pos].CmdQueue = Turtles[pos].CmdQueue[1:]
		}
	}
	// import currentTurtle.Sight into Turtles[pos].Sight
	Turtles[pos].Sight = currentTurtle.Sight
	// Comment: Dont even know why I did this, or if it is even needed. but the code works as is so I am not touching it
}
