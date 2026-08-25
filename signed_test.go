package randomorg_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/sgade/randomorg/v3"
)

// The response bodies below marked "docs example" are taken verbatim (aside
// from re-indentation) from https://api.random.org/json-rpc/4/signed, so
// that decoding is verified against real RANDOM.ORG output, including real
// signatures.

func TestGenerateSignedIntegers(t *testing.T) {
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
				_, err := random.GenerateSignedIntegers(context.Background(), tc.n, tc.min, tc.max)
				if !errors.Is(err, randomorg.ErrParamRange) {
					t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
				}
			})
		}
	})

	// docs example: generateSignedIntegers Example 1 (dice roll)
	const diceExampleBody = `{
		"jsonrpc": "2.0",
		"result": {
			"random": {
				"method": "generateSignedIntegers",
				"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
				"n": 3,
				"min": 1,
				"max": 6,
				"replacement": true,
				"base": 10,
				"pregeneratedRandomization": null,
				"data": [1, 3, 1],
				"license": {
					"type": "developer",
					"text": "Random values licensed strictly for development and testing only",
					"infoUrl": null
				},
				"licenseData": null,
				"userData": null,
				"ticketData": null,
				"completionTime": "2021-03-15 13:51:32Z",
				"serialNumber": 6116
			},
			"signature": "hprai35Zc95uAM47oVpqUTEiVla/GvF+u/8GjZCvcGKRG86fQrnVvuzn1HN5VrJoU13SDE96DmggtTYECzkk9bzfVnhHg47/Zn+7w27GedseB2F4QxNtf7aycvcdBHnSg08IaVo+ohPiqlZcxpx5TVUfmLb6LfYRPirQUHMv5vpT7ba/hDSb7bQ6wGpiV1By48nDC5p/ncZEvfAHQcrNxtrtCbwQoI9BMBxRXqV5DaG6YYPxTpQeg9dWJMhZJuBNWIf4hsCKoOGkyBI/uHPaGgTy5jmSk4cFutK3jQP+9vWkDwYQ9sgok0U9Dgp5jG2zC6JOwaEgosagY7B29r1s6aXxcZCXFtX9yBdAh6Of7Z1PeLeva14lQWdZmqYSYvD56HlYWQfeb0lY2Lgf7Yvr9W/lxUxSg9OUvXi+urR0sprXpGwOcml5dSVRXyG6oyDphwXsvJ8h9ofiCP5rkyxHNphR6s1LF5NQ91OCBDllXiwXAKvJBcBxftFVAJRqpRALuLQB2xTXlrld/XBEBc93Pve3e+B0DancFa1XHgBFLlRSmF+MpSY+8qIT2U4hHSGO38ISSX2RdHYR+talXoQ8Vj6fiibzZCUNMbXp4HcYRjmWUVCii0otGYC/fSg25ZmnpG/SMJXfDbVpzx8sC49qYpaN9GRG5QC5pHfA69nJVqo=",
			"cost": 0,
			"bitsUsed": 8,
			"bitsLeft": 249992,
			"requestsLeft": 999,
			"advisoryDelay": 2310
		},
		"id": "6995"
	}`

	t.Run("generates values (docs example)", func(t *testing.T) {
		var gotReq map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotReq = decodeRequestBody(t, req)
			return jsonResponse(http.StatusOK, diceExampleBody), nil
		})

		got, err := random.GenerateSignedIntegers(context.Background(), 3, 1, 6, randomorg.GenerateSignedIntegersOptions{
			Replacement: randomorg.Bool(true),
		})
		if err != nil {
			t.Fatalf("GenerateSignedIntegers() error = %v", err)
		}

		if want := []int64{1, 3, 1}; !slices.Equal(got.Data, want) {
			t.Errorf("Data = %v, want %v", got.Data, want)
		}
		if got.SerialNumber != 6116 {
			t.Errorf("SerialNumber = %d, want 6116", got.SerialNumber)
		}
		if got.Signature == "" {
			t.Error("Signature is empty")
		}
		if got.BitsUsed != 8 || got.BitsLeft != 249992 || got.RequestsLeft != 999 || got.AdvisoryDelay != 2310 {
			t.Errorf("usage fields = %+v, unexpected", got)
		}
		if got.License.Type != "developer" {
			t.Errorf("License.Type = %q, want developer", got.License.Type)
		}
		wantTime, _ := time.Parse(time.RFC3339, "2021-03-15T13:51:32Z")
		if !got.CompletionTime.Equal(wantTime) {
			t.Errorf("CompletionTime = %v, want %v", got.CompletionTime, wantTime)
		}

		// Random must carry the exact bytes of the "random" object, since
		// it's what VerifySignature (or independent verification) needs.
		var raw map[string]any
		if err := json.Unmarshal(got.Random, &raw); err != nil {
			t.Fatalf("Random did not contain valid JSON: %v", err)
		}
		if raw["serialNumber"].(float64) != 6116 {
			t.Errorf("Random[serialNumber] = %v, want 6116", raw["serialNumber"])
		}

		if gotReq["method"] != "generateSignedIntegers" {
			t.Errorf("request method = %v, want generateSignedIntegers", gotReq["method"])
		}
	})

	t.Run("with base", func(t *testing.T) {
		t.Run("rejects base 10", func(t *testing.T) {
			random := newTestRandom(t, failOnRequest(t))
			_, err := random.GenerateSignedIntegersWithBase(context.Background(), 1, 0, 10, 10)
			if !errors.Is(err, randomorg.ErrParamRange) {
				t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
			}
		})

		t.Run("rejects invalid base", func(t *testing.T) {
			random := newTestRandom(t, failOnRequest(t))
			_, err := random.GenerateSignedIntegersWithBase(context.Background(), 1, 0, 10, 3)
			if !errors.Is(err, randomorg.ErrParamRange) {
				t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
			}
		})

		t.Run("decodes string data", func(t *testing.T) {
			var gotParams map[string]any
			random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
				gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
				return jsonResponse(http.StatusOK, `{
					"jsonrpc": "2.0",
					"result": {
						"random": {
							"method": "generateSignedIntegers",
							"hashedApiKey": "abc==",
							"n": 2,
							"min": 0,
							"max": 255,
							"replacement": true,
							"base": 16,
							"pregeneratedRandomization": null,
							"data": ["0a", "ff"],
							"license": {"type": "developer", "text": "dev only", "infoUrl": null},
							"licenseData": null,
							"userData": null,
							"ticketData": null,
							"completionTime": "2021-03-15 13:51:32Z",
							"serialNumber": 1
						},
						"signature": "sig==",
						"cost": 0,
						"bitsUsed": 16,
						"bitsLeft": 1,
						"requestsLeft": 1,
						"advisoryDelay": 0
					},
					"id": "1"
				}`), nil
			})

			got, err := random.GenerateSignedIntegersWithBase(context.Background(), 2, 0, 255, 16)
			if err != nil {
				t.Fatalf("GenerateSignedIntegersWithBase() error = %v", err)
			}
			if want := []string{"0a", "ff"}; !slices.Equal(got.Data, want) {
				t.Errorf("Data = %v, want %v", got.Data, want)
			}
			if gotParams["base"] != float64(16) {
				t.Errorf("params[base] = %v, want 16", gotParams["base"])
			}
		})
	})
}

func TestGenerateSignedIntegerSequences(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		random := newTestRandom(t, failOnRequest(t))
		_, err := random.GenerateSignedIntegerSequences(context.Background(), 2, []int{1}, []int64{0, 0}, []int64{10, 10})
		if !errors.Is(err, randomorg.ErrParamRange) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
		}
	})

	// docs example: generateSignedIntegerSequences Example 1
	t.Run("generates values (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"random": {
						"method": "generateSignedIntegerSequences",
						"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
						"n": 2,
						"length": [5, 1],
						"min": 1,
						"max": [69, 26],
						"replacement": false,
						"base": 10,
						"pregeneratedRandomization": null,
						"data": [[54, 3, 0, 26, 36], [19]],
						"license": {
							"type": "developer",
							"text": "Random values licensed strictly for development and testing only",
							"infoUrl": null
						},
						"licenseData": null,
						"userData": null,
						"ticketData": null,
						"completionTime": "2021-03-16 10:13:58Z",
						"serialNumber": 6139
					},
					"signature": "ggeQrrjX9M1FFT2Uv4xlz4AIpjMMdPvJfkE0RUIOj6oBsfTLpit+tz9XsNKRgqyoGUnygXW7EWEkzFESXk5QeizLZrEkmEylzC4QOv2Cu5xE7xY+S7jv+BHK/Db5FnBRPOiPiY7KpxSyLBlOZ4PeCshNacsXlFK6nV9SF+CvECMUchA7q8VOr2PYsFVRTVg7vVhRxZD1Qy9ba9TGC+F+TkbNkFJrTGHsqA3KUXeDVDEeueQxDyPsM6Z2gqAt7ciFCRcHcIp52Ik20eLlXiV6PgVppeRk9AHl14cfujB5aUtDsqldGWgARqmxWah4R9RhLzuill3PolB0HTe+VfQr9BcuiHgWazKbDsibhGWCLP3tLKdBpe+ow1xy0fVK6OWsMMzpznahlehl33NmHQI+t6e6uY0yVVlIB0wcTNzTRdrWqJoHD4c36mMgqweZGYwfzpsEnm3SWTyQpCLxxfEqdkuGf9mjLxIN4bvgHtDG9X66sle7jO21Ssvq6F4Hyo6g2yKuJFORSYYu6ukgU3u/MhU08iRBMbT9dvnN30cH1ay/HOULEXu5WHdWYiUMPKWpFqZPiV9yn4rSlGXIqRs6zYCJxz41zL0JTAZMd9eTbx3cCcuK5ws0f8tNfR7Bi1xEOr1F/5U89JXu4wHutqomy4AL7bXd0aNxnQ8taMwH0xk=",
					"cost": 0,
					"bitsUsed": 36,
					"bitsLeft": 249924,
					"requestsLeft": 998,
					"advisoryDelay": 1560
				},
				"id": "6995"
			}`), nil
		})

		got, err := random.GenerateSignedIntegerSequences(context.Background(), 2, []int{5, 1}, []int64{1, 1}, []int64{69, 26}, randomorg.GenerateSignedIntegerSequencesOptions{
			Replacement: []bool{false, false},
		})
		if err != nil {
			t.Fatalf("GenerateSignedIntegerSequences() error = %v", err)
		}
		want := [][]int64{{54, 3, 0, 26, 36}, {19}}
		if len(got.Data) != len(want) {
			t.Fatalf("Data = %v, want %v", got.Data, want)
		}
		for i := range want {
			if !slices.Equal(got.Data[i], want[i]) {
				t.Fatalf("Data[%d] = %v, want %v", i, got.Data[i], want[i])
			}
		}
		if got.SerialNumber != 6139 {
			t.Errorf("SerialNumber = %d, want 6139", got.SerialNumber)
		}
	})

	t.Run("with base", func(t *testing.T) {
		t.Run("base slice wrong size is rejected", func(t *testing.T) {
			random := newTestRandom(t, failOnRequest(t))
			_, err := random.GenerateSignedIntegerSequencesWithBase(context.Background(), 2, []int{1, 1}, []int64{0, 0}, []int64{10, 10}, []int{16})
			if !errors.Is(err, randomorg.ErrParamRange) {
				t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
			}
		})

		t.Run("base 10 is rejected", func(t *testing.T) {
			random := newTestRandom(t, failOnRequest(t))
			_, err := random.GenerateSignedIntegerSequencesWithBase(context.Background(), 1, []int{1}, []int64{0}, []int64{10}, []int{10})
			if !errors.Is(err, randomorg.ErrParamRange) {
				t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
			}
		})

		t.Run("decodes string data", func(t *testing.T) {
			random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
				return jsonResponse(http.StatusOK, `{
					"jsonrpc": "2.0",
					"result": {
						"random": {
							"method": "generateSignedIntegerSequences",
							"hashedApiKey": "abc==",
							"n": 1,
							"length": [2],
							"min": [0],
							"max": [255],
							"replacement": [true],
							"base": [16],
							"pregeneratedRandomization": null,
							"data": [["0a", "ff"]],
							"license": {"type": "developer", "text": "dev only", "infoUrl": null},
							"licenseData": null,
							"userData": null,
							"ticketData": null,
							"completionTime": "2021-03-15 13:51:32Z",
							"serialNumber": 1
						},
						"signature": "sig==",
						"cost": 0,
						"bitsUsed": 16,
						"bitsLeft": 1,
						"requestsLeft": 1,
						"advisoryDelay": 0
					},
					"id": "1"
				}`), nil
			})

			got, err := random.GenerateSignedIntegerSequencesWithBase(context.Background(), 1, []int{2}, []int64{0}, []int64{255}, []int{16})
			if err != nil {
				t.Fatalf("GenerateSignedIntegerSequencesWithBase() error = %v", err)
			}
			want := [][]string{{"0a", "ff"}}
			if len(got.Data) != 1 || !slices.Equal(got.Data[0], want[0]) {
				t.Fatalf("Data = %v, want %v", got.Data, want)
			}
		})
	})
}

func TestGenerateSignedDecimalFractions(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		random := newTestRandom(t, failOnRequest(t))
		_, err := random.GenerateSignedDecimalFractions(context.Background(), 0, 2)
		if !errors.Is(err, randomorg.ErrParamRange) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
		}
	})

	// docs example: generateSignedDecimalFractions Example 1
	t.Run("generates values (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"random": {
						"method": "generateSignedDecimalFractions",
						"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
						"n": 10,
						"decimalPlaces": 8,
						"replacement": true,
						"pregeneratedRandomization": null,
						"data": [0.23213486, 0.97593769, 0.88911014, 0.63882311, 0.90541324, 0.81344571, 0.69891248, 0.6300596, 0.8323724, 0.17882089],
						"license": {
							"type": "developer",
							"text": "Random values licensed strictly for development and testing only",
							"infoUrl": null
						},
						"licenseData": null,
						"userData": null,
						"ticketData": null,
						"completionTime": "2021-03-16 11:30:16Z",
						"serialNumber": 6156
					},
					"signature": "0arTfIyzLtMA1+Nzo++qeKtSEC2pxCuc4Bb5EsoA3FSTfmEctBQJeuWa4Tl4h/DEPvoiNQ4c/awzjFMOlcN8SsyJnRNSDUpOp8L48gFAvLAyDVflTwe3kJT6N9AAeqIt57m/R+Vzl5RAvS0LMH0tMg1L8mB43aRyZP1rF/YrPJ4dMOugOD+G7Dgqi1U3q+SxcpXoQpbsm1DptpIBMXyBQmxSC6kwkzbPXq4dDY8hcbNG+rcUaWDjbS1ptkFIuDlNhgvY6quC3AiuDt6wYgeLFzGl6WYQ/pGd9B0lklyn3Op2WfgPGqwyhe3FpZ0iWULj8WvdN/ZNkJ67Okz1fRIv7Q/zYIq9btL06Sw+IL2UDSin4Jr/gbjoNpBnfVgvbrQJkieg/aFZIB/KTmTOGfYfsRFsLXSebJALj+F4TWSiEuX7B5qsZq+gNN/5B4GNFoA6khUaJQ9yMaiyi/s7VTI4Au5pvu+lJ+I+tucVJf4uuKKwEuzlpUSxmZZ2HZvYoG2mi8rZo8hTZmGEBjQ4wSZeyp81Y1umZ321qwC42d3sgTS/sLpQaAMTs752zxwJSmGwEkS+NzswGurghtsirNhu0ZhN7yNe3iWX5hqtkncJiEYvH/vXqT90QyUQOvDz3xjOpYYhh/muYTL9kJbypSsKmO8vdiapmGfrVkSEcJFS0GU=",
					"cost": 0,
					"bitsUsed": 266,
					"bitsLeft": 246919,
					"requestsLeft": 981,
					"advisoryDelay": 2070
				},
				"id": "6995"
			}`), nil
		})

		got, err := random.GenerateSignedDecimalFractions(context.Background(), 10, 8)
		if err != nil {
			t.Fatalf("GenerateSignedDecimalFractions() error = %v", err)
		}
		if len(got.Data) != 10 {
			t.Fatalf("len(Data) = %d, want 10", len(got.Data))
		}
		if got.Data[0] != 0.23213486 {
			t.Errorf("Data[0] = %v, want 0.23213486", got.Data[0])
		}
		if got.SerialNumber != 6156 {
			t.Errorf("SerialNumber = %d, want 6156", got.SerialNumber)
		}
	})
}

func TestGenerateSignedGaussians(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		random := newTestRandom(t, failOnRequest(t))
		_, err := random.GenerateSignedGaussians(context.Background(), 0, 0, 1, 4)
		if !errors.Is(err, randomorg.ErrParamRange) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
		}
	})

	t.Run("generates values", func(t *testing.T) {
		var gotMethod string
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotMethod, _ = decodeRequestBody(t, req)["method"].(string)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"random": {
						"method": "generateSignedGaussians",
						"hashedApiKey": "abc==",
						"n": 4,
						"mean": 0,
						"standardDeviation": 1,
						"significantDigits": 8,
						"pregeneratedRandomization": null,
						"data": [0.4025041, -1.4918831, 0.64733849, 0.5222242],
						"license": {"type": "developer", "text": "dev only", "infoUrl": null},
						"licenseData": null,
						"userData": null,
						"ticketData": null,
						"completionTime": "2013-01-25 19:16:42Z",
						"serialNumber": 42
					},
					"signature": "sig==",
					"cost": 0,
					"bitsUsed": 106,
					"bitsLeft": 199894,
					"requestsLeft": 5442,
					"advisoryDelay": 0
				},
				"id": "1"
			}`), nil
		})

		got, err := random.GenerateSignedGaussians(context.Background(), 4, 0, 1, 8)
		if err != nil {
			t.Fatalf("GenerateSignedGaussians() error = %v", err)
		}
		if want := []float64{0.4025041, -1.4918831, 0.64733849, 0.5222242}; !slices.Equal(got.Data, want) {
			t.Errorf("Data = %v, want %v", got.Data, want)
		}
		if gotMethod != "generateSignedGaussians" {
			t.Errorf("method = %q, want generateSignedGaussians", gotMethod)
		}
	})
}

func TestGenerateSignedStrings(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		random := newTestRandom(t, failOnRequest(t))
		_, err := random.GenerateSignedStrings(context.Background(), 0, 5, "abc")
		if !errors.Is(err, randomorg.ErrParamRange) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
		}
	})

	t.Run("generates values", func(t *testing.T) {
		var gotMethod string
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotMethod, _ = decodeRequestBody(t, req)["method"].(string)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"random": {
						"method": "generateSignedStrings",
						"hashedApiKey": "abc==",
						"n": 2,
						"length": 3,
						"characters": "abc",
						"replacement": true,
						"pregeneratedRandomization": null,
						"data": ["abc", "cab"],
						"license": {"type": "developer", "text": "dev only", "infoUrl": null},
						"licenseData": null,
						"userData": null,
						"ticketData": null,
						"completionTime": "2011-10-10 13:19:12Z",
						"serialNumber": 42
					},
					"signature": "sig==",
					"cost": 0,
					"bitsUsed": 10,
					"bitsLeft": 1,
					"requestsLeft": 1,
					"advisoryDelay": 0
				},
				"id": "1"
			}`), nil
		})

		got, err := random.GenerateSignedStrings(context.Background(), 2, 3, "abc")
		if err != nil {
			t.Fatalf("GenerateSignedStrings() error = %v", err)
		}
		if want := []string{"abc", "cab"}; !slices.Equal(got.Data, want) {
			t.Errorf("Data = %v, want %v", got.Data, want)
		}
		if gotMethod != "generateSignedStrings" {
			t.Errorf("method = %q, want generateSignedStrings", gotMethod)
		}
	})
}

func TestGenerateSignedUUIDs(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		random := newTestRandom(t, failOnRequest(t))
		_, err := random.GenerateSignedUUIDs(context.Background(), 0)
		if !errors.Is(err, randomorg.ErrParamRange) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
		}
	})

	t.Run("generates values", func(t *testing.T) {
		var gotMethod string
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotMethod, _ = decodeRequestBody(t, req)["method"].(string)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"random": {
						"method": "generateSignedUUIDs",
						"hashedApiKey": "abc==",
						"n": 1,
						"pregeneratedRandomization": null,
						"data": ["47849fd4-b790-492e-8b93-c601a91b662d"],
						"license": {"type": "developer", "text": "dev only", "infoUrl": null},
						"licenseData": null,
						"userData": null,
						"ticketData": null,
						"completionTime": "2013-02-11 16:42:07Z",
						"serialNumber": 42
					},
					"signature": "sig==",
					"cost": 0,
					"bitsUsed": 122,
					"bitsLeft": 998532,
					"requestsLeft": 199996,
					"advisoryDelay": 1000
				},
				"id": "1"
			}`), nil
		})

		got, err := random.GenerateSignedUUIDs(context.Background(), 1)
		if err != nil {
			t.Fatalf("GenerateSignedUUIDs() error = %v", err)
		}
		if want := []string{"47849fd4-b790-492e-8b93-c601a91b662d"}; !slices.Equal(got.Data, want) {
			t.Errorf("Data = %v, want %v", got.Data, want)
		}
		if gotMethod != "generateSignedUUIDs" {
			t.Errorf("method = %q, want generateSignedUUIDs", gotMethod)
		}
	})
}

func TestGenerateSignedBlobs(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		cases := []struct {
			name    string
			n, size int
		}{
			{"n too small", 0, 8},
			{"size not multiple of 8", 1, 7},
			{"aggregate size too large", 2, 1_048_576},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				random := newTestRandom(t, failOnRequest(t))
				_, err := random.GenerateSignedBlobs(context.Background(), tc.n, tc.size)
				if !errors.Is(err, randomorg.ErrParamRange) {
					t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
				}
			})
		}
	})

	t.Run("invalid format is rejected", func(t *testing.T) {
		random := newTestRandom(t, failOnRequest(t))
		_, err := random.GenerateSignedBlobs(context.Background(), 1, 8, randomorg.GenerateSignedBlobsOptions{Format: "bogus"})
		if !errors.Is(err, randomorg.ErrParamRange) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
		}
	})

	t.Run("generates values", func(t *testing.T) {
		var gotMethod string
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotMethod, _ = decodeRequestBody(t, req)["method"].(string)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"random": {
						"method": "generateSignedBlobs",
						"hashedApiKey": "abc==",
						"n": 1,
						"size": 8,
						"format": "base64",
						"pregeneratedRandomization": null,
						"data": ["ZGVhZGJlZWY="],
						"license": {"type": "developer", "text": "dev only", "infoUrl": null},
						"licenseData": null,
						"userData": null,
						"ticketData": null,
						"completionTime": "2011-10-10 13:19:12Z",
						"serialNumber": 42
					},
					"signature": "sig==",
					"cost": 0,
					"bitsUsed": 8,
					"bitsLeft": 1,
					"requestsLeft": 1,
					"advisoryDelay": 0
				},
				"id": "1"
			}`), nil
		})

		got, err := random.GenerateSignedBlobs(context.Background(), 1, 8)
		if err != nil {
			t.Fatalf("GenerateSignedBlobs() error = %v", err)
		}
		if want := []string{"ZGVhZGJlZWY="}; !slices.Equal(got.Data, want) {
			t.Errorf("Data = %v, want %v", got.Data, want)
		}
		if gotMethod != "generateSignedBlobs" {
			t.Errorf("method = %q, want generateSignedBlobs", gotMethod)
		}
	})
}

func TestSignedCommonOptions_Validation(t *testing.T) {
	t.Run("userData too large", func(t *testing.T) {
		random := newTestRandom(t, failOnRequest(t))
		_, err := random.GenerateSignedIntegers(context.Background(), 1, 0, 10, randomorg.GenerateSignedIntegersOptions{
			SignedCommonOptions: randomorg.SignedCommonOptions{
				UserData: strings.Repeat("a", 1_001),
			},
		})
		if !errors.Is(err, randomorg.ErrParamRange) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
		}
	})

	t.Run("pregeneratedRandomization id too long", func(t *testing.T) {
		random := newTestRandom(t, failOnRequest(t))
		_, err := random.GenerateSignedIntegers(context.Background(), 1, 0, 10, randomorg.GenerateSignedIntegersOptions{
			SignedCommonOptions: randomorg.SignedCommonOptions{
				PregeneratedRandomization: randomorg.PregeneratedRandomizationByID(strings.Repeat("a", 65)),
			},
		})
		if !errors.Is(err, randomorg.ErrParamRange) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
		}
	})

	t.Run("licenseData and ticketId are sent", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"random": {
						"method": "generateSignedIntegers",
						"hashedApiKey": "abc==",
						"n": 1,
						"min": 0,
						"max": 10,
						"replacement": true,
						"base": 10,
						"pregeneratedRandomization": null,
						"data": [5],
						"license": {"type": "flexibleGambling", "text": "licensed", "infoUrl": null},
						"licenseData": {"maxPayoutValue": {"currency": "USD", "amount": 100}},
						"userData": null,
						"ticketData": {"ticketId": "abc123", "previousTicketId": null, "nextTicketId": null},
						"completionTime": "2021-03-15 13:51:32Z",
						"serialNumber": 1
					},
					"signature": "sig==",
					"cost": 0.01,
					"bitsUsed": 4,
					"bitsLeft": 1,
					"requestsLeft": 1,
					"advisoryDelay": 0
				},
				"id": "1"
			}`), nil
		})

		got, err := random.GenerateSignedIntegers(context.Background(), 1, 0, 10, randomorg.GenerateSignedIntegersOptions{
			SignedCommonOptions: randomorg.SignedCommonOptions{
				LicenseData: &randomorg.LicenseData{
					MaxPayoutValue: randomorg.MaxPayoutValue{Currency: "USD", Amount: 100},
				},
				TicketID: "abc123",
			},
		})
		if err != nil {
			t.Fatalf("GenerateSignedIntegers() error = %v", err)
		}

		if got.TicketData == nil || got.TicketData.TicketID != "abc123" {
			t.Errorf("TicketData = %+v, want TicketID abc123", got.TicketData)
		}
		if got.Cost != 0.01 {
			t.Errorf("Cost = %v, want 0.01", got.Cost)
		}

		params, ok := gotParams["licenseData"].(map[string]any)
		if !ok {
			t.Fatalf("params[licenseData] missing or wrong type: %v", gotParams["licenseData"])
		}
		mpv, ok := params["maxPayoutValue"].(map[string]any)
		if !ok || mpv["currency"] != "USD" {
			t.Errorf("params[licenseData][maxPayoutValue] = %v, want currency USD", params["maxPayoutValue"])
		}
		if gotParams["ticketId"] != "abc123" {
			t.Errorf("params[ticketId] = %v, want abc123", gotParams["ticketId"])
		}
	})
}

func TestGetResult(t *testing.T) {
	// docs example: getResult Example 1
	t.Run("success (docs example)", func(t *testing.T) {
		var gotReq map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotReq = decodeRequestBody(t, req)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"random": {
						"method": "generateSignedIntegers",
						"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
						"n": 3,
						"min": 1,
						"max": 6,
						"replacement": true,
						"base": 10,
						"pregeneratedRandomization": null,
						"data": [1, 3, 1],
						"license": {"type": "developer", "text": "Random values licensed strictly for development and testing only", "infoUrl": null},
						"licenseData": null,
						"userData": null,
						"ticketData": null,
						"completionTime": "2021-03-15 13:51:32Z",
						"serialNumber": 6116
					},
					"signature": "hprai35Zc95uAM47oVpqUTEiVla/GvF+u/8GjZCvcGKRG86fQrnVvuzn1HN5VrJoU13SDE96DmggtTYECzkk9bzfVnhHg47/Zn+7w27GedseB2F4QxNtf7aycvcdBHnSg08IaVo+ohPiqlZcxpx5TVUfmLb6LfYRPirQUHMv5vpT7ba/hDSb7bQ6wGpiV1By48nDC5p/ncZEvfAHQcrNxtrtCbwQoI9BMBxRXqV5DaG6YYPxTpQeg9dWJMhZJuBNWIf4hsCKoOGkyBI/uHPaGgTy5jmSk4cFutK3jQP+9vWkDwYQ9sgok0U9Dgp5jG2zC6JOwaEgosagY7B29r1s6aXxcZCXFtX9yBdAh6Of7Z1PeLeva14lQWdZmqYSYvD56HlYWQfeb0lY2Lgf7Yvr9W/lxUxSg9OUvXi+urR0sprXpGwOcml5dSVRXyG6oyDphwXsvJ8h9ofiCP5rkyxHNphR6s1LF5NQ91OCBDllXiwXAKvJBcBxftFVAJRqpRALuLQB2xTXlrld/XBEBc93Pve3e+B0DancFa1XHgBFLlRSmF+MpSY+8qIT2U4hHSGO38ISSX2RdHYR+talXoQ8Vj6fiibzZCUNMbXp4HcYRjmWUVCii0otGYC/fSg25ZmnpG/SMJXfDbVpzx8sC49qYpaN9GRG5QC5pHfA69nJVqo=",
					"cost": 0,
					"bitsUsed": 8,
					"bitsLeft": 249992,
					"requestsLeft": 999,
					"advisoryDelay": 2310
				},
				"id": "8337"
			}`), nil
		})

		got, err := randomorg.GetResult[int64](context.Background(), random, 6116)
		if err != nil {
			t.Fatalf("GetResult() error = %v", err)
		}
		if want := []int64{1, 3, 1}; !slices.Equal(got.Data, want) {
			t.Errorf("Data = %v, want %v", got.Data, want)
		}
		if got.SerialNumber != 6116 {
			t.Errorf("SerialNumber = %d, want 6116", got.SerialNumber)
		}

		if gotReq["method"] != "getResult" {
			t.Errorf("request method = %v, want getResult", gotReq["method"])
		}
		params, _ := gotReq["params"].(map[string]any)
		if params["serialNumber"] != float64(6116) {
			t.Errorf("request params[serialNumber] = %v, want 6116", params["serialNumber"])
		}
	})

	// docs example: getResult Example 2 (unknown apiKey)
	t.Run("unknown api key (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"error": {"code": 303, "message": "The resource identified by 'apiKey' was not found", "data": ["apiKey"]},
				"id": "13609"
			}`), nil
		})

		_, err := randomorg.GetResult[int64](context.Background(), random, 2647656)
		var apiErr *randomorg.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("errors.As(err, *APIError) = false, want true (err = %v)", err)
		}
		if apiErr.Code != 303 {
			t.Errorf("apiErr.Code = %d, want 303", apiErr.Code)
		}
	})

	// docs example: getResult Example 3 (unknown serialNumber)
	t.Run("unknown serial number (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"error": {"code": 303, "message": "The resource identified by 'serialNumber' was not found", "data": ["serialNumber"]},
				"id": "28447"
			}`), nil
		})

		_, err := randomorg.GetResult[int64](context.Background(), random, 1)
		var apiErr *randomorg.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("errors.As(err, *APIError) = false, want true (err = %v)", err)
		}
		if apiErr.Code != 303 {
			t.Errorf("apiErr.Code = %d, want 303", apiErr.Code)
		}
	})
}

func TestVerifySignature(t *testing.T) {
	const randomObject = `{
		"method": "generateSignedIntegers",
		"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
		"n": 3,
		"min": 1,
		"max": 6,
		"replacement": true,
		"base": 10,
		"pregeneratedRandomization": null,
		"data": [1, 3, 1],
		"license": {"type": "developer", "text": "Random values licensed strictly for development and testing only", "infoUrl": null},
		"licenseData": null,
		"userData": null,
		"ticketData": null,
		"completionTime": "2021-03-15 13:51:32Z",
		"serialNumber": 6116
	}`
	const signature = "hprai35Zc95uAM47oVpqUTEiVla/GvF+u/8GjZCvcGKRG86fQrnVvuzn1HN5VrJoU13SDE96DmggtTYECzkk9bzfVnhHg47/Zn+7w27GedseB2F4QxNtf7aycvcdBHnSg08IaVo+ohPiqlZcxpx5TVUfmLb6LfYRPirQUHMv5vpT7ba/hDSb7bQ6wGpiV1By48nDC5p/ncZEvfAHQcrNxtrtCbwQoI9BMBxRXqV5DaG6YYPxTpQeg9dWJMhZJuBNWIf4hsCKoOGkyBI/uHPaGgTy5jmSk4cFutK3jQP+9vWkDwYQ9sgok0U9Dgp5jG2zC6JOwaEgosagY7B29r1s6aXxcZCXFtX9yBdAh6Of7Z1PeLeva14lQWdZmqYSYvD56HlYWQfeb0lY2Lgf7Yvr9W/lxUxSg9OUvXi+urR0sprXpGwOcml5dSVRXyG6oyDphwXsvJ8h9ofiCP5rkyxHNphR6s1LF5NQ91OCBDllXiwXAKvJBcBxftFVAJRqpRALuLQB2xTXlrld/XBEBc93Pve3e+B0DancFa1XHgBFLlRSmF+MpSY+8qIT2U4hHSGO38ISSX2RdHYR+talXoQ8Vj6fiibzZCUNMbXp4HcYRjmWUVCii0otGYC/fSg25ZmnpG/SMJXfDbVpzx8sC49qYpaN9GRG5QC5pHfA69nJVqo="

	// docs example: verifySignature Example 1 (authentic)
	t.Run("authentic (docs example)", func(t *testing.T) {
		var gotReq map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotReq = decodeRequestBody(t, req)
			return jsonResponse(http.StatusOK, `{"jsonrpc": "2.0", "result": {"authenticity": true}, "id": "8337"}`), nil
		})

		ok, err := random.VerifySignature(context.Background(), json.RawMessage(randomObject), signature)
		if err != nil {
			t.Fatalf("VerifySignature() error = %v", err)
		}
		if !ok {
			t.Error("VerifySignature() = false, want true")
		}
		if gotReq["method"] != "verifySignature" {
			t.Errorf("request method = %v, want verifySignature", gotReq["method"])
		}
	})

	// docs example: verifySignature Example 2 (tampered data)
	t.Run("tampered data (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{"jsonrpc": "2.0", "result": {"authenticity": false}, "id": "8337"}`), nil
		})

		tampered := strings.Replace(randomObject, `"data": [1, 3, 1]`, `"data": [6, 3, 1]`, 1)
		ok, err := random.VerifySignature(context.Background(), json.RawMessage(tampered), signature)
		if err != nil {
			t.Fatalf("VerifySignature() error = %v", err)
		}
		if ok {
			t.Error("VerifySignature() = true, want false")
		}
	})

	t.Run("http error propagates", func(t *testing.T) {
		random := newTestRandom(t, func(*http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusInternalServerError, ""), nil
		})

		_, err := random.VerifySignature(context.Background(), json.RawMessage(randomObject), signature)
		if !errors.Is(err, randomorg.ErrHTTPStatus) {
			t.Fatalf("err = %v, want %v", err, randomorg.ErrHTTPStatus)
		}
	})
}
