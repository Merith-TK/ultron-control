package main

import (
	"log"
	"os"

	toml "github.com/pelletier/go-toml"
)

type Config struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
	TurtleLua string `toml:"turtle_lua"`
	UltronData string `toml:"ultron_data"`
}

var (
	config Config
	configfile = "config.toml"
	configtoml = `## Define Host for the API server
host = "localhost"
## Define Port for the API server
port = 3300

## Location for Turtle lua files
turtle_lua = "../static"

## Ultron Data Directory
ultron_data = "../ultron_data"
`
)

//read config from config.toml
func readConfig() {
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
	readConfig()
}