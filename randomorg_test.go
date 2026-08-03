package randomorg_test

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"

	"github.com/sgade/randomorg"
)

func TestNewRandom(t *testing.T) {
	t.Run("empty api key", func(t *testing.T) {
		random, err := randomorg.NewRandom("", &http.Client{})
		if !errors.Is(err, randomorg.ErrAPIKey) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrAPIKey)
		}
		if random != nil {
			t.Fatalf("random = %v, want nil", random)
		}
	})

	t.Run("valid api key", func(t *testing.T) {
		random, err := randomorg.NewRandom(testAPIKey, &http.Client{})
		if err != nil {
			t.Fatalf("NewRandom() error = %v", err)
		}
		if random == nil {
			t.Fatal("random = nil, want non-nil")
		}
	})
}

func TestRequest_SendsExpectedRequest(t *testing.T) {
	var gotReq map[string]any
	random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Errorf("req.Method = %q, want %q", req.Method, http.MethodPost)
		}
		if ct := req.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		if accept := req.Header.Get("Accept"); accept != "application/json" {
			t.Errorf("Accept = %q, want application/json", accept)
		}

		gotReq = decodeRequestBody(t, req)

		return jsonResponse(http.StatusOK, `{
			"jsonrpc": "2.0",
			"result": {
				"random": {"data": [7]},
				"bitsLeft": 1,
				"requestsLeft": 1
			},
			"id": "1"
		}`), nil
	})

	values, err := random.GenerateIntegers(context.Background(), 1, 0, 10)
	if err != nil {
		t.Fatalf("GenerateIntegers() error = %v", err)
	}
	if want := []int64{7}; !slices.Equal(values, want) {
		t.Fatalf("GenerateIntegers() = %v, want %v", values, want)
	}

	if gotReq["method"] != "generateIntegers" {
		t.Errorf("request method = %v, want generateIntegers", gotReq["method"])
	}
	params, ok := gotReq["params"].(map[string]any)
	if !ok {
		t.Fatalf("request params missing or wrong type: %v", gotReq["params"])
	}
	if params["apiKey"] != testAPIKey {
		t.Errorf("request apiKey = %v, want %v", params["apiKey"], testAPIKey)
	}
}

func TestRequest_HTTPStatusError(t *testing.T) {
	random := newTestRandom(t, func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusInternalServerError, ""), nil
	})

	_, err := random.GenerateIntegers(context.Background(), 1, 0, 10)
	if !errors.Is(err, randomorg.ErrHTTPStatus) {
		t.Fatalf("err = %v, want %v", err, randomorg.ErrHTTPStatus)
	}
}

func TestRequest_MalformedJSONBody(t *testing.T) {
	random := newTestRandom(t, func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, "not json"), nil
	})

	_, err := random.GenerateIntegers(context.Background(), 1, 0, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "not json" {
		t.Fatalf("err = %q, want %q", err.Error(), "not json")
	}
}

func TestRequest_APIError(t *testing.T) {
	random := newTestRandom(t, func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{
			"jsonrpc": "2.0",
			"error": {"code": 401, "message": "Invalid API key"},
			"id": "1"
		}`), nil
	})

	_, err := random.GenerateIntegers(context.Background(), 1, 0, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	const want = `API Error Code 401: "Invalid API key".`
	if err.Error() != want {
		t.Fatalf("err = %q, want %q", err.Error(), want)
	}
}

func TestRequest_MissingRandomData(t *testing.T) {
	random := newTestRandom(t, func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{
			"jsonrpc": "2.0",
			"result": {"bitsLeft": 1},
			"id": "1"
		}`), nil
	})

	_, err := random.GenerateIntegers(context.Background(), 1, 0, 10)
	if !errors.Is(err, randomorg.ErrJSONFormat) {
		t.Fatalf("err = %v, want %v", err, randomorg.ErrJSONFormat)
	}
}
