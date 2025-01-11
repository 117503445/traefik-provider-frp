package plugin

import (
	"117503445/traefik-provider-frp/pkg/frps/admin"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type Server struct {
	frpsAdmin *admin.FrpsAdminManager
}

func (s *Server) Serve(port int) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var err error

		defer func() {
			// always return accept response
			_, err = w.Write([]byte(`{"reject": false,"unchange": true}`))
			if err != nil {
				log.Error().Err(err).Msg("failed to write response")
			}
		}()

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("failed to read request body")
			return
		}

		op := r.URL.Query().Get("op")

		log.Info().Str("op", op).Str("payload", string(payload)).Msg("req")
		// log.Info().Str("op", op).Msg("req")

		// domainResult := gjson.GetBytes(payload, "content.metas.domain")
		// if !domainResult.Exists() {
		// 	return
		// }

		s.frpsAdmin.FetchProxies()

		// opResult := gjson.GetBytes(payload, "op")

		// serviceName := domainResult.String()
		// if serviceName == "" {
		// 	serviceName = gjson.GetBytes(payload, "content.proxy_name").String()
		// }

		// port := gjson.GetBytes(payload, "content.metas.port").Int()

	})
	log.Info().Int("port", port).Msg("start frpsplugin server")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func NewServer(frpsAdmin *admin.FrpsAdminManager) *Server {
	return &Server{
		frpsAdmin: frpsAdmin,
	}
}
