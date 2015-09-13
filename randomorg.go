package randomOrg

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// Private constants
const (
	// The Random.org API request endpoint URL
	requestEndpoint = "https://api.random.org/json-rpc/1/invoke"
)

// Constants describing error situations.
var (
	// Error with the API key
	ErrAPIKey     = errors.New("provide an api key")
	// Error with the response json
	ErrJsonFormat = errors.New("could not get key from given json")
	// Invalid parameter range
	ErrParamRage = errors.New("invalid parameter range")
)

// Random.org Client.
// For more information, see https://api.random.org/json-rpc/1/.
type RandomOrg struct {
	// the api key
	apiKey string
	// reusable http.Client
	client *http.Client
}

// NewRandomOrg creates a new RandomOrg client with the given apiKey.
func NewRandomOrg(apiKey string) *RandomOrg {
	// check the api key
	if apiKey == "" {
		panic(ErrAPIKey)
	}

	randomOrg := RandomOrg{
		apiKey: apiKey,
		client: &http.Client{},
	}

	return &randomOrg
}

// Get the json object with the given key from the given json object.
func (r *RandomOrg) jsonMap(json map[string]interface{}, key string) (map[string]interface{}, error) {
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

func (r *RandomOrg) invokeRequest(method string, params map[string]interface{}) (map[string]interface{}, error) {
	params["apiKey"] = r.apiKey

	requestUUID := uuid.NewUUID().String()
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

	return responseBody["result"].(map[string]interface{}), nil
}

// Generate n number of random integers in the range from min to max.
func (r *RandomOrg) GenerateIntegers(n, min, max int64) ([]int64, error) {
	params := map[string]interface{}{
		"n":   n,
		"min": min,
		"max": max,
	}

	result, err := r.invokeRequest("generateIntegers", params)
	if err != nil {
		return nil, err
	}
	random, err := r.jsonMap(result, "random")
	if err != nil {
		return nil, err
	}
	data := random["data"].([]interface{})

	ints := make([]int64, len(data))
	for i, dataItem := range data {
		f := dataItem.(float64)
		ints[i] = int64(f)
	}

	return ints, nil
}
