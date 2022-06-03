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
		Z int `json:"z"`
		R int `json:"r"`
		Rname string `json:"rname"`
	} `json:"pos"`
	Fuel struct {
		Current int `json:"current"`
		Max     int `json:"max"`
	} `json:"fuel"`
	
	CmdResult string `json:"cmdResult"`
	CmdQueue []string `json:"cmdQueue"`
	MiscData []interface{} `json:"miscData"`
}

var pockets []Pocket
type Pocket struct {
	Name string `json:"name"`
	ID int `json:"id"`
	Pos struct {
		Y int `json:"y"`
		X int `json:"x"`
		Z int `json:"z"`
		R int `json:"r"`
		Rname string `json:"rname"`
	} `json:"pos"`
	MiscData []interface{} `json:"miscData"`
}

var computers []Computer
type Computer struct {
	Name string `json:"name"`
	ID int `json:"id"`
	CmdResult string `json:"cmdResult"`
	CmdQueue []string `json:"cmdQueue"`
	MiscData []interface{} `json:"miscData"`
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
	// save world data
	worldFile, err := json.Marshal(worldDataBlocks)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(config.UltronData + "/world.json", worldFile, 0644)
}