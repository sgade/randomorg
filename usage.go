package randomorg

import (
	"context"
	"strings"
	"time"
)

// Usage holds information related to the the usage of a given API key.
type Usage struct {
	// A string indicating the API key's current status, which may be stopped, paused or running.
	// An API key must be running for it to be able to serve requests.
	Status string
	// A timestamp at which the API key was created.
	CreationTime time.Time
	// An integer containing the (estimated) number of remaining true random bits available to the client.
	BitsLeft int
	// An integer containing the (estimated) number of remaining API requests available to the client.
	RequestsLeft int
	// An integer containing the number of bits used by this API key since it was created.
	TotalBits int
	// An integer containing the number of requests used by this API key since it was created.
	TotalRequests int
	// Defines if this instance contains all information.
	isComplete bool
}

// usageFields are the usage-related fields present, wholly or partially, in
// nearly every Random.org API response. Pointer fields distinguish "absent"
// from "present with the zero value".
type usageFields struct {
	Status        *string `json:"status"`
	CreationTime  *string `json:"creationTime"`
	BitsLeft      *int    `json:"bitsLeft"`
	RequestsLeft  *int    `json:"requestsLeft"`
	TotalBits     *int    `json:"totalBits"`
	TotalRequests *int    `json:"totalRequests"`
}

// mergeUsage merges fields into the cached Usage, tracking whether every
// field required for a complete Usage was present in this response.
func (r *Random) mergeUsage(fields usageFields) {
	r.usageMutex.Lock()
	defer r.usageMutex.Unlock()

	usage := r.usage
	if usage == nil {
		usage = &Usage{}
	}

	isComplete := true

	if fields.Status != nil {
		usage.Status = *fields.Status
	} else {
		isComplete = false
	}

	if fields.CreationTime != nil {
		// fix so that we can parse it
		creationTimeString := strings.Replace(*fields.CreationTime, " ", "T", 1)
		creationTime, err := time.Parse(iso8601Example, creationTimeString)
		if err == nil {
			usage.CreationTime = creationTime
		} else {
			isComplete = false
		}
	} else {
		isComplete = false
	}

	if fields.BitsLeft != nil {
		usage.BitsLeft = *fields.BitsLeft
	} else {
		isComplete = false
	}

	if fields.RequestsLeft != nil {
		usage.RequestsLeft = *fields.RequestsLeft
	} else {
		isComplete = false
	}

	if fields.TotalBits != nil {
		usage.TotalBits = *fields.TotalBits
	} else {
		isComplete = false
	}

	if fields.TotalRequests != nil {
		usage.TotalRequests = *fields.TotalRequests
	} else {
		isComplete = false
	}

	usage.isComplete = isComplete
	r.usage = usage
}

// GetUsage returns information related to the the usage of a given API key.
func (r *Random) GetUsage(ctx context.Context) (Usage, error) {
	params := baseParams{APIKey: r.apiKey}

	fields, err := invokeRequest[usageFields](ctx, r, "getUsage", params)
	if err != nil {
		return Usage{}, err
	}

	r.mergeUsage(fields)

	r.usageMutex.Lock()
	defer r.usageMutex.Unlock()

	return *r.usage, nil
}

// Usage returns the API usage. This will return a cached version of the last request, if there is one.
func (r *Random) Usage(ctx context.Context) (Usage, error) {
	r.usageMutex.Lock()
	cached := r.usage != nil && r.usage.isComplete
	var usage Usage
	if cached {
		usage = *r.usage
	}
	r.usageMutex.Unlock()

	if cached {
		return usage, nil
	}

	return r.GetUsage(ctx)
}
