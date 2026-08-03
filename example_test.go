package randomorg_test

import (
	"fmt"
	"net/http"

	"github.com/sgade/randomorg"
)

const (
	apiKey = "YOUR API KEY"
)

func Example() {
	// create a new client
	_, err := randomorg.NewRandom(apiKey, &http.Client{})
	if err != nil {
		panic(err)
	}
	// call methods on the return value here
}

// Generate one int64 value from 0 to 10.
func ExampleRandom_GenerateIntegers() {
	random, err := randomorg.NewRandom(apiKey, &http.Client{})
	if err != nil {
		panic(err)
	}
	// generates a random value
	value, _ := random.GenerateIntegers(1, 0, 10)
	fmt.Printf("Random value: %v\n", value)
}
