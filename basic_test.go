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

	t.Run("options are sent", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": [7]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		_, err := random.GenerateIntegers(context.Background(), 1, 0, 10, randomorg.GenerateIntegersOptions{
			Replacement:               randomorg.Bool(false),
			PregeneratedRandomization: randomorg.PregeneratedRandomizationByDate("2021-01-01"),
		})
		if err != nil {
			t.Fatalf("GenerateIntegers() error = %v", err)
		}

		if gotParams["replacement"] != false {
			t.Errorf("params[replacement] = %v, want false", gotParams["replacement"])
		}
		pr, ok := gotParams["pregeneratedRandomization"].(map[string]any)
		if !ok {
			t.Fatalf("params[pregeneratedRandomization] missing or wrong type: %v", gotParams["pregeneratedRandomization"])
		}
		if pr["date"] != "2021-01-01" {
			t.Errorf("params[pregeneratedRandomization][date] = %v, want 2021-01-01", pr["date"])
		}
	})

	t.Run("omitted options are not sent", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": [7]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		if _, err := random.GenerateIntegers(context.Background(), 1, 0, 10); err != nil {
			t.Fatalf("GenerateIntegers() error = %v", err)
		}

		if _, ok := gotParams["replacement"]; ok {
			t.Errorf("params[replacement] present = %v, want absent", gotParams["replacement"])
		}
		if _, ok := gotParams["pregeneratedRandomization"]; ok {
			t.Errorf("params[pregeneratedRandomization] present = %v, want absent", gotParams["pregeneratedRandomization"])
		}
	})

	t.Run("pregeneratedRandomization id too long", func(t *testing.T) {
		random := newTestRandom(t, failOnRequest(t))
		_, err := random.GenerateIntegers(context.Background(), 1, 0, 10, randomorg.GenerateIntegersOptions{
			PregeneratedRandomization: randomorg.PregeneratedRandomizationByID(strings.Repeat("a", 65)),
		})
		if !errors.Is(err, randomorg.ErrParamRange) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
		}
	})
}

func TestGenerateIntegerSequences(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		cases := []struct {
			name     string
			n        int
			length   []int
			min, max []int64
		}{
			{"n too small", 0, []int{1}, []int64{0}, []int64{10}},
			{"n too large", 1001, make([]int, 1001), make([]int64, 1001), make([]int64, 1001)},
			{"length slice wrong size", 2, []int{1}, []int64{0, 0}, []int64{10, 10}},
			{"min slice wrong size", 2, []int{1, 1}, []int64{0}, []int64{10, 10}},
			{"max slice wrong size", 2, []int{1, 1}, []int64{0, 0}, []int64{10}},
			{"length element too small", 1, []int{0}, []int64{0}, []int64{10}},
			{"length element too large", 1, []int{10_001}, []int64{0}, []int64{10}},
			{"min element too small", 1, []int{1}, []int64{-1e9 - 1}, []int64{10}},
			{"max element too large", 1, []int{1}, []int64{0}, []int64{1e9 + 1}},
			{"length sum too large", 2, []int{6_000, 6_000}, []int64{0, 0}, []int64{10, 10}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				random := newTestRandom(t, failOnRequest(t))
				_, err := random.GenerateIntegerSequences(context.Background(), tc.n, tc.length, tc.min, tc.max)
				if !errors.Is(err, randomorg.ErrParamRange) {
					t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
				}
			})
		}
	})

	t.Run("mismatched replacement length", func(t *testing.T) {
		random := newTestRandom(t, failOnRequest(t))
		_, err := random.GenerateIntegerSequences(context.Background(), 2, []int{1, 1}, []int64{0, 0}, []int64{10, 10}, randomorg.GenerateIntegerSequencesOptions{
			Replacement: []bool{true},
		})
		if !errors.Is(err, randomorg.ErrParamRange) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
		}
	})

	// Example from https://api.random.org/json-rpc/4/basic#generateIntegerSequences
	t.Run("generates values (docs example)", func(t *testing.T) {
		var gotReq map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotReq = decodeRequestBody(t, req)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"random": {
						"data": [[28, 31, 41, 65, 42], [14]],
						"completionTime": "2018-01-29 17:34:46Z"
					},
					"bitsUsed": 36,
					"bitsLeft": 833949,
					"requestsLeft": 199598,
					"advisoryDelay": 200
				},
				"id": "45673"
			}`), nil
		})

		got, err := random.GenerateIntegerSequences(context.Background(), 2, []int{5, 1}, []int64{1, 1}, []int64{69, 26}, randomorg.GenerateIntegerSequencesOptions{
			Replacement: []bool{false, false},
		})
		if err != nil {
			t.Fatalf("GenerateIntegerSequences() error = %v", err)
		}
		want := [][]int64{{28, 31, 41, 65, 42}, {14}}
		if len(got) != len(want) {
			t.Fatalf("GenerateIntegerSequences() = %v, want %v", got, want)
		}
		for i := range want {
			if !slices.Equal(got[i], want[i]) {
				t.Fatalf("GenerateIntegerSequences()[%d] = %v, want %v", i, got[i], want[i])
			}
		}

		if gotReq["method"] != "generateIntegerSequences" {
			t.Errorf("method = %q, want generateIntegerSequences", gotReq["method"])
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

	t.Run("replacement option is sent", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": [0.5]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		_, err := random.GenerateDecimalFractions(context.Background(), 1, 2, randomorg.GenerateDecimalFractionsOptions{
			Replacement: randomorg.Bool(false),
		})
		if err != nil {
			t.Fatalf("GenerateDecimalFractions() error = %v", err)
		}
		if gotParams["replacement"] != false {
			t.Errorf("params[replacement] = %v, want false", gotParams["replacement"])
		}
	})
}

func TestGenerateGaussians(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		cases := []struct {
			name                    string
			n                       int
			mean, standardDeviation float64
			significantDigits       int
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

		got, err := random.GenerateGaussians(context.Background(), 2, 0.5, 1.5, 4)
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

	t.Run("pregeneratedRandomization option is sent", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": [0.1]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		_, err := random.GenerateGaussians(context.Background(), 1, 0, 1, 4, randomorg.GenerateGaussiansOptions{
			PregeneratedRandomization: randomorg.PregeneratedRandomizationByID("my-persistent-id"),
		})
		if err != nil {
			t.Fatalf("GenerateGaussians() error = %v", err)
		}
		pr, ok := gotParams["pregeneratedRandomization"].(map[string]any)
		if !ok {
			t.Fatalf("params[pregeneratedRandomization] missing or wrong type: %v", gotParams["pregeneratedRandomization"])
		}
		if pr["id"] != "my-persistent-id" {
			t.Errorf("params[pregeneratedRandomization][id] = %v, want my-persistent-id", pr["id"])
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

	t.Run("replacement option is sent", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": ["abc"]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		_, err := random.GenerateStrings(context.Background(), 1, 3, "abc", randomorg.GenerateStringsOptions{
			Replacement: randomorg.Bool(false),
		})
		if err != nil {
			t.Fatalf("GenerateStrings() error = %v", err)
		}
		if gotParams["replacement"] != false {
			t.Errorf("params[replacement] = %v, want false", gotParams["replacement"])
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

	t.Run("pregeneratedRandomization option is sent", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": ["11111111-1111-4111-8111-111111111111"]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		_, err := random.GenerateUUIDs(context.Background(), 1, randomorg.GenerateUUIDsOptions{
			PregeneratedRandomization: randomorg.PregeneratedRandomizationByDate("2020-06-15"),
		})
		if err != nil {
			t.Fatalf("GenerateUUIDs() error = %v", err)
		}
		pr, ok := gotParams["pregeneratedRandomization"].(map[string]any)
		if !ok {
			t.Fatalf("params[pregeneratedRandomization] missing or wrong type: %v", gotParams["pregeneratedRandomization"])
		}
		if pr["date"] != "2020-06-15" {
			t.Errorf("params[pregeneratedRandomization][date] = %v, want 2020-06-15", pr["date"])
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
			{"aggregate size too large", 2, 1_048_576},
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

	t.Run("format option is sent", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {"random": {"data": ["deadbeef"]}, "bitsLeft": 1, "requestsLeft": 1},
				"id": "1"
			}`), nil
		})

		_, err := random.GenerateBlobs(context.Background(), 1, 8, randomorg.GenerateBlobsOptions{Format: randomorg.BlobFormatHex})
		if err != nil {
			t.Fatalf("GenerateBlobs() error = %v", err)
		}
		if gotParams["format"] != "hex" {
			t.Errorf("params[format] = %v, want hex", gotParams["format"])
		}
	})

	t.Run("invalid format is rejected", func(t *testing.T) {
		random := newTestRandom(t, failOnRequest(t))
		_, err := random.GenerateBlobs(context.Background(), 1, 8, randomorg.GenerateBlobsOptions{Format: "bogus"})
		if !errors.Is(err, randomorg.ErrParamRange) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
		}
	})
}
