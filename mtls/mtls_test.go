package mtls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"testing"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

func TestPeerSPIFFEIDFromTLSInfo(t *testing.T) {
	spiffeURL := mustParseURL(t, "spiffe://example.org/client")
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: credentials.TLSInfo{SPIFFEID: spiffeURL},
	})

	id, err := PeerSPIFFEID(ctx)
	if err != nil {
		t.Fatal(err)
	}

	want := spiffeid.RequireFromString("spiffe://example.org/client")
	if id != want {
		t.Fatalf("unexpected SPIFFE ID: got %q want %q", id, want)
	}
}

func TestPeerSPIFFEIDFromPeerCertificate(t *testing.T) {
	spiffeURL := mustParseURL(t, "spiffe://example.org/client")
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					{URIs: []*url.URL{spiffeURL}},
				},
			},
		},
	})

	id, err := PeerSPIFFEID(ctx)
	if err != nil {
		t.Fatal(err)
	}

	want := spiffeid.RequireFromString("spiffe://example.org/client")
	if id != want {
		t.Fatalf("unexpected SPIFFE ID: got %q want %q", id, want)
	}
}

func TestPeerSPIFFEIDMissingPeer(t *testing.T) {
	if _, err := PeerSPIFFEID(context.Background()); err == nil {
		t.Fatal("expected missing peer error")
	}
}

func TestPeerSPIFFEIDWithoutCertificate(t *testing.T) {
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: credentials.TLSInfo{},
	})

	if _, err := PeerSPIFFEID(ctx); err == nil {
		t.Fatal("expected missing certificate error")
	}
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
