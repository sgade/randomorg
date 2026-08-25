package randomorg_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/sgade/randomorg/v3"
)

// usageFieldsEqual compares the exported fields of two Usage values.
// time.Time is compared with Equal rather than == to avoid the monotonic
// reading pitfalls of direct struct comparison.
func usageFieldsEqual(a, b randomorg.Usage) bool {
	return a.Status == b.Status &&
		a.CreationTime.Equal(b.CreationTime) &&
		a.BitsLeft == b.BitsLeft &&
		a.RequestsLeft == b.RequestsLeft &&
		a.TotalBits == b.TotalBits &&
		a.TotalRequests == b.TotalRequests
}

func TestUsage_CachesCompleteResponse(t *testing.T) {
	var requestCount int
	random := newTestRandom(t, func(*http.Request) (*http.Response, error) {
		requestCount++
		return jsonResponse(http.StatusOK, `{
			"jsonrpc": "2.0",
			"result": {
				"status": "running",
				"creationTime": "2013-02-01 17:53:40Z",
				"bitsLeft": 998556,
				"requestsLeft": 199,
				"totalBits": 1441,
				"totalRequests": 1
			},
			"id": "1"
		}`), nil
	})

	got, err := random.GetUsage(context.Background())
	if err != nil {
		t.Fatalf("GetUsage() error = %v", err)
	}
	if got.Status != "running" || got.BitsLeft != 998556 {
		t.Fatalf("GetUsage() = %+v, unexpected values", got)
	}
	if requestCount != 1 {
		t.Fatalf("requestCount = %d, want 1", requestCount)
	}

	// A second call should be served from cache, without another network call.
	got2, err := random.Usage(context.Background())
	if err != nil {
		t.Fatalf("Usage() error = %v", err)
	}
	if !usageFieldsEqual(got, got2) {
		t.Fatalf("Usage() = %+v, want cached %+v", got2, got)
	}
	if requestCount != 1 {
		t.Fatalf("requestCount after cached Usage() = %d, want still 1", requestCount)
	}
}

func TestUsage_IncompleteResponseIsNotCached(t *testing.T) {
	var requestCount int
	random := newTestRandom(t, func(*http.Request) (*http.Response, error) {
		requestCount++
		return jsonResponse(http.StatusOK, `{
			"jsonrpc": "2.0",
			"result": {"status": "running"},
			"id": "1"
		}`), nil
	})

	if _, err := random.GetUsage(context.Background()); err != nil {
		t.Fatalf("GetUsage() error = %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("requestCount = %d, want 1", requestCount)
	}

	if _, err := random.Usage(context.Background()); err != nil {
		t.Fatalf("Usage() error = %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("requestCount after Usage() = %d, want 2 (incomplete response should not be cached)", requestCount)
	}
}

func TestGetUsage_PropagatesNonFormatErrors(t *testing.T) {
	random := newTestRandom(t, func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusInternalServerError, ""), nil
	})

	_, err := random.GetUsage(context.Background())
	if !errors.Is(err, randomorg.ErrHTTPStatus) {
		t.Fatalf("err = %v, want %v", err, randomorg.ErrHTTPStatus)
	}
}

// TestUsage_ConcurrentAccess exercises the usage cache from many goroutines
// at once; run with -race to verify the mutex added around Random.usage
// actually prevents data races.
func TestUsage_ConcurrentAccess(t *testing.T) {
	random := newTestRandom(t, func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{
			"jsonrpc": "2.0",
			"result": {
				"status": "running",
				"creationTime": "2013-02-01 17:53:40Z",
				"bitsLeft": 1,
				"requestsLeft": 1,
				"totalBits": 1,
				"totalRequests": 1
			},
			"id": "1"
		}`), nil
	})

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := random.Usage(context.Background()); err != nil {
				t.Errorf("Usage() error = %v", err)
			}
		}()
	}
	wg.Wait()
}
