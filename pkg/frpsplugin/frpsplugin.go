package frpsplugin

import (
	"117503445/traefik-provider-frp/pkg/state"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type Server struct {
	serviceDestState *state.ServiceDestState
}

func (s *Server) Serve(port int) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var err error
		// get args['op']
		op := r.URL.Query().Get("op")

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("failed to read request body")
			return
		}

		log.Info().Str("op", op).Str("payload", string(payload)).Msg("req")

		_, err = w.Write([]byte(`{"reject": false,"unchange": true}`))
		if err != nil {
			log.Error().Err(err).Msg("failed to write response")
		}
	})
	log.Info().Int("port", port).Msg("start frpsplugin server")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func NewServer(serviceDestState *state.ServiceDestState) *Server {
	return &Server{
		serviceDestState: serviceDestState,
	}
}
