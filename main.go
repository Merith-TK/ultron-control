package main

import (
	"github.com/merith-tk/ultron-control/api"
	"github.com/merith-tk/ultron-control/config"
)

func main() {
	config.ReadConfig()
	configData := config.GetConfig()
	api.CreateApiServer(configData.Host, configData.Port, configData.LuaFiles, configData.UltronData)
}
