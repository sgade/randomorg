// Package randomorg is a Random.org API client as described at https://api.random.org/json-rpc/4.
// This is a third-party client. See https://github.com/sgade/randomorg.
// For any method documentation you should take a look at the official API documentation.
// An API key can be acquired here: https://api.random.org/dashboard.
package randomorg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Private constants
const (
	// The Random.org API request endpoint URL
	requestEndpoint = "https://api.random.org/json-rpc/4/invoke"
	// The time.Parse layout for the creationTime field the API returns, e.g. "2013-02-20 17:53:40Z"
	creationTimeLayout = time.RFC3339Nano
	// API Error template string
	errAPI = "api error code %v: %q"
)

// Constants describing error situations.
var (
	// ErrAPIKey is the error returned when an invalid API key was given.
	ErrAPIKey = errors.New("provide an api key")
	// ErrJSONFormat is the error returned when the response JSON had an unexpected format.
	ErrJSONFormat = errors.New("could not get key from given json")
	// ErrParamRange is returned when invalid parameter ranges where given to a method.
	// See the method API documentation for further details.
	ErrParamRange = errors.New("invalid parameter range")
	// ErrHTTPStatus is the error returned when the server's HTTP response status is unexpected.
	ErrHTTPStatus = errors.New("unexpected HTTP status")
	// ErrHTTPClient is the error returned when a nil http.Client was given.
	ErrHTTPClient = errors.New("provide an http client")
)

// A Random defines a Random.org API Client.
type Random struct {
	// the api key
	apiKey string
	// reusable http.Client
	client *http.Client
	// guards usage
	usageMutex sync.Mutex
	// usage cache
	usage *Usage
}

// NewRandom creates a new Random client with the given apiKey.
func NewRandom(apiKey string, client *http.Client) (*Random, error) {
	// check the api key
	if apiKey == "" {
		return nil, ErrAPIKey
	}
	if client == nil {
		return nil, ErrHTTPClient
	}

	random := Random{
		apiKey: apiKey,
		client: client,
	}

	return &random, nil
}

// baseParams embeds the API key required by every Random.org API method.
type baseParams struct {
	APIKey string `json:"apiKey"`
}

// jsonRPCRequest is the envelope for every Random.org JSON-RPC 2.0 request.
type jsonRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      string `json:"id"`
}

// jsonRPCResponse is the envelope for every Random.org JSON-RPC 2.0 response.
// R is the method-specific shape of a successful result.
type jsonRPCResponse[R any] struct {
	Result *R        `json:"result"`
	Error  *APIError `json:"error"`
}

// APIError describes an error returned by the Random.org API itself, as
// opposed to a transport or decoding failure. Use errors.As to retrieve it.
// See https://api.random.org/json-rpc/4/error-codes.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf(errAPI, e.Code, e.Message)
}

// invokeRequest sends method with params and decodes the JSON-RPC result into R.
func invokeRequest[R any](ctx context.Context, r *Random, method string, params any) (R, error) {
	var zero R

	// generate request UUID
	requestUUID, err := uuid.NewUUID()
	if err != nil {
		return zero, err
	}

	// build request body
	requestBody := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      requestUUID.String(),
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return zero, err
	}
	requestBodyReader := bytes.NewReader(requestBodyJSON)

	req, err := http.NewRequestWithContext(ctx, "POST", requestEndpoint, requestBodyReader)
	if err != nil {
		return zero, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return zero, ErrHTTPStatus
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, err
	}

	var responseBody jsonRPCResponse[R]
	if err := json.Unmarshal(body, &responseBody); err != nil {
		if len(body) > 0 {
			err = fmt.Errorf("%s: %w", body, err)
		}

		return zero, err
	}

	if responseBody.Error != nil {
		return zero, responseBody.Error
	}
	if responseBody.Result == nil {
		return zero, ErrJSONFormat
	}

	return *responseBody.Result, nil
}
