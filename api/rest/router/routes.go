package router

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/tab58/code-context/api/rest/resolver"
	"github.com/tab58/code-context/api/rest/router/route"
	"github.com/tab58/code-context/api/rest/router/route/models"
)

func RegisterRoutes(api huma.API, r resolver.Resolver) {
	route.Register(route.RegisterArgs[models.HealthcheckRequest, models.HealthcheckResponse]{
		API: api,
		Operation: huma.Operation{
			Method:      "GET",
			Path:        "/health",
			Description: "Check the health of the server",
		},
		Handler: route.HandleHealthcheck(r),
	})

	route.Register(route.RegisterArgs[models.QueryRequest, models.QueryResponse]{
		API: api,
		Operation: huma.Operation{
			Method:      "POST",
			Path:        "/query",
			Description: "Query the knowledge graph",
		},
		Handler: route.HandleQuery(r),
	})
}
