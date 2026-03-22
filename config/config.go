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
	UltronData string `toml:"ultron_data"`
}

var (
	config     Config
	configfile = "config.toml"
)

// make this function public
func GetConfig() Config {
	return config
}

// setup config
func SetupConfig(defaultContent []byte) {
	// read config for default values
	ReadConfig(defaultContent)

	// create flag for config.Port
	flag.IntVar(&config.Port, "port", 3300, "Port for the API server")
	// create flag for config.Host
	flag.StringVar(&config.Host, "host", "localhost", "Host for the API server")
	// create flag for config.UltronData
	flag.StringVar(&config.UltronData, "ultron_data", "../ultron_data", "Ultron Data Directory")

	// parse command line arguments
	flag.Parse()
}

// ReadConfig loads config.toml from disk. If the file does not exist,
// defaultContent is written to disk first (first-run bootstrap).
func ReadConfig(defaultContent []byte) {
	str, err := os.ReadFile(configfile)
	if err != nil {
		log.Println("[Config] No config file found, writing default")
		if werr := os.WriteFile(configfile, defaultContent, 0644); werr != nil {
			log.Fatal("[Config] Cannot write default config:", werr)
		}
		str = defaultContent
	}
	if err := toml.Unmarshal(str, &config); err != nil {
		log.Fatalln(err)
	}
	log.Println("[Config]", config)

	if _, err := os.Stat(config.UltronData); os.IsNotExist(err) {
		log.Println("[Config] Creating data directory:", config.UltronData)
		os.Mkdir(config.UltronData, 0755)
	}
}
