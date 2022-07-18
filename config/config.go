package config

import (
	"flag"
	"log"
	"os"

	toml "github.com/pelletier/go-toml"
)

type Config struct {
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	LuaFiles   string `toml:"luafiles"`
	UltronData string `toml:"ultron_data"`
	ModuleDir  string `toml:"module_dir"`
}

var (
	config     Config
	configfile = "config.toml"
	configtoml = `## Define Host for the API server
	host = ""
	## Define Port for the API server
	port = 3300
	
	## Location for Turtle lua files
	luafiles = "static"
	
	## Ultron Data Directory
	ultron_data = "workdir/ultron_data"
	
	module_dir = "workdir/modules"	
`
)

// make this function public
func GetConfig() Config {
	return config
}

// setup config
func SetupConfig() {
	// read config for default values
	ReadConfig()

	// create flag for config.Port
	flag.IntVar(&config.Port, "port", 3300, "Port for the API server")
	// create flag for config.Host
	flag.StringVar(&config.Host, "host", "localhost", "Host for the API server")
	// create flag for config.LuaFiles
	flag.StringVar(&config.LuaFiles, "luafiles", "../static", "Location for Turtle lua files")
	// create flag for config.UltronData
	flag.StringVar(&config.UltronData, "ultron_data", "../ultron_data", "Ultron Data Directory")

	// parse command line arguments
	flag.Parse()
}

//read config from config.toml
func ReadConfig() {
	// Check if file is readable, if not, make the file
	str, fileErr := os.ReadFile(configfile)
	if fileErr != nil {
		// log error that file could not be read
		log.Println("Config file could not be read, \nCreating new config file")
		// create config.toml file
		createConfig()
	}
	err := toml.Unmarshal([]byte(str), &config)
	if err != nil {
		log.Fatalln(err)
	}
	// print config
	log.Println("Config:", config)

	// if config.UltronData deoes not exist, create it
	if _, err := os.Stat(config.UltronData); os.IsNotExist(err) {
		log.Println("Ultron data directory does not exist, creating it")
		os.Mkdir(config.UltronData, 0777)
	}
}

// create config.toml file and fill it with string
func createConfig() {
	// create config.toml file
	file, err := os.Create(configfile)
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	// write configtoml to file
	_, err = file.WriteString(configtoml)
	if err != nil {
		log.Fatal("Cannot write to file", err)
	}
	// read config.toml file
	ReadConfig()
}
