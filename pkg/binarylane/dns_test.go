package binarylane

import (
	"context"
	"testing"
	"time"
)

func TestFindZoneByFqdn_NoZone(t *testing.T) {
	// In the unit-test environment there are no real NS records for this
	// non-existent domain, so FindZoneByFqdn returns an error.
	ctx := context.Background()
	_, err := FindZoneByFqdn(ctx, "no-ns-records-12345.example.invalid")
	if err == nil {
		t.Fatal("expected error for domain with no NS records in test env")
	}
}

func TestFindZoneByFqdn_Cached(t *testing.T) {
	// Populate the cache directly and verify a cached hit.
	zoneCache.Store("_acme-challenge.cached.example.com", cacheEntry{
		zone:      "cached.example.com",
		expiresAt: time.Now().Add(1 * time.Hour),
	})

	ctx := context.Background()
	zone, err := FindZoneByFqdn(ctx, "_acme-challenge.cached.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zone != "cached.example.com" {
		t.Errorf("expected zone 'cached.example.com', got %q", zone)
	}
}

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		name string
		fqdn string
		zone string
		want string
	}{
		{
			name: "exact match",
			fqdn: "example.com",
			zone: "example.com",
			want: "",
		},
		{
			name: "single subdomain",
			fqdn: "_acme-challenge.example.com",
			zone: "example.com",
			want: "_acme-challenge",
		},
		{
			name: "nested subdomain",
			fqdn: "deep.nested.example.com",
			zone: "example.com",
			want: "deep.nested",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractSubdomain(tt.fqdn, tt.zone)
			if got != tt.want {
				t.Errorf("ExtractSubdomain(%q, %q) = %q, want %q",
					tt.fqdn, tt.zone, got, tt.want)
			}
		})
	}
}
