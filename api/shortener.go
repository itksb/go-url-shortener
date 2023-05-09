package api

// ShortenRequest - .
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse - .
type ShortenResponse struct {
	Result string `json:"result"`
}

// ShortenBatchItemRequest - .
type ShortenBatchItemRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ShortenBatchRequest - .
type ShortenBatchRequest []ShortenBatchItemRequest

// ShortenBatchItemResponse - .
type ShortenBatchItemResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ShortenBatchResponse -.
type ShortenBatchResponse []ShortenBatchItemResponse

// ShortenDeleteBatchRequest - .
type ShortenDeleteBatchRequest []string

// ShortenInternalStatsResponse - .
type ShortenInternalStatsResponse struct {
	Urls  int `json:"urls"`
	Users int `json:"users"`
}
