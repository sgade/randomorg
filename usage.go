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

func (r *Random) parseAndSaveUsage(json map[string]any) {
	r.usageMutex.Lock()
	defer r.usageMutex.Unlock()

	usage := &Usage{}
	if r.usage != nil {
		usage = r.usage
	}

	isComplete := true

	status, _ := json["status"]
	if status != nil {
		usage.Status = status.(string)
	} else {
		isComplete = false
	}

	creationTimeValue, _ := json["creationTime"]
	if creationTimeValue != nil {
		creationTimeString := creationTimeValue.(string)
		// fix so that we can parse it
		creationTimeString = strings.Replace(creationTimeString, " ", "T", 1)
		creationTime, err := time.Parse(iso8601Example, creationTimeString)
		if err == nil {
			usage.CreationTime = creationTime
		} else {
			isComplete = false
		}
	} else {
		isComplete = false
	}

	bitsLeft, _ := json["bitsLeft"]
	if bitsLeft != nil {
		usage.BitsLeft = int(bitsLeft.(float64))
	} else {
		isComplete = false
	}

	requestsLeft, _ := json["requestsLeft"]
	if requestsLeft != nil {
		usage.RequestsLeft = int(requestsLeft.(float64))
	} else {
		isComplete = false
	}

	totalBits, _ := json["totalBits"]
	if totalBits != nil {
		usage.TotalBits = int(totalBits.(float64))
	} else {
		isComplete = false
	}

	totalRequests, _ := json["totalRequests"]
	if totalRequests != nil {
		usage.TotalRequests = int(totalRequests.(float64))
	} else {
		isComplete = false
	}

	usage.isComplete = isComplete
	r.usage = usage
}

// GetUsage returns information related to the the usage of a given API key.
func (r *Random) GetUsage(ctx context.Context) (Usage, error) {
	params := map[string]any{}

	_, err := r.requestCommand(ctx, "getUsage", params)
	if err != nil && err != ErrJSONFormat {
		return Usage{}, err
	}

	r.usageMutex.Lock()
	defer r.usageMutex.Unlock()

	if r.usage == nil {
		return Usage{}, ErrJSONFormat
	}

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
