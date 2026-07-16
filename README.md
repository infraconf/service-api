# Service API

Go library for the `service.v1.BackendRuntime` protobuf API with gRPC client
and server helpers that use SPIFFE-based mTLS.

## Generate protobuf code

```sh
make build-proto
```

The Makefile expects `protoc-gen-go` and `protoc-gen-go-grpc` in `PATH`.

## Configuration

All connection and peer authorization settings are YAML-driven. See
`config.example.yaml`.

```yaml
workload_api_path: unix:///tmp/spire-agent/public/api.sock

server:
  address: :8443
  authorized_peers:
    spiffe_ids:
      - spiffe://example.org/client
    trust_domains: []
    allow_any: false

client:
  target: localhost:8443
  authorized_peers:
    spiffe_ids:
      - spiffe://example.org/server
    trust_domains: []
    allow_any: false
```

If `workload_api_path` is empty, the SPIFFE Workload API client uses
`SPIFFE_ENDPOINT_SOCKET`.

## Server

```go
cfg, err := mtls.LoadConfigFile("config.yaml")
if err != nil {
    return err
}

srv, err := servicegrpc.ListenAndServe(ctx, *cfg, func(reg grpc.ServiceRegistrar) {
    servicev1.RegisterBackendRuntimeServer(reg, backend)
})
if err != nil {
    return err
}
defer srv.Close()
```

Inside a gRPC handler, the authenticated peer identity can be read from the
request context:

```go
id, err := mtls.PeerSPIFFEID(ctx)
if err != nil {
    return nil, err
}
```

## Client

```go
conn, err := servicegrpc.DialConfigFile(ctx, "config.yaml")
if err != nil {
    return err
}
defer conn.Close()

client := servicev1.NewBackendRuntimeClient(conn)
```
