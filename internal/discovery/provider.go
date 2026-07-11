package discovery

import "context"

// Provider defines the interface for dynamic service discovery
type Provider interface {
	// GetInstances returns the current list of healthy instances for a service
	GetInstances(serviceName string) ([]string, error)

	// Watch monitors a service for changes and pushes updates to the channel.
	// It should run until the context is canceled.
	Watch(ctx context.Context, serviceName string, updateCh chan<- []string)
}
