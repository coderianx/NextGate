package models

type CreateNewAPIKey struct {
	Title string `json:"title"`
}

type ApiKeySummary struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}
