package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)


type TurtleData []struct {
	Name string `json:"name"`
	ID  int `json:"id"`
	Pos struct {
		X int    `json:"x"`
		Y int    `json:"y"`
		Z int    `json:"z"`
		R string `json:"r"`
	} `json:"pos"`
	Inventory struct {
		SelectedSlot int `json:"selectedSlot"`
		Items        []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Count int    `json:"count"`
			Slot  int    `json:"slot"`
		} `json:"items"`
	} `json:"inventory,omitempty"`
}

type WorldData struct {
	Blocks []struct {
		Block string `json:"block"`
		Pos   struct {
			X int `json:"x"`
			Y int `json:"y"`
			Z int `json:"z"`
		} `json:"pos"`
		Data string `json:"data"`
	} `json:"blocks"`
}
var (
	// assign turtle data to variable
	turtleData TurtleData
	// assign world data to variable
	worldData WorldData
)

// create function to load data from files to memory
func loadData() {
	// load turtle data
	turtleFile, err := ioutil.ReadFile("./data/turtle.json")
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(turtleFile, &turtleData)

	// show turtle data
	log.Println("[TurtleData]:", turtleData)

	// load world data
	worldFile, err := ioutil.ReadFile("data/world.json")
	if err != nil {
		log.Fatal(err)
	}
	// unmarshal json
	json.Unmarshal(worldFile, &worldData)

	// show world data
	log.Println("[WorldData]: ", worldData)
}