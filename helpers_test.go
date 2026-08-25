package randomorg_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/sgade/randomorg/v3"
)

// testAPIKey is a placeholder used throughout the tests. It is never sent
// over the network, since every test client below uses a mocked transport.
const testAPIKey = "test-api-key-for-unit-tests-only"

// roundTripFunc adapts a function to http.RoundTripper so tests can stub
// HTTP responses without performing any real network I/O.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// newTestRandom returns a Random client whose requests are served entirely
// in-process by handler; no request ever reaches the network.
func newTestRandom(t *testing.T, handler func(req *http.Request) (*http.Response, error)) *randomorg.Random {
	t.Helper()

	client := &http.Client{Transport: roundTripFunc(handler)}

	random, err := randomorg.NewRandom(testAPIKey, client)
	if err != nil {
		t.Fatalf("NewRandom() error = %v", err)
	}

	return random
}

// failOnRequest returns a handler that fails the test if it is ever invoked.
// It's used to assert that invalid parameters are rejected before any
// request is sent.
func failOnRequest(t *testing.T) func(*http.Request) (*http.Response, error) {
	return func(*http.Request) (*http.Response, error) {
		t.Fatal("unexpected network call")
		return nil, nil
	}
}

// jsonResponse builds a canned *http.Response carrying the given status code
// and JSON body.
func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}

// decodeRequestBody parses the JSON-RPC request body of req for assertions.
func decodeRequestBody(t *testing.T, req *http.Request) map[string]any {
	t.Helper()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("reading request body: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decoding request body: %v", err)
	}

	return decoded
}
