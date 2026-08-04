package randomorg

import (
	"context"
	"encoding/json"
	"time"
)

// Tickets
// see https://api.random.org/json-rpc/4/signed

// Ticket is a single ticket as returned by CreateTickets.
type Ticket struct {
	TicketID         string
	CreationTime     time.Time
	PreviousTicketID *string
	NextTicketID     *string
}

// ticketEnvelope is the wire shape of a single ticket in the createTickets response.
type ticketEnvelope struct {
	TicketID         string  `json:"ticketId"`
	CreationTime     string  `json:"creationTime"`
	PreviousTicketID *string `json:"previousTicketId"`
	NextTicketID     *string `json:"nextTicketId"`
}

func (e ticketEnvelope) toTicket() (Ticket, error) {
	creationTime, err := parseAPITime(e.CreationTime)
	if err != nil {
		return Ticket{}, err
	}

	return Ticket{
		TicketID:         e.TicketID,
		CreationTime:     creationTime,
		PreviousTicketID: e.PreviousTicketID,
		NextTicketID:     e.NextTicketID,
	}, nil
}

type createTicketsParams struct {
	baseParams
	N          int  `json:"n"`
	ShowResult bool `json:"showResult"`
}

// CreateTickets creates n tickets (n in [1, 50]) that can be used, one at a
// time, in place of a TicketID when calling a GenerateSigned* method.
// showResult controls how much detail GetTicket later returns for each
// ticket: if false, only basic ticket information; if true, the full
// signed result produced when the ticket was used (see
// TicketDetails.Result and DecodeTicketResult).
func (r *Random) CreateTickets(ctx context.Context, n int, showResult bool) ([]Ticket, error) {
	if n < 1 || n > 50 {
		return nil, ErrParamRange
	}

	params := createTicketsParams{
		baseParams: baseParams{APIKey: r.apiKey},
		N:          n,
		ShowResult: showResult,
	}

	envelopes, err := invokeRequest[[]ticketEnvelope](ctx, r, "createTickets", params)
	if err != nil {
		return nil, err
	}

	tickets := make([]Ticket, len(envelopes))
	for i, e := range envelopes {
		ticket, err := e.toTicket()
		if err != nil {
			return nil, err
		}
		tickets[i] = ticket
	}

	return tickets, nil
}

type revealTicketsParams struct {
	baseParams
	TicketID string `json:"ticketId"`
}

type revealTicketsResult struct {
	TicketCount int `json:"ticketCount"`
}

// RevealTickets marks ticketID and every predecessor in its chain as
// revealed, meaning subsequent GetTicket calls return their full details
// (as if they had been created with showResult true). It only affects
// tickets that have already been used, and reports how many tickets were
// revealed by the call.
func (r *Random) RevealTickets(ctx context.Context, ticketID string) (int, error) {
	params := revealTicketsParams{
		baseParams: baseParams{APIKey: r.apiKey},
		TicketID:   ticketID,
	}

	result, err := invokeRequest[revealTicketsResult](ctx, r, "revealTickets", params)
	if err != nil {
		return 0, err
	}

	return result.TicketCount, nil
}

// TicketType selects which tickets ListTickets returns.
type TicketType string

const (
	// TicketTypeSingleton selects tickets that are the only ticket in
	// their chain, i.e. that have neither a previous nor a next ticket.
	TicketTypeSingleton TicketType = "singleton"
	// TicketTypeHead selects tickets that are the first in their chain but
	// have a next ticket.
	TicketTypeHead TicketType = "head"
	// TicketTypeTail selects tickets that have a previous ticket but are
	// the last (and always unused) ticket in their chain.
	TicketTypeTail TicketType = "tail"
)

// TicketDetails describes a ticket as returned by ListTickets or GetTicket.
type TicketDetails struct {
	TicketID         string
	HashedAPIKey     string
	ShowResult       bool
	CreationTime     time.Time
	UsedTime         *time.Time
	SerialNumber     *int
	ExpirationTime   *time.Time
	PreviousTicketID *string
	NextTicketID     *string

	// Result is the full signed result produced when the ticket was used.
	// It is only ever populated by GetTicket (never by ListTickets), and
	// only when the ticket was created with showResult true and has
	// already been used; otherwise it is nil. Decode it with
	// DecodeTicketResult.
	Result json.RawMessage
}

// ticketDetailsEnvelope is the wire shape returned by listTickets (as an
// array element) and getTicket (as the whole result).
type ticketDetailsEnvelope struct {
	TicketID         string          `json:"ticketId"`
	HashedAPIKey     string          `json:"hashedApiKey"`
	ShowResult       bool            `json:"showResult"`
	CreationTime     string          `json:"creationTime"`
	UsedTime         *string         `json:"usedTime"`
	SerialNumber     *int            `json:"serialNumber"`
	ExpirationTime   *string         `json:"expirationTime"`
	PreviousTicketID *string         `json:"previousTicketId"`
	NextTicketID     *string         `json:"nextTicketId"`
	Result           json.RawMessage `json:"result"`
}

func (e ticketDetailsEnvelope) toTicketDetails() (TicketDetails, error) {
	creationTime, err := parseAPITime(e.CreationTime)
	if err != nil {
		return TicketDetails{}, err
	}

	details := TicketDetails{
		TicketID:         e.TicketID,
		HashedAPIKey:     e.HashedAPIKey,
		ShowResult:       e.ShowResult,
		CreationTime:     creationTime,
		SerialNumber:     e.SerialNumber,
		PreviousTicketID: e.PreviousTicketID,
		NextTicketID:     e.NextTicketID,
	}

	if e.UsedTime != nil {
		usedTime, err := parseAPITime(*e.UsedTime)
		if err != nil {
			return TicketDetails{}, err
		}
		details.UsedTime = &usedTime
	}
	if e.ExpirationTime != nil {
		expirationTime, err := parseAPITime(*e.ExpirationTime)
		if err != nil {
			return TicketDetails{}, err
		}
		details.ExpirationTime = &expirationTime
	}

	// A ticket created with showResult true but not yet used reports a
	// literal JSON null for "result" (rather than omitting the key).
	// Normalize that to nil, same as an omitted or showResult-false ticket,
	// so callers only need one nil check.
	if len(e.Result) > 0 && string(e.Result) != "null" {
		details.Result = e.Result
	}

	return details, nil
}

type listTicketsParams struct {
	baseParams
	TicketType TicketType `json:"ticketType"`
}

// ListTickets returns up to 4000 tickets of ticketType belonging to the
// client's API key.
func (r *Random) ListTickets(ctx context.Context, ticketType TicketType) ([]TicketDetails, error) {
	params := listTicketsParams{
		baseParams: baseParams{APIKey: r.apiKey},
		TicketType: ticketType,
	}

	envelopes, err := invokeRequest[[]ticketDetailsEnvelope](ctx, r, "listTickets", params)
	if err != nil {
		return nil, err
	}

	tickets := make([]TicketDetails, len(envelopes))
	for i, e := range envelopes {
		details, err := e.toTicketDetails()
		if err != nil {
			return nil, err
		}
		tickets[i] = details
	}

	return tickets, nil
}

type getTicketParams struct {
	TicketID string `json:"ticketId"`
}

// GetTicket returns the details of a single ticket by its ticketID. Unlike
// most methods, it does not take an API key: any ticket ID can be looked up
// by anyone who has it, though the full random-value Result is only ever
// present when the ticket was created with showResult true.
func (r *Random) GetTicket(ctx context.Context, ticketID string) (TicketDetails, error) {
	params := getTicketParams{TicketID: ticketID}

	env, err := invokeRequest[ticketDetailsEnvelope](ctx, r, "getTicket", params)
	if err != nil {
		return TicketDetails{}, err
	}

	return env.toTicketDetails()
}

// DecodeTicketResult decodes TicketDetails.Result, as returned by
// GetTicket, into a typed SignedResult[T]. T must match the data type
// originally generated, as with GetResult.
func DecodeTicketResult[T any](raw json.RawMessage) (SignedResult[T], error) {
	var env signedResultEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return SignedResult[T]{}, err
	}

	return decodeSignedResult[T](env)
}
