package solver

import (
	"context"
	"testing"

	"github.com/drachemann/cert-manager-webhook-binarylane/pkg/binarylane"
	whapi "github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/rest"
)

// mockClient is a testify/mock implementation of binarylane.Client.
type mockClient struct {
	mock.Mock
}

func (m *mockClient) CreateRecord(ctx context.Context, domain string, record binarylane.Record) (*binarylane.Record, error) {
	args := m.Called(ctx, domain, record)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*binarylane.Record), args.Error(1)
}

func (m *mockClient) DeleteRecord(ctx context.Context, domain string, recordID int) error {
	args := m.Called(ctx, domain, recordID)
	return args.Error(0)
}

func (m *mockClient) GetRecord(ctx context.Context, domain string, recordID int) (*binarylane.Record, error) {
	args := m.Called(ctx, domain, recordID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*binarylane.Record), args.Error(1)
}

func TestSolver_Name(t *testing.T) {
	s := &Solver{}
	assert.Equal(t, "binarylane", s.Name())
}

func TestSolver_Present_CreatesTXTRecord(t *testing.T) {
	client := new(mockClient)
	s := &Solver{client: client}

	ch := &whapi.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com",
		Key:          "test-key-value",
	}

	expectedRecord := binarylane.Record{
		Name: "_acme-challenge",
		Type: "TXT",
		Data: "test-key-value",
		TTL:  60,
	}

	client.On("CreateRecord",
		mock.Anything,
		"example.com",
		mock.MatchedBy(func(r binarylane.Record) bool {
			return r.Name == expectedRecord.Name &&
				r.Type == expectedRecord.Type &&
				r.Data == expectedRecord.Data
		}),
	).Return(&expectedRecord, nil)

	err := s.Present(ch)
	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestSolver_Present_PropagatesAPIErrors(t *testing.T) {
	client := new(mockClient)
	s := &Solver{client: client}

	ch := &whapi.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com",
		Key:          "test-key-value",
	}

	apiErr := &binarylane.APIError{StatusCode: 500, Body: "internal error"}

	client.On("CreateRecord",
		mock.Anything,
		"example.com",
		mock.AnythingOfType("binarylane.Record"),
	).Return(nil, apiErr)

	err := s.Present(ch)
	assert.Error(t, err)
	assert.Equal(t, apiErr, err)
	client.AssertExpectations(t)
}

func TestSolver_Present_EmptyFQDN(t *testing.T) {
	client := new(mockClient)
	s := &Solver{client: client}

	ch := &whapi.ChallengeRequest{
		ResolvedFQDN: "",
		Key:          "test-key-value",
	}

	// No CreateRecord should be called for empty FQDN
	err := s.Present(ch)
	if err == nil {
		t.Fatal("expected error for empty FQDN")
	}
	assert.Contains(t, err.Error(), "FQDN")
	// Ensure CreateRecord was never called
	client.AssertNotCalled(t, "CreateRecord", mock.Anything, mock.Anything, mock.Anything)
}

func TestSolver_Present_EmptyKey(t *testing.T) {
	client := new(mockClient)
	s := &Solver{client: client}

	ch := &whapi.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com",
		Key:          "",
	}

	err := s.Present(ch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key")
	client.AssertNotCalled(t, "CreateRecord", mock.Anything, mock.Anything, mock.Anything)
}

func TestSolver_Present_Idempotent(t *testing.T) {
	client := new(mockClient)
	s := &Solver{client: client}

	ch := &whapi.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com",
		Key:          "test-key-value",
	}

	expectedRecord := &binarylane.Record{
		ID:   123,
		Name: "_acme-challenge",
		Type: "TXT",
		Data: "test-key-value",
		TTL:  60,
	}

	client.On("CreateRecord",
		mock.Anything,
		"example.com",
		mock.AnythingOfType("binarylane.Record"),
	).Return(expectedRecord, nil).Once()

	// First call creates the record.
	err := s.Present(ch)
	assert.NoError(t, err)

	// Second call is idempotent — no additional CreateRecord.
	err = s.Present(ch)
	assert.NoError(t, err)

	client.AssertExpectations(t)
}

func TestSolver_CleanUp_DeletesTXTRecord(t *testing.T) {
	client := new(mockClient)
	s := &Solver{client: client}

	ch := &whapi.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com",
		Key:          "test-key-value",
	}

	expectedRecord := &binarylane.Record{
		ID:   99,
		Name: "_acme-challenge",
		Type: "TXT",
		Data: "test-key-value",
		TTL:  60,
	}

	client.On("CreateRecord",
		mock.Anything,
		"example.com",
		mock.AnythingOfType("binarylane.Record"),
	).Return(expectedRecord, nil)

	err := s.Present(ch)
	assert.NoError(t, err)

	client.On("DeleteRecord",
		mock.Anything,
		"example.com",
		99,
	).Return(nil)

	err = s.CleanUp(ch)
	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestSolver_CleanUp_NoRecordStored(t *testing.T) {
	client := new(mockClient)
	s := &Solver{client: client}

	ch := &whapi.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.no-record.example.com",
	}

	// No DeleteRecord should be called when no record was previously stored
	// (idempotent — CleanUp should not error)
	err := s.CleanUp(ch)
	assert.NoError(t, err)
	client.AssertNotCalled(t, "DeleteRecord",
		mock.Anything, mock.Anything, mock.Anything)
}

func TestSolver_Initialize_CreatesClient(t *testing.T) {
	s := &Solver{}
	err := s.Initialize(&rest.Config{
		Host: "https://localhost:6443",
	}, nil)
	// kubernetes.NewForConfig creates the client eagerly; config is only
	// validated at API-call time, not construction time.
	assert.NoError(t, err)
	assert.NotNil(t, s.kubeClient, "kubeClient should be set after Initialize")
}

func TestSolver_TTL_Default(t *testing.T) {
	client := new(mockClient)
	s := &Solver{client: client}

	ch := &whapi.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com",
		Key:          "test-key-value",
	}

	client.On("CreateRecord",
		mock.Anything,
		"example.com",
		mock.MatchedBy(func(r binarylane.Record) bool {
			return r.TTL == 60 // default when TTL is zero
		}),
	).Return(&binarylane.Record{ID: 1, TTL: 60}, nil)

	err := s.Present(ch)
	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestSolver_TTL_Configured(t *testing.T) {
	client := new(mockClient)
	s := &Solver{client: client, TTL: 300}

	ch := &whapi.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com",
		Key:          "test-key-value",
	}

	client.On("CreateRecord",
		mock.Anything,
		"example.com",
		mock.MatchedBy(func(r binarylane.Record) bool {
			return r.TTL == 300
		}),
	).Return(&binarylane.Record{ID: 1, TTL: 300}, nil)

	err := s.Present(ch)
	assert.NoError(t, err)
	client.AssertExpectations(t)
}
