package rest

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tab58/code-context/api/rest/resolver"
	"github.com/tab58/code-context/api/rest/router"
)

type Server struct {
	api    huma.API
	router *http.ServeMux
	srv    *http.Server
}

func (s *Server) Start(addr string) {
	s.srv.Addr = addr
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %s\n", err)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) API() huma.API {
	return s.api
}

func NewServer(resolver resolver.Resolver) *Server {
	config := huma.Config{
		OpenAPI: &huma.OpenAPI{
			OpenAPI: "3.1.0",
			Info: &huma.Info{
				Title:       "code-context",
				Version:     resolver.GetAppVersion(),
				Description: "code-context",
			},
		},
	}

	r := router.New(config, resolver)

	return &Server{
		api:    r.API(),
		router: r.Mux(),
		srv:    &http.Server{Handler: r.Mux()},
	}
}
