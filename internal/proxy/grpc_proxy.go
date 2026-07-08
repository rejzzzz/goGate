package proxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/mwitkow/grpc-proxy/proxy"
	"github.com/rejzzzz/goGate/internal/loadbalancer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TargetContextKey is used to pass the selected upstream to the director
const TargetContextKey = "grpc_target"

// GRPCProxy represents our transparent gRPC proxy
type GRPCProxy struct {
	Server *grpc.Server

	mu          sync.RWMutex
	clientConns map[string]*grpc.ClientConn
}

// NewGRPCProxy creates a new gRPC transparent proxy
func NewGRPCProxy() *GRPCProxy {
	p := &GRPCProxy{
		clientConns: make(map[string]*grpc.ClientConn),
	}

	director := func(ctx context.Context, fullMethodName string) (context.Context, grpc.ClientConnInterface, error) {
		target, ok := ctx.Value(TargetContextKey).(*loadbalancer.Upstream)
		if !ok || target == nil {
			return nil, nil, fmt.Errorf("no target provided in context")
		}

		conn, err := p.getClientConn(target.URL)
		if err != nil {
			return nil, nil, err
		}

		return ctx, conn, nil
	}

	p.Server = grpc.NewServer(
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
	)

	return p
}

// getClientConn retrieves or creates a cached gRPC connection to the given address
func (p *GRPCProxy) getClientConn(addr string) (*grpc.ClientConn, error) {
	p.mu.RLock()
	conn, exists := p.clientConns[addr]
	p.mu.RUnlock()

	if exists {
		return conn, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double check
	if conn, exists := p.clientConns[addr]; exists {
		return conn, nil
	}

	// Dial the upstream gRPC server
	newConn, err := grpc.NewClient(addr,
		grpc.WithDefaultCallOptions(grpc.CallCustomCodec(proxy.Codec().(grpc.Codec))),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	p.clientConns[addr] = newConn
	return newConn, nil
}
