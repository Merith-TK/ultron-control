package main

import (
	"encoding/json"
	"net/http"
)

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
