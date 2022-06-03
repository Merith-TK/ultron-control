package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

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
