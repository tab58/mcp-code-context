package models

type QueryRequest struct {
	Body struct {
		Query string `json:"query"`
	}
}

type QueryResponse struct {
	Body struct {
		Result string `json:"result"`
	}
}
