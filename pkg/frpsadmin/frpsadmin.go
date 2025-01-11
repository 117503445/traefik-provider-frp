package frpsadmin

import (
	"117503445/traefik-provider-frp/pkg/state"
	"encoding/json"
	"fmt"

	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
)

type FrpsAdminCfg struct {
	BaseUrl  string `help:"frps-admin base url" default:"http://localhost:7500"`
	Username string `help:"frps-admin username"`
	Password string `help:"frps-admin password"`
}

type FrpsAdminManager struct {
	cfg *FrpsAdminCfg
}

func NewFrpsAdminManager(cfg *FrpsAdminCfg) *FrpsAdminManager {
	return &FrpsAdminManager{
		cfg: cfg,
	}
}
func (m *FrpsAdminManager) SetState(serviceDestState *state.ServiceDestState) {
	log.Info().Interface("cfg", m.cfg).Msg("getFull")
	// client := req.C().EnableDumpAll()
	client := req.C()
	url := fmt.Sprintf("%s/api/proxy/tcp", m.cfg.BaseUrl)
	resp, err := client.R().SetBasicAuth(m.cfg.Username, m.cfg.Password).Get(url)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get full config from frps-admin")
	}

	type ProxyConfig struct {
		Name       string            `json:"name"`
		Type       string            `json:"type"`
		Metadatas  map[string]string `json:"metadatas"`
		RemotePort int               `json:"remotePort"`
	}
	type Proxy struct {
		Name            string       `json:"name"`
		Conf            *ProxyConfig `json:"conf"`
		TodayTrafficIn  uint64       `json:"todayTrafficIn"`
		TodayTrafficOut uint64       `json:"todayTrafficOut"`
		CurConns        int          `json:"curConns"`
		Status          string       `json:"status"`
	}
	var response struct {
		Proxies []Proxy `json:"proxies"`
	}
	if err := json.Unmarshal(resp.Bytes(), &response); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	result := make(map[string]int)

	for _, proxy := range response.Proxies {
		if proxy.Status == "online" && proxy.Conf != nil {
			domain := proxy.Conf.Metadatas["domain"]
			if domain == "" {
				domain = proxy.Conf.Name
			}
			result[domain] = proxy.Conf.RemotePort
		}
	}

	serviceDestState.SetMap(result)
}
