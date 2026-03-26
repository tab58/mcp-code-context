package route

import (
	"context"

	"github.com/tab58/code-context/api/rest/resolver"
	"github.com/tab58/code-context/api/rest/router/route/models"
)

func HandleHealthcheck(r resolver.Resolver) Handler[models.HealthcheckRequest, models.HealthcheckResponse] {
	return func(ctx context.Context, req *models.HealthcheckRequest) (*models.HealthcheckResponse, error) {
		return &models.HealthcheckResponse{
			Body: struct {
				Message string `json:"message"`
				Version string `json:"version"`
			}{
				Message: "OK",
				Version: r.GetAppVersion(),
			},
		}, nil
	}
}
