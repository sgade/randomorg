// A Random.org API client as described at https://api.random.org/json-rpc/1/.
// This is a third-party client. See https://github.com/sgade/randomorg.
package randomorg

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Private constants
const (
	// The Random.org API request endpoint URL
	requestEndpoint = "https://api.random.org/json-rpc/1/invoke"
	// Example time format for ISO 8601
	iso8601Example = time.RFC3339Nano //"2013-02-20 17:53:40Z"
)

// Constants describing error situations.
var (
	// Error with the API key
	ErrAPIKey = errors.New("provide an api key")
	// Error with the response json
	ErrJsonFormat = errors.New("could not get key from given json")
	// Invalid parameter range
	ErrParamRange = errors.New("invalid parameter range")
	// API Error template string
	errAPI = "API Error Code %v: %q."
)

// Random.org Client.
// For more information, see https://api.random.org/json-rpc/1/.
type Random struct {
	// the api key
	apiKey string
	// reusable http.Client
	client *http.Client
	// usage cache
	usage *Usage
}

// NewRandom creates a new Random client with the given apiKey.
func NewRandom(apiKey string) *Random {
	// check the api key
	if apiKey == "" {
		panic(ErrAPIKey)
	}

	random := Random{
		apiKey: apiKey,
		client: &http.Client{},
	}

	return &random
}

// Get the json object with the given key from the given json object.
func (r *Random) jsonMap(json map[string]interface{}, key string) (map[string]interface{}, error) {
	value := json[key]
	if value == nil {
		return nil, ErrJsonFormat
	}

	newMap, ok := value.(map[string]interface{})
	if !ok {
		return nil, ErrJsonFormat
	}

	return newMap, nil
}

func (r *Random) invokeRequest(method string, params map[string]interface{}) (map[string]interface{}, error) {
	// always append api key
	params["apiKey"] = r.apiKey

	// generate request UUID
	requestUUID := uuid.NewUUID().String()
	// build request body
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      requestUUID,
	}
	requestBodyJson, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	requestBodyReader := bytes.NewReader(requestBodyJson)

	req, err := http.NewRequest("POST", requestEndpoint, requestBodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json-rpc")
	req.Header.Add("Accept", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var responseBody map[string]interface{} = make(map[string]interface{})
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
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
func (r *Random) requestCommand(method string, params map[string]interface{}) ([]interface{}, error) {
	result, err := r.invokeRequest(method, params)
	if err != nil {
		return nil, err
	}

	r.parseAndSaveUsage(result)

	random, err := r.jsonMap(result, "random")
	if err != nil {
		return nil, err
	}

	data := random["data"].([]interface{})

	return data, nil
}
