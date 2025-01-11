package main

import (
	"117503445/traefik-provider-frp/pkg/frps/admin"
	"117503445/traefik-provider-frp/pkg/frps/plugin"
	"117503445/traefik-provider-frp/pkg/traefik/writer"

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
		FrpsAdmin     *admin.FrpsAdminCfg      `embed:"" prefix:"frps-admin."`
		FrpsPlugin    *FrpsPluginCfg           `embed:"" prefix:"frps-plugin."`
		TraefikWriter *writer.TraefikWriterCfg `embed:"" prefix:"traefik-writer."`
	}
	kong.Parse(&cfg, kong.Configuration(kongtoml.Loader, "./config.toml"))
	log.Info().Interface("cfg", cfg).Msg("")

	traefikWriter := writer.NewTraefikWriter(cfg.TraefikWriter)

	frpsAdminManager := admin.NewFrpsAdminManager(cfg.FrpsAdmin, traefikWriter)
	frpsAdminManager.Start()
	// fetch all proxies at the beginning
	frpsAdminManager.FetchProxies()

	frpsPlugin := plugin.NewServer(frpsAdminManager)

	err := frpsPlugin.Serve(cfg.FrpsPlugin.Port)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to serve frps-plugin server")
	}
}
