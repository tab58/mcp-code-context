package models

type HealthcheckRequest struct{}

type HealthcheckResponse struct {
	Body struct {
		Message string `json:"message"`
		Version string `json:"version"`
	}
}
