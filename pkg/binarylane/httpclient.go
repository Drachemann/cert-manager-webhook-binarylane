package binarylane

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// httpClient is an HTTP-based implementation of Client.
type httpClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
	maxRetries int           // maximum retry attempts for 5xx/network errors
	backoff    time.Duration // initial backoff duration
}

// NewHTTPClient creates a new Client backed by an HTTP client.
// Defaults: maxRetries=3, backoff=1s.
func NewHTTPClient(baseURL, token string) Client {
	return &httpClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{},
		maxRetries: 3,
		backoff:    1 * time.Second,
	}
}

// do performs the HTTP request with retry on 5xx and network errors.
func (c *httpClient) do(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.backoff * time.Duration(1<<(attempt-1))):
			}
		}
		req, err := c.buildRequest(ctx, method, url, body)
		if err != nil {
			return nil, err
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode < 500 {
			return resp, nil
		}
		resp.Body.Close()
		lastErr = parseAPIError(resp)
	}
	return nil, lastErr
}

// buildRequest constructs an HTTP request for the given parameters.
func (c *httpClient) buildRequest(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *httpClient) CreateRecord(ctx context.Context, domain string, record Record) (*Record, error) {
	url := fmt.Sprintf("%s/domains/%s/records", c.baseURL, domain)
	resp, err := c.do(ctx, http.MethodPost, url, record)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var rec Record
		if err := json.NewDecoder(resp.Body).Decode(&rec); err != nil {
			return nil, err
		}
		return &rec, nil
	}
	return nil, parseAPIError(resp)
}

func (c *httpClient) DeleteRecord(ctx context.Context, domain string, recordID int) error {
	url := fmt.Sprintf("%s/domains/%s/records/%d", c.baseURL, domain, recordID)
	resp, err := c.do(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return parseAPIError(resp)
}

func (c *httpClient) GetRecord(ctx context.Context, domain string, recordID int) (*Record, error) {
	url := fmt.Sprintf("%s/domains/%s/records/%d", c.baseURL, domain, recordID)
	resp, err := c.do(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var rec Record
		if err := json.NewDecoder(resp.Body).Decode(&rec); err != nil {
			return nil, err
		}
		return &rec, nil
	}
	return nil, parseAPIError(resp)
}

// parseAPIError reads the response body (capped at 1MB) and constructs
// an APIError. If the body is valid JSON matching APIResponse, the
// Response field is populated.
func parseAPIError(resp *http.Response) *APIError {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB cap
	ae := &APIError{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}
	var apiResp APIResponse
	if json.Unmarshal(body, &apiResp) == nil {
		ae.Response = &apiResp
	}
	return ae
}
