package solver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	whapi "github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/drachemann/cert-manager-webhook-binarylane/pkg/binarylane"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// secretKeyRef references a Kubernetes Secret containing an API key.
type secretKeyRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Key       string `json:"key,omitempty"`
}

// binarylaneConfig is the solver configuration stored in the
// Issuer/ClusterIssuer resource's spec.acme.solvers.dns01.webhook.config.
type binarylaneConfig struct {
	APIKeySecretRef secretKeyRef `json:"apiKeySecretRef"`
}

// recordStore tracks DNS record IDs created by Present,
// keyed by FQDN, for later deletion in CleanUp.
type recordStore struct {
	mu  sync.Mutex
	ids map[string]int // FQDN → record ID
}

func (rs *recordStore) store(fqdn string, id int) {
	rs.mu.Lock()
	if rs.ids == nil {
		rs.ids = make(map[string]int)
	}
	rs.ids[fqdn] = id
	rs.mu.Unlock()
}

func (rs *recordStore) load(fqdn string) (int, bool) {
	rs.mu.Lock()
	id, ok := rs.ids[fqdn]
	rs.mu.Unlock()
	return id, ok
}

func (rs *recordStore) exists(fqdn string) bool {
	rs.mu.Lock()
	_, ok := rs.ids[fqdn]
	rs.mu.Unlock()
	return ok
}

// Solver implements the DNS01 solver for BinaryLane.
type Solver struct {
	client     binarylane.Client
	kubeClient kubernetes.Interface
	records    recordStore
	TTL        int // configurable TTL; defaults to 60 when zero
}

// Name returns the solver name.
func (s *Solver) Name() string { return "binarylane" }

// Present creates a TXT record for the DNS01 challenge.
func (s *Solver) Present(ch *whapi.ChallengeRequest) error {
	if ch.ResolvedFQDN == "" {
		return fmt.Errorf("FQDN is required")
	}
	if ch.Key == "" {
		return fmt.Errorf("challenge key is required")
	}

	ctx := context.Background()

	// Ensure a client exists (created in Initialize or injected in tests).
	if s.client == nil {
		if s.kubeClient == nil {
			return fmt.Errorf("solver not initialized: no client")
		}
		if err := s.createClientFromConfig(ctx, ch); err != nil {
			return fmt.Errorf("creating api client: %w", err)
		}
	}

	zone, err := binarylane.FindZoneByFqdn(ctx, ch.ResolvedFQDN)
	if err != nil {
		return err
	}

	subdomain := binarylane.ExtractSubdomain(ch.ResolvedFQDN, zone)
	recordName := subdomain
	if recordName == "" {
		// Apex record: subdomain is empty when FQDN == zone
		recordName = ch.ResolvedFQDN
	}

	ttl := s.TTL
	if ttl == 0 {
		ttl = 60
	}

	record := binarylane.Record{
		Name: recordName,
		Type: "TXT",
		Data: ch.Key,
		TTL:  ttl,
	}

	// Idempotent: if we already created a record for this FQDN, skip.
	if s.records.exists(ch.ResolvedFQDN) {
		return nil
	}

	created, err := s.client.CreateRecord(ctx, zone, record)
	if err != nil {
		return err
	}

	s.records.store(ch.ResolvedFQDN, created.ID)
	return nil
}

// CleanUp removes the TXT record created by Present.
func (s *Solver) CleanUp(ch *whapi.ChallengeRequest) error {
	id, ok := s.records.load(ch.ResolvedFQDN)
	if !ok {
		return nil
	}

	ctx := context.Background()

	// Ensure a client exists.
	if s.client == nil {
		if s.kubeClient == nil {
			return fmt.Errorf("solver not initialized: no client")
		}
		if err := s.createClientFromConfig(ctx, ch); err != nil {
			return fmt.Errorf("creating api client: %w", err)
		}
	}

	zone, err := binarylane.FindZoneByFqdn(ctx, ch.ResolvedFQDN)
	if err != nil {
		return err
	}

	return s.client.DeleteRecord(ctx, zone, id)
}

// Initialize creates the Kubernetes client from the provided config.
func (s *Solver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return fmt.Errorf("creating kubernetes client: %w", err)
	}
	s.kubeClient = cl
	return nil
}

// createClientFromConfig reads the API key from the referenced Kubernetes
// secret and constructs a BinaryLane HTTP client.
func (s *Solver) createClientFromConfig(ctx context.Context, ch *whapi.ChallengeRequest) error {
	var cfg binarylaneConfig
	if ch.Config != nil {
		if err := json.Unmarshal(ch.Config.Raw, &cfg); err != nil {
			return fmt.Errorf("unmarshaling solver config: %w", err)
		}
	}
	if cfg.APIKeySecretRef.Name == "" {
		return fmt.Errorf("apiKeySecretRef.name is required in solver config")
	}

	ns := cfg.APIKeySecretRef.Namespace
	if ns == "" {
		ns = ch.ResourceNamespace
	}

	key := cfg.APIKeySecretRef.Key
	if key == "" {
		key = "api-key"
	}

	secret, err := s.kubeClient.CoreV1().Secrets(ns).Get(ctx, cfg.APIKeySecretRef.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("reading secret %s/%s: %w", ns, cfg.APIKeySecretRef.Name, err)
	}

	apiKey, ok := secret.Data[key]
	if !ok {
		return fmt.Errorf("secret %s/%s has no key %q", ns, cfg.APIKeySecretRef.Name, key)
	}

	s.client = binarylane.NewHTTPClient(s.baseURL(), string(apiKey))
	return nil
}

// baseURL returns the BinaryLane API base URL.
func (s *Solver) baseURL() string { return "https://api.binarylane.com.au/v2" }
