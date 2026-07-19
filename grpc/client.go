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

type TargetConfig struct {
	WorkloadAPIPath string
	Addr            string
	SPIFFEID        string
}

func Dial(ctx context.Context, target TargetConfig, opts ...grpc.DialOption) (*ClientConn, error) {
	if target.Addr == "" {
		return nil, fmt.Errorf("missing client address")
	}

	source, err := mtls.NewSource(ctx, target.WorkloadAPIPath)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := mtls.ClientTLSConfig(source, mtls.AuthorizationConfig{
		SPIFFEIDs: []string{target.SPIFFEID},
	})
	if err != nil {
		source.Close()
		return nil, err
	}

	dialOpts := append([]grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	}, opts...)

	conn, err := grpc.NewClient(target.Addr, dialOpts...)
	if err != nil {
		source.Close()
		return nil, fmt.Errorf("create gRPC client: %w", err)
	}

	return &ClientConn{ClientConn: conn, Source: source}, nil
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
