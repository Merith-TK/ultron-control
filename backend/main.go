package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	readConfig()
	loadData()
	createApiServer()
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