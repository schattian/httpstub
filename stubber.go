package httpstub

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
	Config StubberConfig
}

type StubberConfig struct {
	DontAssertUnstubbed bool
}

func (s *Stubber) Serve(t *testing.T) (Close func()) {
	t.Helper()
	if s == nil {
		return func() {}
	}
	sv := httptest.NewTLSServer(s.Router(t))
	s.Client.SetClient(stubberClient(sv))
	return sv.Close
}

func (s *Stubber) Router(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		stub := s.stubByURL(req.URL.Path)
		if !s.Config.DontAssertUnstubbed {
			if stub == nil {
				t.Fatalf("couldnt match stub for %s url", req.URL.Path)
			}
		}
		stub.intercept(t).ServeHTTP(res, req)
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

func stubberClient(sv *httptest.Server) http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
			return net.Dial(network, sv.Listener.Addr().String())
		},
	}
	return http.Client{Transport: transport}
}
