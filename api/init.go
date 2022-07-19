package api

import (
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

	// Serve Turtle Files
	r.PathPrefix("/api/static/").Handler(http.StripPrefix("/api/static/", http.FileServer(http.Dir(luaFiles))))

	// handle /api/computer
	r.HandleFunc("/api/computer", handleComputerApi).Methods("GET", "POST")
	r.HandleFunc("/api/computer/{id}", handleComputerApi).Methods("GET", "POST")


	// todo: create api for /api/world
	//create api for /api/world
	r.HandleFunc("/api/world", handleWorldApi)

	// todo: make global api more than placeholder
	//handle global api on /api/v1
	r.HandleFunc("/api", handleGlobalApi)

	r.HandleFunc("/api/turtlews", turtleWs)
	r.HandleFunc("/api/computerws", computerWs)

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
