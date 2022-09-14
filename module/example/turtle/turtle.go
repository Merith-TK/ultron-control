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
	m.HandleFunc("/api/turtle/fs", HandleFs)
	m.HandleFunc("/api/turtle/fs/{file}", HandleFs)
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
	if id == "fs" {
		HandleFs(w, r)
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

func HandleFs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["file"]
	log.Println("[File] Serving file:", file)
	if file == "module.lua" {
		w.Write([]byte(`
local ultron = require("ultron")
assert(ultron)
		
if not fs.exists("/skyrtle.lua") then
	local localfile = fs.open("skyrtle.lua", "w")
	local dl = http.get("https://raw.githubusercontent.com/SkyTheCodeMaster/SkyDocs/main/src/main/misc/skyrtle.lua")
	if dl then
		localfile.write(dl.readAll())
	else
		print("[Error]: Unable to download " .. "skyrtle.lua")
	end
	localfile.close()
end

_G.skyrtle = require("skyrtle")
skyrtle.hijack()

ultron.config.api.ws = ultron.config.api.host:gsub("http", "ws") .. "/turtle/ws"
ultron.ws("open", ultron.config.api.ws)
ultron.debugPrint()

ultron.debugPrint("ApiDelay: " .. ultron.config.api.delay)
ultron.debugPrint("Websocket URL: " .. "\n" ..ultron.config.api.ws)
--ultron.debugPrint("Websocket Header: " .. textutils.serialize(wsHeader))

ultron.data = {
	name = "",
	id = 0,
	pos = {
		x = 0,
		y = 0,
		z = 0,
		r = 0,
		rname = "",
	},
	fuel = {
		current = 0,
		max = 0,
	},
	selectedSlot = 0,
	inventory = {},
	cmdResult = {},
	cmdQueue = {},
	miscData = {},
}

		
-- function to send turtle data to websocket
local function updateControl()
	ultron.data.id = os.getComputerID()
	local label = os.getComputerLabel()
	if label and not label == "" then
		ultron.data.name = label
	else
		os.setComputerLabel(tostring(ultron.data.id))
		ultron.data.name = tostring(ultron.data.id)
	end

	local x,y,z = skyrtle.getPosition()
	local r, rname = skyrtle.getFacing()
	ultron.data.pos.x = x
	ultron.data.pos.y = y
	ultron.data.pos.z = z
	ultron.data.pos.r = r
	ultron.data.pos.rname = rname
		
	ultron.data.fuel.current = turtle.getFuelLevel()
	ultron.data.fuel.max = turtle.getFuelLimit()
	
	ultron.data.selectedSlot = turtle.getSelectedSlot()
		
	for i = 1, 16 do
		local item = turtle.getItemDetail(i, true)
		if item then
			ultron.data.inventory[i] = item
		else
			ultron.data.inventory[i] = {}
		end
	end
	turtle.select(ultron.data.selectedSlot)

	local TurtleData =  textutils.serializeJSON(ultron.data)
	ultron.ws("send",TurtleData)
end
		
-- process cmdQueue as functionlocal function recieveOrders()
local function recieveOrders()
	ultron.data.cmdQueue = ultron.recieveOrders(ultron.data.cmdQueue)
end
local function processCmdQueue()
	local result = ultron.processCmdQueue(ultron.data.cmdQueue)
	if result then
		ultron.data.cmdResult = result
	end
end

		
local function waitForDelay()
	sleep(ultron.config.api.delay)
end
		
local function event_TurtleInventory()
	os.pullEvent("turtle_inventory")
end
local function apiLoop()
	while true do
		updateControl()
		parallel.waitForAny(waitForDelay,  event_TurtleInventory)
	end
end
		
local function main()
	parallel.waitForAll(apiLoop, recieveOrders, processCmdQueue)
end
		
-- load cmdQueue from file /cmdQueue.json
local file = fs.open("/cmdQueue.json", "r")
if file then
	local cmdQueue = textutils.unserializeJSON(file.readAll())
	file.close()
	if cmdQueue then
		ultron.data.cmdQueue = cmdQueue
	end
end
	
while true do
	local succ, err = pcall(main)
	if not succ then
		print("[Error] " .. err)
		break
	end
end
ultron.ws("close")
`))
	}

	log.Println("[FS]", file)
}
