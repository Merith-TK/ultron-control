package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"ultron/config"
)

var WorldDataBlocks []WorldDataBlock

var configData = config.GetConfig()

type WorldDataBlock struct {
	Block string `json:"block"`
	Pos   struct {
		X int `json:"x"`
		Y int `json:"y"`
		Z int `json:"z"`
	} `json:"pos"`
	Data []interface{} `json:"data"`
}

//handle world api
func handleWorldApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// return world data
		json.NewEncoder(w).Encode(WorldDataBlocks)
	} else if r.Method == "POST" {
		var data WorldDataBlock
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.Write([]byte("Error: " + err.Error()))
		} else {
			WorldDataBlocks = append(WorldDataBlocks, data)
			saveWorldData()
		}
	}
}

// save data to files
func saveWorldData() {
	// save world data
	worldFile, err := json.Marshal(WorldDataBlocks)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(configData.UltronData+"/world.json", worldFile, 0644)
}

// create function to load data from files to memory
func LoadData() {
	// load world data
	worldFile, err := ioutil.ReadFile(configData.UltronData + "/world.json")
	if err != nil {
		// state that file does not exist
		log.Println("[WorldData]: File does not exist")

	}
	json.Unmarshal(worldFile, &WorldDataBlocks)
}
