# Randomorg

*Golang Random.org API Client*

[![GoDoc](https://godoc.org/github.com/sgade/randomorg?status.svg)](https://godoc.org/github.com/sgade/randomorg)

A client for the [Random.org JSON-RPC API](https://api.random.org/json-rpc/4), which
serves true random numbers generated from atmospheric noise. An API key is required and
can be acquired at https://api.random.org/dashboard.

## Install

```sh
go get github.com/sgade/randomorg
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/sgade/randomorg"
)

func main() {
	random, err := randomorg.NewRandom("YOUR API KEY", &http.Client{})
	if err != nil {
		log.Fatal(err)
	}

	values, err := random.GenerateIntegers(context.Background(), 1, 0, 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(values)
}
```

See the [GoDoc](https://godoc.org/github.com/sgade/randomorg) for the full API, or
`example_test.go` for more runnable examples.
