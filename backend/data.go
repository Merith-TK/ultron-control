package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var turtles []Turtle
type Turtle struct {
	Name         string        `json:"name"`
	ID           int           `json:"id"`
	Inventory    []interface{} `json:"inventory"`
	SelectedSlot int           `json:"selectedSlot"`
	Pos          struct {
		Y int `json:"y"`
		X int `json:"x"`
		R int `json:"r"`
		Z int `json:"z"`
	} `json:"pos"`
	Fuel int `json:"fuel"`
	MaxFuel int `json:"maxFuel"`
	CmdResult string `json:"cmdResult"`
	CmdQueue []string `json:"cmdQueue"`
}

var worldDataBlocks []WorldDataBlock
type WorldDataBlock struct {
		Block string `json:"block"`
		Pos   struct {
			X int `json:"x"`
			Y int `json:"y"`
			Z int `json:"z"`
		} `json:"pos"`
		Data []interface{} `json:"data"`
}

// create function to load data from files to memory
func loadData() {
	// if config.UltronData deoes not exist, create it
	if _, err := os.Stat(config.UltronData); os.IsNotExist(err) {
		log.Println("Ultron data directory does not exist, creating it")
		os.Mkdir(config.UltronData, 0777)
	}

	// load turtle data
	// turtleFile, err := ioutil.ReadFile(config.UltronData + "/turtle.json")
	// if err != nil {
	// 	log.Println("[TurtleData]: File does not exist")
	// }
	// json.Unmarshal(turtleFile, &turtles)

	// load world data
	worldFile, err := ioutil.ReadFile(config.UltronData + "/world.json")
	if err != nil {
		// state that file does not exist
		log.Println("[WorldData]: File does not exist")

	}
	json.Unmarshal(worldFile, &worldDataBlocks)
}

// save data to files
func saveData() {
	// save turtle data
	// turtleFile, err := json.Marshal(turtles)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// ioutil.WriteFile(config.UltronData + "/turtle.json", turtleFile, 0644)

	// save world data
	worldFile, err := json.Marshal(worldDataBlocks)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(config.UltronData + "/world.json", worldFile, 0644)
}