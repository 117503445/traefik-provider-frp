package writer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/117503445/goutils"
	"github.com/rs/zerolog/log"
)

type TraefikWriterCfg struct {
	TemplatePath string `help:"traefik dynamic config template path" default:"./traefik.tmpl"`
	// dynamic_configs
	OutputPath string `help:"traefik dynamic config output path" default:"./frp_dynamic.yml"`

	FrpBaseUrl string `help:"frp base url" default:"http://frps"`
}

type TraefikWriter struct {
	cfg      *TraefikWriterCfg
	template string
}

func NewTraefikWriter(cfg *TraefikWriterCfg) *TraefikWriter {
	template, err := goutils.ReadText(cfg.TemplatePath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read traefik template")
	}

	return &TraefikWriter{
		cfg:      cfg,
		template: template,
	}
}

func (w *TraefikWriter) Write(DomainPort map[string]int) {
	type TraefikService struct {
		Service string `json:"service"`
		Url     string `json:"url"`
	}
	services := make([]TraefikService, 0, len(DomainPort))
	for domain, port := range DomainPort {
		services = append(services, TraefikService{
			Service: domain,
			// Url:     w.cfg.FrpBaseUrl + ":" + port,
			Url: fmt.Sprintf("%s:%d", w.cfg.FrpBaseUrl, port),
		})
	}

	servicesBytes, err := json.Marshal(services)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal services")
	}
	servicesStr := string(servicesBytes)

	outputContent := strings.ReplaceAll(w.template, "$SERVICES$", servicesStr)

	err = goutils.WriteText(w.cfg.OutputPath, outputContent)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to write output")
	}else{
		log.Info().Str("output", w.cfg.OutputPath).Str("content", outputContent).Msg("write output success")
	}
}
