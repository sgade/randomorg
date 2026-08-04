package randomorg_test

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/sgade/randomorg"
)

func TestGenerateIntegers(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		cases := []struct {
			name     string
			n        int
			min, max int64
		}{
			{"n too small", 0, 0, 10},
			{"n too large", 10001, 0, 10},
			{"min too small", 1, -1e9 - 1, 10},
			{"max too large", 1, 0, 1e9 + 1},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				random := newTestRandom(t, failOnRequest(t))
				_, err := random.GenerateIntegers(context.Background(), tc.n, tc.min, tc.max)
				if !errors.Is(err, randomorg.ErrParamRange) {
					t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
				}
			})
		}
	})

	t.Run("generates values", func(t *testing.T) {
		var gotMethod string
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotMethod, _ = decodeRequestBody(t, req)["method"].(string)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": [1, 5, 10]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		got, err := random.GenerateIntegers(context.Background(), 3, 0, 10)
		if err != nil {
			t.Fatalf("GenerateIntegers() error = %v", err)
		}
		if want := []int64{1, 5, 10}; !slices.Equal(got, want) {
			t.Fatalf("GenerateIntegers() = %v, want %v", got, want)
		}
		if gotMethod != "generateIntegers" {
			t.Errorf("method = %q, want generateIntegers", gotMethod)
		}
	})
}

func TestGenerateDecimalFractions(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		cases := []struct {
			name          string
			n             int
			decimalPlaces int
		}{
			{"n too small", 0, 2},
			{"n too large", 10001, 2},
			{"decimalPlaces too small", 2, 0},
			{"decimalPlaces too large", 2, 15},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				random := newTestRandom(t, failOnRequest(t))
				_, err := random.GenerateDecimalFractions(context.Background(), tc.n, tc.decimalPlaces)
				if !errors.Is(err, randomorg.ErrParamRange) {
					t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
				}
			})
		}
	})

	t.Run("generates values", func(t *testing.T) {
		var gotMethod string
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotMethod, _ = decodeRequestBody(t, req)["method"].(string)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": [0.5, 0.25]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		got, err := random.GenerateDecimalFractions(context.Background(), 2, 2)
		if err != nil {
			t.Fatalf("GenerateDecimalFractions() error = %v", err)
		}
		if want := []float64{0.5, 0.25}; !slices.Equal(got, want) {
			t.Fatalf("GenerateDecimalFractions() = %v, want %v", got, want)
		}
		if gotMethod != "generateDecimalFractions" {
			t.Errorf("method = %q, want generateDecimalFractions", gotMethod)
		}
	})
}

func TestGenerateGaussians(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		cases := []struct {
			name                                          string
			n, mean, standardDeviation, significantDigits int
		}{
			{"n too small", 0, 0, 1, 4},
			{"n too large", 10001, 0, 1, 4},
			{"mean too small", 2, -1e6 - 1, 1, 4},
			{"mean too large", 2, 1e6 + 1, 1, 4},
			{"standardDeviation too small", 2, 0, -1e6 - 1, 4},
			{"standardDeviation too large", 2, 0, 1e6 + 1, 4},
			{"significantDigits too small", 2, 0, 1, 1},
			{"significantDigits too large", 2, 0, 1, 15},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				random := newTestRandom(t, failOnRequest(t))
				_, err := random.GenerateGaussians(context.Background(), tc.n, tc.mean, tc.standardDeviation, tc.significantDigits)
				if !errors.Is(err, randomorg.ErrParamRange) {
					t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
				}
			})
		}
	})

	t.Run("generates values", func(t *testing.T) {
		var gotMethod string
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotMethod, _ = decodeRequestBody(t, req)["method"].(string)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": [0.1, -0.2]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		got, err := random.GenerateGaussians(context.Background(), 2, 0, 1, 4)
		if err != nil {
			t.Fatalf("GenerateGaussians() error = %v", err)
		}
		if want := []float64{0.1, -0.2}; !slices.Equal(got, want) {
			t.Fatalf("GenerateGaussians() = %v, want %v", got, want)
		}
		if gotMethod != "generateGaussians" {
			t.Errorf("method = %q, want generateGaussians", gotMethod)
		}
	})
}

func TestGenerateStrings(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		cases := []struct {
			name       string
			n, length  int
			characters string
		}{
			{"n too small", 0, 5, "abc"},
			{"n too large", 10001, 5, "abc"},
			{"length too small", 2, 0, "abc"},
			{"length too large", 2, 33, "abc"},
			{"characters empty", 2, 5, ""},
			{"characters too long", 2, 5, strings.Repeat("a", 129)},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				random := newTestRandom(t, failOnRequest(t))
				_, err := random.GenerateStrings(context.Background(), tc.n, tc.length, tc.characters)
				if !errors.Is(err, randomorg.ErrParamRange) {
					t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
				}
			})
		}
	})

	t.Run("generates values", func(t *testing.T) {
		var gotMethod string
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotMethod, _ = decodeRequestBody(t, req)["method"].(string)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": ["abc", "cab"]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		got, err := random.GenerateStrings(context.Background(), 2, 3, "abc")
		if err != nil {
			t.Fatalf("GenerateStrings() error = %v", err)
		}
		if want := []string{"abc", "cab"}; !slices.Equal(got, want) {
			t.Fatalf("GenerateStrings() = %v, want %v", got, want)
		}
		if gotMethod != "generateStrings" {
			t.Errorf("method = %q, want generateStrings", gotMethod)
		}
	})
}

func TestGenerateUUIDs(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		cases := []struct {
			name string
			n    int
		}{
			{"n too small", 0},
			{"n too large", 1001},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				random := newTestRandom(t, failOnRequest(t))
				_, err := random.GenerateUUIDs(context.Background(), tc.n)
				if !errors.Is(err, randomorg.ErrParamRange) {
					t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
				}
			})
		}
	})

	t.Run("generates values", func(t *testing.T) {
		var gotMethod string
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotMethod, _ = decodeRequestBody(t, req)["method"].(string)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": ["11111111-1111-4111-8111-111111111111"]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		got, err := random.GenerateUUIDs(context.Background(), 1)
		if err != nil {
			t.Fatalf("GenerateUUIDs() error = %v", err)
		}
		if want := []string{"11111111-1111-4111-8111-111111111111"}; !slices.Equal(got, want) {
			t.Fatalf("GenerateUUIDs() = %v, want %v", got, want)
		}
		if gotMethod != "generateUUIDs" {
			t.Errorf("method = %q, want generateUUIDs", gotMethod)
		}
	})
}

func TestGenerateBlobs(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		cases := []struct {
			name    string
			n, size int
		}{
			{"n too small", 0, 8},
			{"n too large", 101, 8},
			{"size too small", 1, 0},
			{"size too large", 1, 1048584},
			{"size not multiple of 8", 1, 7},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				random := newTestRandom(t, failOnRequest(t))
				_, err := random.GenerateBlobs(context.Background(), tc.n, tc.size)
				if !errors.Is(err, randomorg.ErrParamRange) {
					t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
				}
			})
		}
	})

	t.Run("generates values", func(t *testing.T) {
		var gotMethod string
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotMethod, _ = decodeRequestBody(t, req)["method"].(string)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": ["ZGVhZGJlZWY="]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		got, err := random.GenerateBlobs(context.Background(), 1, 8)
		if err != nil {
			t.Fatalf("GenerateBlobs() error = %v", err)
		}
		if want := []string{"ZGVhZGJlZWY="}; !slices.Equal(got, want) {
			t.Fatalf("GenerateBlobs() = %v, want %v", got, want)
		}
		if gotMethod != "generateBlobs" {
			t.Errorf("method = %q, want generateBlobs", gotMethod)
		}
	})
}
