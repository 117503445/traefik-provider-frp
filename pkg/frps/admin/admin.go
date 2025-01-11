package admin

import (
	"fmt"

	"github.com/117503445/goutils"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

type FrpsAdminCfg struct {
	BaseUrl  string `help:"frps-admin base url" default:"http://localhost:7500"`
	Username string `help:"frps-admin username"`
	Password string `help:"frps-admin password"`
}

type FrpsAdminManager struct {
	cfg                *FrpsAdminCfg
	latestTaskExecutor *goutils.LatestTaskExecutor
}

func NewFrpsAdminManager(cfg *FrpsAdminCfg) *FrpsAdminManager {
	return &FrpsAdminManager{
		cfg:                cfg,
		latestTaskExecutor: goutils.NewLatestTaskExecutor(),
	}
}
func (m *FrpsAdminManager) FetchProxies() {
	m.latestTaskExecutor.AddTask(func() {
		// log.Info().Interface("cfg", m.cfg).Msg("getFull")

		// client := req.C().EnableDumpAll()
		client := req.C()
		url := fmt.Sprintf("%s/api/proxy/tcp", m.cfg.BaseUrl)
		resp, err := client.R().SetBasicAuth(m.cfg.Username, m.cfg.Password).Get(url)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get full config from frps-admin")
		}

		result := make(map[string]int)

		proxiesArray := gjson.GetBytes(resp.Bytes(), "proxies")
		proxiesArray.ForEach(func(key, proxy gjson.Result) bool {
			if proxy.Get("status").String() == "online" {
				conf := proxy.Get("conf")
				domainResult := conf.Get("metadatas.domain")
				if !domainResult.Exists() {
					return true
				}
				domain := domainResult.String()
				if domain == "" {
					domain = conf.Get("name").String()
				}
				result[domain] = int(conf.Get("remotePort").Int())
			}
			return true
		})
	})
}
