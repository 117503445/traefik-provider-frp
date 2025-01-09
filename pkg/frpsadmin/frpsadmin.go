package frpsadmin

import (
	"fmt"

	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
)

type FrpsAdminCfg struct {
	BaseUrl  string `help:"frps-admin base url" default:"http://localhost:7500"`
	Username string `help:"frps-admin username"`
	Password string `help:"frps-admin password"`
}

func GetRoutes(cfg *FrpsAdminCfg) {
	log.Info().Interface("cfg", cfg).Msg("getFull")
	client := req.C().EnableDumpAll()
	url := fmt.Sprintf("%s/api/proxy/tcp", cfg.BaseUrl)
	resp, err := client.R().SetBasicAuth(cfg.Username, cfg.Password).Get(url)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get full config from frps-admin")
	}

	log.Info().Str("resp", resp.Dump()).Msg("get full config from frps-admin")
}
