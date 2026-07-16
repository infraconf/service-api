package servicegrpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/infraconf/service-api/mtls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	*grpc.Server
	Listener net.Listener
	Source   *mtls.Source
}

func NewServer(ctx context.Context, cfg mtls.Config, opts ...grpc.ServerOption) (*Server, error) {
	source, err := mtls.NewSource(ctx, cfg.WorkloadAPIPath)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := mtls.ServerTLSConfig(source, cfg.Server)
	if err != nil {
		source.Close()
		return nil, err
	}

	serverOpts := append([]grpc.ServerOption{
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	}, opts...)

	return &Server{
		Server: grpc.NewServer(serverOpts...),
		Source: source,
	}, nil
}

func ListenAndServe(ctx context.Context, cfg mtls.Config, register func(grpc.ServiceRegistrar), opts ...grpc.ServerOption) (*Server, error) {
	if cfg.Server.Address == "" {
		return nil, fmt.Errorf("missing server address")
	}

	listener, err := net.Listen("tcp", cfg.Server.Address)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", cfg.Server.Address, err)
	}

	server, err := NewServer(ctx, cfg, opts...)
	if err != nil {
		listener.Close()
		return nil, err
	}
	server.Listener = listener

	if register != nil {
		register(server.Server)
	}

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	go func() {
		if err := server.Serve(listener); err != nil {
			server.Stop()
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	return server, nil
}

func ListenAndServeConfigFile(ctx context.Context, path string, register func(grpc.ServiceRegistrar), opts ...grpc.ServerOption) (*Server, error) {
	cfg, err := mtls.LoadConfigFile(path)
	if err != nil {
		return nil, err
	}
	return ListenAndServe(ctx, *cfg, register, opts...)
}

func (s *Server) Close() error {
	if s.Server != nil {
		s.GracefulStop()
	}
	if s.Listener != nil {
		_ = s.Listener.Close()
	}
	if s.Source != nil {
		return s.Source.Close()
	}
	return nil
}
