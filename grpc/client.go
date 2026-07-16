package servicegrpc

import (
	"context"
	"fmt"

	"github.com/infraconf/service-api/mtls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ClientConn struct {
	*grpc.ClientConn
	Source *mtls.Source
}

func Dial(ctx context.Context, cfg mtls.Config, opts ...grpc.DialOption) (*ClientConn, error) {
	if cfg.Client.Target == "" {
		return nil, fmt.Errorf("missing client target")
	}

	source, err := mtls.NewSource(ctx, cfg.WorkloadAPIPath)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := mtls.ClientTLSConfig(source, cfg.Client)
	if err != nil {
		source.Close()
		return nil, err
	}

	dialOpts := append([]grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	}, opts...)

	conn, err := grpc.NewClient(cfg.Client.Target, dialOpts...)
	if err != nil {
		source.Close()
		return nil, fmt.Errorf("create gRPC client: %w", err)
	}

	return &ClientConn{ClientConn: conn, Source: source}, nil
}

func DialConfigFile(ctx context.Context, path string, opts ...grpc.DialOption) (*ClientConn, error) {
	cfg, err := mtls.LoadConfigFile(path)
	if err != nil {
		return nil, err
	}
	return Dial(ctx, *cfg, opts...)
}

func (c *ClientConn) Close() error {
	var connErr error
	if c.ClientConn != nil {
		connErr = c.ClientConn.Close()
	}
	if c.Source != nil {
		if err := c.Source.Close(); err != nil && connErr == nil {
			connErr = err
		}
	}
	return connErr
}
