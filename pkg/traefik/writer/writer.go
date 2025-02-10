package writer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/117503445/goutils"
	"github.com/Masterminds/sprig/v3"
	"github.com/rs/zerolog/log"
)

func GetHttpContent(outputContent string) string {
	tmpl, err := template.New("example").Funcs(sprig.FuncMap()).Parse(outputContent)
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse template")
		return ""
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, nil); err != nil {
		log.Warn().Err(err).Msg("failed to execute template")
		return ""
	}

	return buf.String()
}

type TraefikWriterCfg struct {
	TemplatePath string `help:"traefik dynamic config template path" default:"./traefik.tmpl"`
	// dynamic_configs
	OutputPath string `help:"traefik dynamic config output path" default:"./frp_dynamic.yml"`

	FrpBaseUrl string `help:"frp base url" default:"http://frps"`
}

type TraefikWriter struct {
	cfg      *TraefikWriterCfg
	template string

	httpContent     string
	httpContentLock sync.RWMutex
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

func (w *TraefikWriter) Run() {
	http.HandleFunc("/traefik", func(writer http.ResponseWriter, r *http.Request) {
		w.httpContentLock.RLock()
		content := w.httpContent
		w.httpContentLock.RUnlock()

		writer.Write([]byte(content))
	})
	address, port := "0.0.0.0", "8081"
	log.Info().Str("address", address).Str("port", port).Msg("start traefik dynamic config server")
	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", address, port), nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
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

	content := strings.ReplaceAll(w.template, "$SERVICES$", servicesStr)

	w.httpContentLock.Lock()
	w.httpContent = GetHttpContent(content)
	w.httpContentLock.Unlock()

	err = goutils.WriteText(w.cfg.OutputPath, content)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to write output")
	} else {
		log.Info().Str("output", w.cfg.OutputPath).Str("content", content).Msg("write output success")
	}
}
