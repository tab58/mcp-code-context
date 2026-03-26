package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/tab58/code-context/api/rest/resolver"
)

type Router struct {
	api huma.API
	mux *http.ServeMux
}

func (r *Router) API() huma.API {
	return r.api
}

func (r *Router) Mux() *http.ServeMux {
	return r.mux
}

func New(cfg huma.Config, r resolver.Resolver) *Router {
	mux := http.NewServeMux()
	api := humago.New(mux, cfg)

	RegisterRoutes(api, r)

	return &Router{
		api: api,
		mux: mux,
	}
}
