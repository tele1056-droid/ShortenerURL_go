package storage

type URLRecord struct {
	UUID          string `json:"uuid"`
	ShortURL      string `json:"short_url"`
	OriginalURL   string `json:"original_url"`
	UserID        string `json:"user_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
}
