package main

import (
	"log"
	"os"

	toml "github.com/pelletier/go-toml"
)

type Config struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

var (
	config Config
	configfile = "config.toml"
	configtoml = `## Define Host for the API server
host = "localhost"
## Define Port for the API server
port = "3300"
`
)

//read config from config.toml
func readConfig() {
	// Check if file is readable, if not, make the file
	str, fileErr := os.ReadFile(configfile)
	if fileErr != nil {
		log.Fatalln("Config file not found, creating new one")
		createConfig()
	}
	err := toml.Unmarshal([]byte(str), &config)
	if err != nil {
		log.Fatalln(err)
	}
}

// create config.toml file and fill it with string
func createConfig() {
	//write string to file
	file, err := os.Create(configfile)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	file.WriteString(configtoml)
}