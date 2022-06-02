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

	// handle / 
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// return all api routes
		w.Write([]byte("Welcome to the Ultron API!"))
	})

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

	r.HandleFunc("/turtlews", turtleWs)
	r.HandleFunc("/pocketws", pocketWs)

	// start webserver on config.Port
	port := strconv.Itoa(config.Port)
	http.ListenAndServe(":"+port, r)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}


// handle pocket websocket
func pocketWs(w http.ResponseWriter, r *http.Request) {
	// upgrade to websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// wait for request
	for {
		// read message
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// if msg is not empty
		if len(msg) > 0 {
			// check if message requesting turtle data
			if string(msg) == "turtle" {
				// convert turtles to json
				jsonTurtles, _ := json.Marshal(turtles)
				// send turtle data
				err = ws.WriteMessage(websocket.TextMessage, []byte(jsonTurtles))	
				if err != nil {
					log.Println(err)
					break
				} else {
					log.Println("[Pocket] Sent turtle data")
					log.Println("[Pocket]", turtles)
				}
			}
		}
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


//handle GET turtle api
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