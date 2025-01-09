package main

import (
	"fmt"

	"github.com/117503445/goutils"
	"github.com/alecthomas/kong"
	kongtoml "github.com/alecthomas/kong-toml"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
)

type frpsAdmin struct {
	BaseUrl  string `help:"frps-admin base url" default:"http://localhost:7500"`
	Username string `help:"frps-admin username"`
	Password string `help:"frps-admin password"`
}

func getFull(cfg *frpsAdmin) {
	log.Info().Interface("cfg", cfg).Msg("getFull")
	client := req.C().EnableDumpAll()
	url := fmt.Sprintf("%s/api/proxy/tcp", cfg.BaseUrl)
	resp, err := client.R().SetBasicAuth(cfg.Username, cfg.Password).Get(url)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get full config from frps-admin")
	}

	log.Info().Str("resp", resp.Dump()).Msg("get full config from frps-admin")
}

func main() {
	goutils.InitZeroLog()
	var cfg struct {
		FrpsAdmin *frpsAdmin `embed:"" prefix:"frps-admin."`
	}
	kong.Parse(&cfg, kong.Configuration(kongtoml.Loader, "./config.toml"))
	log.Info().Interface("cfg", cfg).Msg("main")

	getFull(cfg.FrpsAdmin)
}
