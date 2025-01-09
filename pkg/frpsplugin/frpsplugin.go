package frpsplugin

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type Server struct {
}

func (s *Server) Serve(port int) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	log.Info().Int("port", port).Msg("start frpsplugin server")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func NewServer() *Server {
	return &Server{}
}
