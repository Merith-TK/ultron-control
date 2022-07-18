package main

import (
	"ultron/api"
	"ultron/config"
)

func main() {
	config.ReadConfig()
	configData := config.GetConfig()
	api.CreateApiServer(configData.Host, configData.Port, configData.LuaFiles, configData.ModuleDir)
}
