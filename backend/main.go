package main

import (
	"backend/api"
	"backend/config"
)

func main() {
	config.ReadConfig()
	configData := config.GetConfig()
	api.CreateApiServer(configData.Host, configData.Port, configData.LuaFiles)
}
