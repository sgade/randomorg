# Randomorg

*Golang Random.org API Client*

[![Go Reference](https://pkg.go.dev/badge/github.com/sgade/randomorg/v3.svg)](https://pkg.go.dev/github.com/sgade/randomorg/v3)

A client for the [Random.org JSON-RPC API](https://api.random.org/json-rpc/4), which
serves true random numbers generated from atmospheric noise. An API key is required and
can be acquired at https://api.random.org/dashboard.

## Install

```sh
go get github.com/sgade/randomorg/v3
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/sgade/randomorg/v3"
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

See the [reference docs](https://pkg.go.dev/github.com/sgade/randomorg/v3) for the full API, or
`example_test.go` for more runnable examples.

## Basic vs. Signed API

This client implements both halves of the Random.org [Core
API](https://api.random.org/json-rpc/4):

- The [Basic API](https://api.random.org/json-rpc/4/basic) (`GenerateIntegers`,
  `GenerateIntegerSequences`, `GenerateDecimalFractions`, `GenerateGaussians`,
  `GenerateStrings`, `GenerateUUIDs`, `GenerateBlobs`) is the simplest way to
  fetch true random values.
- The [Signed API](https://api.random.org/json-rpc/4/signed)
  (`GenerateSignedIntegers` and friends, plus `GetResult` and
  `VerifySignature`) additionally returns a cryptographic signature proving
  the values came from RANDOM.ORG, along with support for tickets
  (`CreateTickets`, `RevealTickets`, `ListTickets`, `GetTicket`) that let a
  third party audit individual values without exposing your API key.

Optional parameters for every method (e.g. `replacement`,
`pregeneratedRandomization`, and — for the Signed API —
`licenseData`/`userData`/`ticketId`) are passed as a trailing, optional
`MethodOptions` struct, so existing calls keep compiling unchanged when
upgrading.
