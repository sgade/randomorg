package randomorg_test

import (
	"fmt"

	"github.com/sgade/randomorg"
)

const (
	apiKey = "YOUR API KEY"
)

func Example() {
	// create a new client
	_ = randomorg.NewRandom(apiKey)
	// call methods on the return value here
}

// Generate one int64 value from 0 to 10.
func ExampleRandom_GenerateIntegers() {
	random := randomorg.NewRandom(apiKey)
	// generates a random value
	value, _ := random.GenerateIntegers(1, 0, 10)
	fmt.Printf("Random value: %v\n", value)
}
