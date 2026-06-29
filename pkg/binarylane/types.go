package binarylane

import "fmt"

type Record struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Data     string `json:"data"`
	TTL      int    `json:"ttl"`
	Priority *int   `json:"priority,omitempty"`
}

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type APIError struct {
	StatusCode int
	Body       string
	Response   *APIResponse // populated when error body is valid JSON
}

func (e *APIError) Error() string {
	return fmt.Sprintf("BinaryLane API error: status=%d body=%s", e.StatusCode, e.Body)
}
