package binarylane

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	zoneCache    sync.Map
	zoneCacheTTL = 5 * time.Minute
)

// cacheEntry holds a cached zone result with expiry.
type cacheEntry struct {
	zone      string
	expiresAt time.Time
}

// FindZoneByFqdn resolves the DNS zone for the given FQDN by progressively
// walking up the domain hierarchy until a zone with NS records is found.
// Results are cached for zoneCacheTTL.
func FindZoneByFqdn(ctx context.Context, fqdn string) (string, error) {
	if entry, ok := zoneCache.Load(fqdn); ok {
		ce := entry.(cacheEntry)
		if time.Now().Before(ce.expiresAt) {
			return ce.zone, nil
		}
		zoneCache.Delete(fqdn)
	}

	parts := strings.Split(fqdn, ".")
	// Try each suffix from longest to shortest
	for i := 0; i < len(parts)-1; i++ {
		candidate := strings.Join(parts[i:], ".")
		_, err := net.DefaultResolver.LookupNS(ctx, candidate)
		if err == nil {
			zoneCache.Store(fqdn, cacheEntry{zone: candidate, expiresAt: time.Now().Add(zoneCacheTTL)})
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no DNS zone with NS records found for %s", fqdn)
}

// ExtractSubdomain extracts the subdomain portion of the FQDN by removing
// the zone suffix. For example, ExtractSubdomain("_acme-challenge.example.com", "example.com")
// returns "_acme-challenge".
func ExtractSubdomain(fqdn, zone string) string {
	if fqdn == zone {
		return ""
	}
	suffix := "." + zone
	if strings.HasSuffix(fqdn, suffix) {
		return fqdn[:len(fqdn)-len(suffix)]
	}
	return fqdn
}
