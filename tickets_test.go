package randomorg_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/sgade/randomorg"
)

// The response bodies below marked "docs example" are taken from
// https://api.random.org/json-rpc/4/signed (the createTickets, revealTickets,
// listTickets and getTicket sections).

func TestCreateTickets(t *testing.T) {
	t.Run("param validation", func(t *testing.T) {
		cases := []struct {
			name string
			n    int
		}{
			{"n too small", 0},
			{"n too large", 51},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				random := newTestRandom(t, failOnRequest(t))
				_, err := random.CreateTickets(context.Background(), tc.n, false)
				if !errors.Is(err, randomorg.ErrParamRange) {
					t.Fatalf("err = %v, want %v", err, randomorg.ErrParamRange)
				}
			})
		}
	})

	// docs example: createTickets Example 1 (showResult: false)
	t.Run("showResult false (docs example)", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": [
					{"ticketId": "ca71d8928623cee5", "creationTime": "2021-03-26 14:43:54Z", "previousTicketId": null, "nextTicketId": null},
					{"ticketId": "8d8b2bded2c3ed28", "creationTime": "2021-03-26 14:43:54Z", "previousTicketId": null, "nextTicketId": null}
				],
				"id": "22746"
			}`), nil
		})

		got, err := random.CreateTickets(context.Background(), 2, false)
		if err != nil {
			t.Fatalf("CreateTickets() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(CreateTickets()) = %d, want 2", len(got))
		}
		if got[0].TicketID != "ca71d8928623cee5" {
			t.Errorf("got[0].TicketID = %q, want ca71d8928623cee5", got[0].TicketID)
		}
		if got[0].PreviousTicketID != nil || got[0].NextTicketID != nil {
			t.Errorf("got[0] chain pointers = %+v, want both nil", got[0])
		}
		if got[0].CreationTime.IsZero() {
			t.Error("got[0].CreationTime is zero")
		}

		if gotParams["showResult"] != false {
			t.Errorf("params[showResult] = %v, want false", gotParams["showResult"])
		}
		if gotParams["n"] != float64(2) {
			t.Errorf("params[n] = %v, want 2", gotParams["n"])
		}
	})

	// docs example: createTickets Example 2 (showResult: true)
	t.Run("showResult true (docs example)", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": [
					{"ticketId": "992104b84ba7aed6", "creationTime": "2021-03-26 14:43:27Z", "previousTicketId": null, "nextTicketId": null},
					{"ticketId": "5f279f7a7aecdcd3", "creationTime": "2021-03-26 14:43:27Z", "previousTicketId": null, "nextTicketId": null}
				],
				"id": "22746"
			}`), nil
		})

		got, err := random.CreateTickets(context.Background(), 2, true)
		if err != nil {
			t.Fatalf("CreateTickets() error = %v", err)
		}
		if len(got) != 2 || got[1].TicketID != "5f279f7a7aecdcd3" {
			t.Fatalf("CreateTickets() = %+v, unexpected", got)
		}
		if gotParams["showResult"] != true {
			t.Errorf("params[showResult] = %v, want true", gotParams["showResult"])
		}
	})
}

func TestRevealTickets(t *testing.T) {
	// docs example: revealTickets Example 1 (first in chain)
	t.Run("first in chain (docs example)", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{"jsonrpc": "2.0", "result": {"ticketCount": 1}, "id": "18873"}`), nil
		})

		got, err := random.RevealTickets(context.Background(), "ca71d8928623cee5")
		if err != nil {
			t.Fatalf("RevealTickets() error = %v", err)
		}
		if got != 1 {
			t.Errorf("RevealTickets() = %d, want 1", got)
		}
		if gotParams["ticketId"] != "ca71d8928623cee5" {
			t.Errorf("params[ticketId] = %v, want ca71d8928623cee5", gotParams["ticketId"])
		}
	})

	// docs example: revealTickets Example 2 (cascades to predecessors)
	t.Run("cascades to predecessors (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{"jsonrpc": "2.0", "result": {"ticketCount": 3}, "id": "5862"}`), nil
		})

		got, err := random.RevealTickets(context.Background(), "2b08a317fa982ec6")
		if err != nil {
			t.Fatalf("RevealTickets() error = %v", err)
		}
		if got != 3 {
			t.Errorf("RevealTickets() = %d, want 3", got)
		}
	})

	// docs example: revealTickets Example 3 (tail ticket, not yet used)
	t.Run("unused tail ticket errors (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"error": {"code": 426, "message": "The ticket you specified has not yet been used", "data": null},
				"id": "20944"
			}`), nil
		})

		_, err := random.RevealTickets(context.Background(), "ea2c0d31720d66e4")
		var apiErr *randomorg.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("errors.As(err, *APIError) = false, want true (err = %v)", err)
		}
		if apiErr.Code != 426 {
			t.Errorf("apiErr.Code = %d, want 426", apiErr.Code)
		}
	})
}

func TestListTickets(t *testing.T) {
	// docs example: listTickets Example 1 (ticketType: singleton)
	t.Run("singleton (docs example)", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": [
					{
						"ticketId": "5f279f7a7aecdcd3",
						"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
						"showResult": true,
						"creationTime": "2021-03-26 14:43:27Z",
						"usedTime": null,
						"serialNumber": null,
						"expirationTime": "2021-04-25 14:43:27Z",
						"previousTicketId": null,
						"nextTicketId": null
					},
					{
						"ticketId": "8d8b2bded2c3ed28",
						"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
						"showResult": false,
						"creationTime": "2021-03-26 14:43:54Z",
						"usedTime": null,
						"serialNumber": null,
						"expirationTime": "2021-04-25 14:43:54Z",
						"previousTicketId": null,
						"nextTicketId": null
					}
				],
				"id": "22746"
			}`), nil
		})

		got, err := random.ListTickets(context.Background(), randomorg.TicketTypeSingleton)
		if err != nil {
			t.Fatalf("ListTickets() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(ListTickets()) = %d, want 2", len(got))
		}
		if !got[0].ShowResult {
			t.Error("got[0].ShowResult = false, want true")
		}
		if got[0].UsedTime != nil {
			t.Errorf("got[0].UsedTime = %v, want nil (unused ticket)", got[0].UsedTime)
		}
		if got[0].SerialNumber != nil {
			t.Errorf("got[0].SerialNumber = %v, want nil (unused ticket)", got[0].SerialNumber)
		}
		if got[0].ExpirationTime == nil || got[0].ExpirationTime.IsZero() {
			t.Error("got[0].ExpirationTime is nil or zero")
		}

		if gotParams["ticketType"] != "singleton" {
			t.Errorf("params[ticketType] = %v, want singleton", gotParams["ticketType"])
		}
	})

	// docs example: listTickets Example 2 (ticketType: head, used tickets)
	t.Run("head, used tickets (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": [
					{
						"ticketId": "992104b84ba7aed6",
						"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
						"showResult": true,
						"creationTime": "2021-03-26 14:43:27Z",
						"usedTime": "2021-03-26 15:19:59Z",
						"serialNumber": 6277,
						"expirationTime": "2021-04-25 14:43:27Z",
						"previousTicketId": null,
						"nextTicketId": "448c8e3467a07577"
					}
				],
				"id": "22746"
			}`), nil
		})

		got, err := random.ListTickets(context.Background(), randomorg.TicketTypeHead)
		if err != nil {
			t.Fatalf("ListTickets() error = %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("len(ListTickets()) = %d, want 1", len(got))
		}
		if got[0].UsedTime == nil || got[0].UsedTime.IsZero() {
			t.Error("got[0].UsedTime is nil or zero")
		}
		if got[0].SerialNumber == nil || *got[0].SerialNumber != 6277 {
			t.Errorf("got[0].SerialNumber = %v, want 6277", got[0].SerialNumber)
		}
		if got[0].NextTicketID == nil || *got[0].NextTicketID != "448c8e3467a07577" {
			t.Errorf("got[0].NextTicketID = %v, want 448c8e3467a07577", got[0].NextTicketID)
		}
	})
}

func TestGetTicket(t *testing.T) {
	// docs example: getTicket Example 1 (used, showResult false)
	t.Run("used, showResult false (docs example)", func(t *testing.T) {
		var gotParams map[string]any
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			gotParams, _ = decodeRequestBody(t, req)["params"].(map[string]any)
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"ticketId": "ca71d8928623cee5",
					"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
					"showResult": false,
					"creationTime": "2021-03-26 14:43:54Z",
					"usedTime": "2021-03-26 15:18:32Z",
					"serialNumber": 6276,
					"expirationTime": "2021-04-25 14:43:54Z",
					"previousTicketId": null,
					"nextTicketId": "d7563dedd09b6b80"
				},
				"id": "22746"
			}`), nil
		})

		got, err := random.GetTicket(context.Background(), "ca71d8928623cee5")
		if err != nil {
			t.Fatalf("GetTicket() error = %v", err)
		}
		if got.SerialNumber == nil || *got.SerialNumber != 6276 {
			t.Errorf("SerialNumber = %v, want 6276", got.SerialNumber)
		}
		if got.Result != nil {
			t.Errorf("Result = %v, want nil (showResult was false)", got.Result)
		}
		if gotParams["ticketId"] != "ca71d8928623cee5" {
			t.Errorf("params[ticketId] = %v, want ca71d8928623cee5", gotParams["ticketId"])
		}
		if _, ok := gotParams["apiKey"]; ok {
			t.Errorf("params[apiKey] present = %v, want absent (getTicket takes no apiKey)", gotParams["apiKey"])
		}
	})

	// docs example: getTicket Example 2 (unused)
	t.Run("unused (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"ticketId": "8d8b2bded2c3ed28",
					"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
					"showResult": false,
					"creationTime": "2021-03-26 14:43:54Z",
					"usedTime": null,
					"serialNumber": null,
					"expirationTime": "2021-04-25 14:43:54Z",
					"previousTicketId": null,
					"nextTicketId": null
				},
				"id": "22746"
			}`), nil
		})

		got, err := random.GetTicket(context.Background(), "8d8b2bded2c3ed28")
		if err != nil {
			t.Fatalf("GetTicket() error = %v", err)
		}
		if got.UsedTime != nil || got.SerialNumber != nil {
			t.Errorf("unused ticket = %+v, want UsedTime and SerialNumber nil", got)
		}
	})

	// docs example: getTicket Example 3 (used, showResult true, full nested result)
	t.Run("used, showResult true (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"ticketId": "992104b84ba7aed6",
					"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
					"showResult": true,
					"creationTime": "2021-03-26 14:43:27Z",
					"usedTime": "2021-03-26 15:19:59Z",
					"serialNumber": 6277,
					"expirationTime": "2021-04-25 14:43:27Z",
					"previousTicketId": null,
					"nextTicketId": "448c8e3467a07577",
					"result": {
						"random": {
							"method": "generateSignedIntegers",
							"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
							"n": 1,
							"min": 0,
							"max": 36,
							"replacement": true,
							"base": 10,
							"pregeneratedRandomization": null,
							"data": [6],
							"license": {"type": "developer", "text": "Random values licensed strictly for development and testing only", "infoUrl": null},
							"licenseData": null,
							"userData": null,
							"ticketData": {"ticketId": "992104b84ba7aed6", "previousTicketId": null, "nextTicketId": "448c8e3467a07577"},
							"completionTime": "2021-03-26 15:19:59Z",
							"serialNumber": 6277
						},
						"signature": "DNELqzKkBC78nAXPk5+TnrolPSY3mzpXYXHdmrOHjyWSDAPE2YICg+02qP5pJR2xjqv+UUl0o52GHRqAB6o75cAa8qd6b6F724M7tAzlZWHKH7Z16/HGDPf82HnMvyd4xA5n0/A4vlvoX9A63hjz30O0qaivqdYEHJOevu9l3e6Q2QVMrMkd3GxCrILOquNZAjrWorMKvHITrJh8zwVxZSDU4mjGX3GEHuFBsImJloQaDrxabgZH5Sc15F6076ULfZ7dzE0W8x3+xm4IeckOo4/Z8jMV0W6AxSmAJEK2dq4xLSyWIVR6wpiPS5v0z9aHkhu6+uXh3UyQVLq3hglCm6Gx4cRTFO+vq1I5xOCXHvQc1RtWYsLbSvLWUCnDQdDwpXIq5kpYhgbbnR1tQmlsmkQOzaHF7IYoSKcg8JGM5y1fDldE+RaUgkQmMEmAMJ9SLs/67W5OW5Gjetqlg4k1rENx7PiZQ91DxJWaIA+G3v3qABDuSNVNSkqLJS6eUvAZu8lLX57FBvwYbbMH41d4fdxnJNk1jkzxeLn3PoUZ6OEnDNCdQv37xaeP/McHwQF9yazNuc/8LyNv6amkzHM0qZooXfbFgV/q3MSDhQ97wEDUiMlEJkIwFw1HFMT8aHA7ChhLxSJtINmWCfjRKZ0p3FRiDTk+uz6doKeXRuckKWo=",
						"cost": 0,
						"bitsUsed": 5,
						"bitsLeft": 249990,
						"requestsLeft": 992,
						"advisoryDelay": 2430
					}
				},
				"id": "22746"
			}`), nil
		})

		got, err := random.GetTicket(context.Background(), "992104b84ba7aed6")
		if err != nil {
			t.Fatalf("GetTicket() error = %v", err)
		}
		if got.Result == nil {
			t.Fatal("Result = nil, want the nested signed result")
		}

		decoded, err := randomorg.DecodeTicketResult[int64](got.Result)
		if err != nil {
			t.Fatalf("DecodeTicketResult() error = %v", err)
		}
		if want := []int64{6}; decoded.Data[0] != want[0] {
			t.Errorf("decoded.Data = %v, want %v", decoded.Data, want)
		}
		if decoded.TicketData == nil || decoded.TicketData.TicketID != "992104b84ba7aed6" {
			t.Errorf("decoded.TicketData = %+v, want TicketID 992104b84ba7aed6", decoded.TicketData)
		}
	})

	// docs example: getTicket Example 4 (unused, showResult true -> result: null)
	t.Run("unused, showResult true (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"result": {
					"ticketId": "5f279f7a7aecdcd3",
					"hashedApiKey": "ncGk4bCmDT7GSc64MzGzNvRUoDT++pTPjntmtuu075JFqKbz/G4nKerq0JQoldvtQxYOCePxMN5gcYZSOC2DTg==",
					"showResult": true,
					"creationTime": "2021-03-26 14:43:27Z",
					"usedTime": null,
					"serialNumber": null,
					"expirationTime": "2021-04-25 14:43:27Z",
					"previousTicketId": null,
					"nextTicketId": null,
					"result": null
				},
				"id": "22746"
			}`), nil
		})

		got, err := random.GetTicket(context.Background(), "5f279f7a7aecdcd3")
		if err != nil {
			t.Fatalf("GetTicket() error = %v", err)
		}
		if got.Result != nil {
			t.Errorf("Result = %v, want nil (literal JSON null should normalize to nil)", got.Result)
		}
	})

	// docs example: getTicket Example 5 (nonexistent)
	t.Run("nonexistent (docs example)", func(t *testing.T) {
		random := newTestRandom(t, func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{
				"jsonrpc": "2.0",
				"error": {"code": 420, "message": "The ticket you specified does not exist", "data": null},
				"id": "13354"
			}`), nil
		})

		_, err := random.GetTicket(context.Background(), "7777777777777777")
		var apiErr *randomorg.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("errors.As(err, *APIError) = false, want true (err = %v)", err)
		}
		if apiErr.Code != 420 {
			t.Errorf("apiErr.Code = %d, want 420", apiErr.Code)
		}
	})
}
