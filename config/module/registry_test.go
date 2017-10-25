package module

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform/svchost/disco"
)

// Return a transport to use for this test server.
// This not only loads the tls.Config from the test server for proper cert
// validation, but also inserts a Dialer that resolves localhost and
// example.com to 127.0.0.1 with the correct port, since 127.0.0.1 on its own
// isn't a valid registry hostname.
// TODO: cert validation not working here, so we use don't verify for now.
func mockTransport(server *httptest.Server) *http.Transport {
	u, _ := url.Parse(server.URL)
	_, port, _ := net.SplitHostPort(u.Host)

	transport := cleanhttp.DefaultTransport()
	transport.TLSClientConfig = server.TLS
	transport.TLSClientConfig.InsecureSkipVerify = true
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _, _ := net.SplitHostPort(addr)
		switch host {
		case "example.com", "localhost", "localhost.localdomain":
			addr = "127.0.0.1"
			if port != "" {
				addr += ":" + port
			}
		}
		return (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext(ctx, network, addr)
	}
	return transport
}

func TestMockDiscovery(t *testing.T) {
	server := mockTLSRegistry()
	defer server.Close()

	regDisco := disco.NewDisco()
	regDisco.Transport = mockTransport(server)

	regURL := regDisco.DiscoverServiceURL("example.com", serviceID)

	if regURL == nil {
		t.Fatal("no registry service discovered")
	}

	if regURL.Host != "example.com" {
		t.Fatal("expected registry host example.com, got:", regURL.Host)
	}
}
