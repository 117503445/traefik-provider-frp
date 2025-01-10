package main

import (
	"117503445/traefik-provider-frp/pkg/frpsadmin"
	"117503445/traefik-provider-frp/pkg/frpsplugin"
	"117503445/traefik-provider-frp/pkg/state"

	"github.com/117503445/goutils"
	"github.com/alecthomas/kong"
	kongtoml "github.com/alecthomas/kong-toml"
	"github.com/rs/zerolog/log"
)

type FrpsPluginCfg struct {
	Port int `help:"frps plugin port" default:"8021"`
}

func main() {
	goutils.InitZeroLog()
	var cfg struct {
		FrpsAdmin  *frpsadmin.FrpsAdminCfg `embed:"" prefix:"frps-admin."`
		FrpsPlugin *FrpsPluginCfg          `embed:"" prefix:"frps-plugin."`
	}
	kong.Parse(&cfg, kong.Configuration(kongtoml.Loader, "./config.toml"))
	log.Info().Interface("cfg", cfg).Msg("main")

	stateUpdateCallback := func(state map[string]int) {
		log.Info().Interface("state", state).Msg("state update")
	}

	serviceDestState := state.NewServiceDestState(stateUpdateCallback)

	frpsAdminManager := frpsadmin.NewFrpsAdminManager(cfg.FrpsAdmin)
	frpsAdminManager.SetState(serviceDestState)

	err := frpsplugin.NewServer(serviceDestState).Serve(cfg.FrpsPlugin.Port)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to serve frps-plugin server")
	}
}
