// Package discovery provides speaker discovery functionality.
package discovery

import (
	"context"
	"time"
)

// Discoverer defines the interface for speaker discovery.
type Discoverer interface {
	Discover(ctx context.Context, timeout time.Duration) (string, error)
}

// Discover attempts to find a KEF speaker on the network.
// It tries SSDP first, then falls back to network scanning.
func Discover(ctx context.Context, timeout time.Duration) (string, error) {
	// Try SSDP discovery first
	if ip, err := DiscoverViaSSDP(ctx, timeout/2); err == nil {
		return ip, nil
	}

	// Fallback to network scanning
	return DiscoverViaNetworkScan(ctx, timeout/2)
}
