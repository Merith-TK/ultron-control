package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"ultron/module"
)

func CreateApiServer(domain string, port int, luaFiles string, dataDir string) {
	// // create webserver on port 3300
	r := mux.NewRouter()

	// handle /
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// return all api routes
		w.Write([]byte("Welcome to the Ultron API!"))
	})

	// load plugins
	r = module.LoadModules(r, dataDir+"/modules")
	r.HandleFunc("/api/modules", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(module.ModuleList)
	})

	// Serve Turtle Files
	r.PathPrefix("/api/static/").Handler(http.StripPrefix("/api/static/", http.FileServer(http.Dir(luaFiles))))

	// handle /api/computer
	r.HandleFunc("/api/computer/ws", computerWs)
	r.HandleFunc("/api/computer", handleComputerApi).Methods("GET", "POST")
	r.HandleFunc("/api/computer/{id}", handleComputerApi).Methods("GET", "POST")

	//handle global api on /api/v1
	r.HandleFunc("/api", handleGlobalApi)

	// if page not found, return server error
	r.NotFoundHandler = http.HandlerFunc(handleServerError)

	// start webserver on config.Port
	portstr := strconv.Itoa(port)
	http.ListenAndServe(domain+":"+portstr, r)
}

// make function to handle server errors
func handleServerError(w http.ResponseWriter, r *http.Request) {
	returnError(w, http.StatusNotImplemented, "Server Error: Check for trailing / in url, or verify against documentation of API")
}

func returnError(w http.ResponseWriter, code int, message string) {
	w.Write([]byte("{ \"error\": { \"code\":" + strconv.Itoa(code) + ", \"message\": \"" + message + "\" } }"))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}
