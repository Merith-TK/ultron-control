package main

import (
	"github.com/merith-tk/ultron-control/api"
	"github.com/merith-tk/ultron-control/config"
)

func main() {
	config.ReadConfig(DefaultConfig)
	cfg := config.GetConfig()
	api.CreateApiServer(cfg.Host, cfg.Port, cfg.UltronData, StaticFiles, DocsFiles)
}
