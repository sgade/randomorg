/*
 * Copyright 2015 Sören Gade
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package randomorg is a Random.org API client as described at https://api.random.org/json-rpc/2.
// This is a third-party client. See https://github.com/sgade/randomorg.
// For any method documentation you should take a look at the official API documentation.
// An API key can be acquired here: https://api.random.org/dashboard.
package randomorg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Private constants
const (
	// The Random.org API request endpoint URL
	requestEndpoint = "https://api.random.org/json-rpc/2/invoke"
	// Example time format for ISO 8601
	iso8601Example = time.RFC3339Nano //"2013-02-20 17:53:40Z"
	// API Error template string
	errAPI = "API Error Code %v: %q."
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
)

// A Random defines a Random.org API Client.
// For more information, see https://api.random.org/json-rpc/2.
type Random struct {
	// the api key
	apiKey string
	// reusable http.Client
	client *http.Client
	// usage cache
	usage *Usage
}

// NewRandom creates a new Random client with the given apiKey.
func NewRandom(apiKey string, client *http.Client) *Random {
	// check the api key
	if apiKey == "" {
		panic(ErrAPIKey)
	}

	random := Random{
		apiKey: apiKey,
		client: client,
	}

	return &random
}

// Get the json object with the given key from the given json object.
func (r *Random) jsonMap(json map[string]any, key string) (map[string]any, error) {
	value := json[key]
	if value == nil {
		return nil, ErrJSONFormat
	}

	newMap, ok := value.(map[string]any)
	if !ok {
		return nil, ErrJSONFormat
	}

	return newMap, nil
}

func (r *Random) invokeRequest(method string, params map[string]any) (map[string]any, error) {
	// always append api key
	params["apiKey"] = r.apiKey

	// generate request UUID
	requestUUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	// build request body
	requestBody := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      requestUUID.String(),
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	requestBodyReader := bytes.NewReader(requestBodyJSON)

	req, err := http.NewRequest("POST", requestEndpoint, requestBodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	responseBody := make(map[string]any)
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		if len(body) > 0 {
			err = errors.New(string(body))
		}

		return nil, err
	}

	result, err := r.jsonMap(responseBody, "result")
	if err != nil {
		error, err := r.jsonMap(responseBody, "error")
		if err != nil {
			return nil, err
		}

		errorCode, _ := error["code"]
		errorMessage, _ := error["message"]
		err = fmt.Errorf(errAPI, errorCode, errorMessage)
		return nil, err
	}

	return result, nil
}

// requestCommand invokes the request and parses all information down to the requested data block.
func (r *Random) requestCommand(method string, params map[string]any) ([]any, error) {
	result, err := r.invokeRequest(method, params)
	if err != nil {
		return nil, err
	}

	r.parseAndSaveUsage(result)

	random, err := r.jsonMap(result, "random")
	if err != nil {
		return nil, err
	}

	data := random["data"].([]any)

	return data, nil
}
