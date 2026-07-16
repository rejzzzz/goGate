package discovery

import (
	"context"
	"fmt"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// EtcdProvider implements Provider using ETCD v3
type EtcdProvider struct {
	client *clientv3.Client
	logger *zap.Logger
}

// NewEtcdProvider initializes a new ETCD client
func NewEtcdProvider(address string) (*EtcdProvider, error) {
	if address == "" {
		address = os.Getenv("ETCD_ADDRESS")
		if address == "" {
			address = "127.0.0.1:2379"
		}
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{address},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	logger, _ := zap.NewProduction()

	return &EtcdProvider{
		client: client,
		logger: logger,
	}, nil
}

// GetInstances returns instances for a given service
func (p *EtcdProvider) GetInstances(serviceName string) ([]string, error) {
	prefix := fmt.Sprintf("/services/%s/", serviceName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := p.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get instances from etcd: %w", err)
	}

	var instances []string
	for _, ev := range resp.Kvs {
		instances = append(instances, string(ev.Value))
	}

	return instances, nil
}

// Watch uses ETCD's watch API to detect changes to the service prefix
func (p *EtcdProvider) Watch(ctx context.Context, serviceName string, updateCh chan<- []string) {
	prefix := fmt.Sprintf("/services/%s/", serviceName)
	rch := p.client.Watch(ctx, prefix, clientv3.WithPrefix())

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("ETCD watch stopped", zap.String("service", serviceName))
			return
		case wresp, ok := <-rch:
			if !ok {
				p.logger.Error("ETCD watch channel closed", zap.String("service", serviceName))
				return
			}
			if wresp.Canceled {
				p.logger.Error("ETCD watch canceled", zap.String("service", serviceName), zap.Error(wresp.Err()))
				return
			}

			// On any event, re-fetch all instances for simplicity and correctness
			instances, err := p.GetInstances(serviceName)
			if err != nil {
				p.logger.Error("Failed to fetch instances during watch update", zap.Error(err), zap.String("service", serviceName))
				continue
			}

			updateCh <- instances
		}
	}
}
