package binarylane

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServerAndClient(t *testing.T, handler http.HandlerFunc) (Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client := NewHTTPClient(srv.URL, "test-api-token")
	return client, srv
}

func TestCreateRecord_Success(t *testing.T) {
	want := &Record{
		ID:   123,
		Name: "_acme-challenge.example.com",
		Type: "TXT",
		Data: "test-token-value",
		TTL:  60,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/domains/example.com/records" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-token" {
			t.Errorf("unexpected auth header: %s", auth)
		}

		var rec Record
		if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
			t.Errorf("failed to decode body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(want)
	}

	client, srv := newTestServerAndClient(t, handler)
	defer srv.Close()

	ctx := context.Background()
	record := Record{
		Name: "_acme-challenge.example.com",
		Type: "TXT",
		Data: "test-token-value",
		TTL:  60,
	}

	got, err := client.CreateRecord(ctx, "example.com", record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil record")
	}
	if got.ID != want.ID {
		t.Errorf("expected ID %d, got %d", want.ID, got.ID)
	}
	if got.Name != want.Name {
		t.Errorf("expected Name %q, got %q", want.Name, got.Name)
	}
	if got.Type != want.Type {
		t.Errorf("expected Type %q, got %q", want.Type, got.Type)
	}
	if got.Data != want.Data {
		t.Errorf("expected Data %q, got %q", want.Data, got.Data)
	}
	if got.TTL != want.TTL {
		t.Errorf("expected TTL %d, got %d", want.TTL, got.TTL)
	}
}

func TestCreateRecord_201(t *testing.T) {
	want := &Record{
		ID:   456,
		Name: "_acme-challenge.example.com",
		Type: "TXT",
		Data: "test-token-value",
		TTL:  60,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(want)
	}

	client, srv := newTestServerAndClient(t, handler)
	defer srv.Close()

	ctx := context.Background()
	record := Record{
		Name: "_acme-challenge.example.com",
		Type: "TXT",
		Data: "test-token-value",
		TTL:  60,
	}

	got, err := client.CreateRecord(ctx, "example.com", record)
	if err != nil {
		t.Fatalf("unexpected error for 201 response: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("expected ID %d, got %d", want.ID, got.ID)
	}
}

func TestRetryOnServerError(t *testing.T) {
	callCount := 0
	want := &Record{
		ID:   789,
		Name: "_acme-challenge.example.com",
		Type: "TXT",
		Data: "test-token-value",
		TTL:  60,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(want)
	}

	client, srv := newTestServerAndClient(t, handler)
	defer srv.Close()

	ctx := context.Background()
	record := Record{
		Name: "_acme-challenge.example.com",
		Type: "TXT",
		Data: "test-token-value",
		TTL:  60,
	}

	got, err := client.CreateRecord(ctx, "example.com", record)
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if got.ID != 789 {
		t.Errorf("expected ID 789, got %d", got.ID)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry), got %d", callCount)
	}
}

func TestCreateRecord_AuthFailure(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "Invalid API token",
		})
	}

	client, srv := newTestServerAndClient(t, handler)
	defer srv.Close()

	ctx := context.Background()
	record := Record{
		Name: "_acme-challenge.example.com",
		Type: "TXT",
		Data: "test-token-value",
		TTL:  60,
	}

	_, err := client.CreateRecord(ctx, "example.com", record)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestCreateRecord_APIError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}

	client, srv := newTestServerAndClient(t, handler)
	defer srv.Close()

	ctx := context.Background()
	record := Record{
		Name: "_acme-challenge.example.com",
		Type: "TXT",
		Data: "test-token-value",
		TTL:  60,
	}

	_, err := client.CreateRecord(ctx, "example.com", record)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

func TestDeleteRecord_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/domains/example.com/records/123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-token" {
			t.Errorf("unexpected auth header: %s", auth)
		}
		w.WriteHeader(http.StatusNoContent)
	}

	client, srv := newTestServerAndClient(t, handler)
	defer srv.Close()

	ctx := context.Background()
	err := client.DeleteRecord(ctx, "example.com", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRecord_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "Record not found",
		})
	}

	client, srv := newTestServerAndClient(t, handler)
	defer srv.Close()

	ctx := context.Background()
	err := client.DeleteRecord(ctx, "example.com", 999)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestGetRecord_Success(t *testing.T) {
	want := &Record{
		ID:   456,
		Name: "_acme-challenge.example.com",
		Type: "TXT",
		Data: "another-token",
		TTL:  120,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/domains/example.com/records/456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-token" {
			t.Errorf("unexpected auth header: %s", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}

	client, srv := newTestServerAndClient(t, handler)
	defer srv.Close()

	ctx := context.Background()
	got, err := client.GetRecord(ctx, "example.com", 456)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil record")
	}
	if got.ID != want.ID {
		t.Errorf("expected ID %d, got %d", want.ID, got.ID)
	}
	if got.Name != want.Name {
		t.Errorf("expected Name %q, got %q", want.Name, got.Name)
	}
	if got.Data != want.Data {
		t.Errorf("expected Data %q, got %q", want.Data, got.Data)
	}
	if got.TTL != want.TTL {
		t.Errorf("expected TTL %d, got %d", want.TTL, got.TTL)
	}
}

func TestContextCancellation(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}

	client, srv := newTestServerAndClient(t, handler)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	record := Record{
		Name: "_acme-challenge.example.com",
		Type: "TXT",
		Data: "deadline-test",
		TTL:  60,
	}

	_, err := client.CreateRecord(ctx, "example.com", record)
	if err == nil {
		t.Fatal("expected context deadline exceeded error")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") && err != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded error, got: %v", err)
	}
}

func TestAuthHeader(t *testing.T) {
	var capturedAuth string

	handler := func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&Record{
			ID:   1,
			Name: "test.example.com",
			Type: "TXT",
			Data: "data",
			TTL:  60,
		})
	}

	client, srv := newTestServerAndClient(t, handler)
	defer srv.Close()

	ctx := context.Background()
	_, err := client.CreateRecord(ctx, "example.com", Record{
		Name: "test.example.com",
		Type: "TXT",
		Data: "data",
		TTL:  60,
	})
	// Even if the implementation returns an error, we can still check the auth header
	// For the RED phase, the current no-op won't set it — test will fail
	_ = err

	if capturedAuth != "Bearer test-api-token" {
		t.Errorf("expected Authorization header 'Bearer test-api-token', got %q", capturedAuth)
	}
}
