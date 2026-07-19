package mtls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"slices"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type Source struct {
	*workloadapi.X509Source
}

func NewSource(ctx context.Context, workloadAPIPath string) (*Source, error) {
	var opts []workloadapi.X509SourceOption
	if workloadAPIPath != "" {
		opts = append(opts, workloadapi.WithClientOptions(workloadapi.WithAddr(workloadAPIPath)))
	}

	source, err := workloadapi.NewX509Source(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create SPIFFE X.509 source: %w", err)
	}

	return &Source{X509Source: source}, nil
}

func ServerTLSConfig(source *Source, cfg ServerConfig) (*tls.Config, error) {
	if source == nil || source.X509Source == nil {
		return nil, fmt.Errorf("missing SPIFFE X.509 source")
	}

	authorizer, err := NewAuthorizer(cfg.AuthorizedPeers)
	if err != nil {
		return nil, err
	}

	return tlsconfig.MTLSServerConfig(source.X509Source, source.X509Source, authorizer), nil
}

func ClientTLSConfig(source *Source, authorizedPeers AuthorizationConfig) (*tls.Config, error) {
	if source == nil || source.X509Source == nil {
		return nil, fmt.Errorf("missing SPIFFE X.509 source")
	}

	authorizer, err := NewAuthorizer(authorizedPeers)
	if err != nil {
		return nil, err
	}

	return tlsconfig.MTLSClientConfig(source.X509Source, source.X509Source, authorizer), nil
}

func PeerSPIFFEID(ctx context.Context) (spiffeid.ID, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return spiffeid.ID{}, fmt.Errorf("missing gRPC peer in context")
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return spiffeid.ID{}, fmt.Errorf("gRPC peer auth info is %T, expected credentials.TLSInfo", p.AuthInfo)
	}

	if tlsInfo.SPIFFEID != nil {
		id, err := spiffeid.FromURI(tlsInfo.SPIFFEID)
		if err != nil {
			return spiffeid.ID{}, fmt.Errorf("invalid peer SPIFFE ID: %w", err)
		}
		return id, nil
	}

	if len(tlsInfo.State.PeerCertificates) == 0 {
		return spiffeid.ID{}, fmt.Errorf("peer TLS state contains no certificates")
	}

	id, err := x509svid.IDFromCert(tlsInfo.State.PeerCertificates[0])
	if err != nil {
		return spiffeid.ID{}, fmt.Errorf("extract peer SPIFFE ID from certificate: %w", err)
	}
	return id, nil
}

func NewAuthorizer(cfg AuthorizationConfig) (tlsconfig.Authorizer, error) {
	logger := slog.Default()
	if cfg.Logger != nil {
		logger = cfg.Logger
	}
	if cfg.AllowAny && len(cfg.SPIFFEIDs) == 0 && len(cfg.TrustDomains) == 0 {
		return tlsconfig.AuthorizeAny(), nil
	}

	ids := make([]spiffeid.ID, 0, len(cfg.SPIFFEIDs))
	for _, raw := range cfg.SPIFFEIDs {
		id, err := spiffeid.FromString(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid SPIFFE ID %q: %w", raw, err)
		}
		ids = append(ids, id)
	}

	trustDomains := make([]spiffeid.TrustDomain, 0, len(cfg.TrustDomains))
	for _, raw := range cfg.TrustDomains {
		td, err := spiffeid.TrustDomainFromString(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid SPIFFE trust domain %q: %w", raw, err)
		}
		trustDomains = append(trustDomains, td)
	}

	if len(ids) == 0 && len(trustDomains) == 0 {
		return nil, fmt.Errorf("no authorized SPIFFE IDs or trust domains configured")
	}

	return func(actual spiffeid.ID, _ [][]*x509.Certificate) error {
		if slices.Contains(ids, actual) {
			return nil
		}

		if slices.Contains(trustDomains, actual.TrustDomain()) {
			return nil
		}

		logger.Debug("SPIFFE ID authorization failed", "id", actual.String())
		return fmt.Errorf("SPIFFE ID %q is not authorized", actual.String())
	}, nil
}
