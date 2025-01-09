package main

import (
	"117503445/traefik-provider-frp/pkg/frpsadmin"

	"github.com/117503445/goutils"
	"github.com/alecthomas/kong"
	kongtoml "github.com/alecthomas/kong-toml"
	"github.com/rs/zerolog/log"
)

func main() {
	goutils.InitZeroLog()
	var cfg struct {
		FrpsAdmin *frpsadmin.FrpsAdminCfg `embed:"" prefix:"frps-admin."`
	}
	kong.Parse(&cfg, kong.Configuration(kongtoml.Loader, "./config.toml"))
	log.Info().Interface("cfg", cfg).Msg("main")

	frpsadmin.GetRoutes(cfg.FrpsAdmin)
}
