package frpsplugin

import (
	"117503445/traefik-provider-frp/pkg/frpsadmin"
	"117503445/traefik-provider-frp/pkg/state"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

type Server struct {
	serviceDestState *state.ServiceDestState
	frpsAdmin        *frpsadmin.FrpsAdminManager
}

func (s *Server) Serve(port int) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var err error

		defer func() {
			_, err = w.Write([]byte(`{"reject": false,"unchange": true}`))
			if err != nil {
				log.Error().Err(err).Msg("failed to write response")
			}
		}()

		// get args['op']

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("failed to read request body")
			return
		}

		op := r.URL.Query().Get("op")

		log.Info().Str("op", op).Str("payload", string(payload)).Msg("req")

		domainResult := gjson.GetBytes(payload, "content.metas.domain")
		if !domainResult.Exists() {
			return
		}

		// opResult := gjson.GetBytes(payload, "op")

		serviceName := domainResult.String()
		if serviceName == "" {
			serviceName = gjson.GetBytes(payload, "content.proxy_name").String()
		}

		// port := gjson.GetBytes(payload, "content.metas.port").Int()

	})
	log.Info().Int("port", port).Msg("start frpsplugin server")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func NewServer(serviceDestState *state.ServiceDestState, frpsAdmin *frpsadmin.FrpsAdminManager) *Server {
	return &Server{
		serviceDestState: serviceDestState,
		frpsAdmin:        frpsAdmin,
	}
}
