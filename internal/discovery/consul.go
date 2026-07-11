package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

// ConsulProvider implements Provider using HashiCorp Consul
type ConsulProvider struct {
	client *api.Client
	logger *zap.Logger
}

// NewConsulProvider initializes a new Consul client
func NewConsulProvider(address string) (*ConsulProvider, error) {
	config := api.DefaultConfig()
	if address != "" {
		config.Address = address
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	logger, _ := zap.NewProduction()

	return &ConsulProvider{
		client: client,
		logger: logger,
	}, nil
}

// GetInstances returns healthy instances for a given service
func (p *ConsulProvider) GetInstances(serviceName string) ([]string, error) {
	entries, _, err := p.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances from consul: %w", err)
	}

	var instances []string
	for _, entry := range entries {
		// Example: "http://127.0.0.1:8080"
		// Depending on tags/metadata, you might infer the scheme. Defaulting to http.
		scheme := "http"
		address := entry.Service.Address
		if address == "" {
			address = entry.Node.Address
		}
		instances = append(instances, fmt.Sprintf("%s://%s:%d", scheme, address, entry.Service.Port))
	}

	return instances, nil
}

// Watch uses blocking queries to watch for changes to the service
func (p *ConsulProvider) Watch(ctx context.Context, serviceName string, updateCh chan<- []string) {
	var lastIndex uint64

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Consul watch stopped", zap.String("service", serviceName))
			return
		default:
			q := &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  2 * time.Minute,
			}

			entries, meta, err := p.client.Health().Service(serviceName, "", true, q.WithContext(ctx))
			if err != nil {
				// If context was canceled, just break out
				if ctx.Err() != nil {
					return
				}
				p.logger.Error("Error querying consul health", zap.Error(err))
				time.Sleep(5 * time.Second) // backoff
				continue
			}

			if meta.LastIndex > lastIndex {
				lastIndex = meta.LastIndex

				var instances []string
				for _, entry := range entries {
					scheme := "http"
					address := entry.Service.Address
					if address == "" {
						address = entry.Node.Address
					}
					instances = append(instances, fmt.Sprintf("%s://%s:%d", scheme, address, entry.Service.Port))
				}

				updateCh <- instances
			}
		}
	}
}
