package httpstub

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Stub is any single request stub
type Stub struct {
	URL string

	Status int
	Body   interface{}

	Config  StubConfig
	Receive Receive
}

type Receive struct {
	Body   []byte
	Params url.Values
}

type StubConfig struct {
	DontAssertReceive bool
}

func (s *Stub) intercept(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if !s.Config.DontAssertReceive {
			s.assertReceive(t, req)
		}
		res.WriteHeader(s.Status)
		err := json.NewEncoder(res).Encode(s.Body)
		if err != nil {
			t.Fatalf("couldn't intercept due body marshalling error: %v", err)
		}
	})
}

func (s *Stub) assertReceive(t *testing.T, req *http.Request) {
	t.Helper()
	assertReceivedBody(t, req, s.Receive.Body)
	if s.Receive.Params == nil {
		s.Receive.Params = url.Values{}
	}
	assertReceivedParams(t, req, s.Receive.Params)
}

func assertReceivedBody(t *testing.T, req *http.Request, assertion []byte) {
	t.Helper()
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil && err != io.EOF {
		t.Fatalf("couldn't read request body: %v", err)
	}
	if diff := cmp.Diff(string(assertion), string(body)); diff != "" {
		t.Errorf("body received mismatch (-want +got): %s", diff)
	}
}

func assertReceivedParams(t *testing.T, req *http.Request, assertion url.Values) {
	t.Helper()
	got := req.URL.Query()

	if diff := cmp.Diff(assertion, got); diff != "" {
		t.Errorf("params received mismatch (-want +got): %s", diff)
	}
}
