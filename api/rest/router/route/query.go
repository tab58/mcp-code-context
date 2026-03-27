package route

import (
	"context"

	"github.com/tab58/code-context/api/rest/resolver"
	"github.com/tab58/code-context/api/rest/router/route/models"
)

func HandleQuery(r resolver.Resolver) Handler[models.QueryRequest, models.QueryResponse] {
	return func(ctx context.Context, req *models.QueryRequest) (*models.QueryResponse, error) {
		response, err := r.Query(ctx, req.Body.Query)
		if err != nil {
			return nil, err
		}
		return &models.QueryResponse{
			Body: struct {
				Result string `json:"result"`
			}{
				Result: response,
			},
		}, nil
	}
}
