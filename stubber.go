package melitest

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Stubber struct {
	Stubs  []*Stub
	Client Client
}

func (s *Stubber) Serve(t *testing.T) (Close func()) {
	t.Helper()
	if s == nil {
		return func() {}
	}
	sv := httptest.NewTLSServer(s.Router(t))
	s.Client.SetClient(StubberClient(sv))
	return sv.Close
}

func (s *Stubber) Router(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		stub := s.stubByURL(req.URL.RawPath)
		if stub == nil {
			t.Fatalf("couldnt match stub for %s url", req.URL.RawPath)
		}
		stub.intercept(t)
	})
}

func (s *Stubber) stubByURL(url string) *Stub {
	for _, stub := range s.Stubs {
		if stub.URL == url {
			return stub
		}
	}
	return nil
}

func StubberClient(sv *httptest.Server) http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
			return net.Dial(network, sv.Listener.Addr().String())
		},
	}
	return http.Client{Transport: transport}
}
