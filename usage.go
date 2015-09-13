/*
 * Copyright 2015 Sören Gade
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package randomorg

import (
	"strings"
	"time"
)

// Information related to the the usage of a given API key.
type Usage struct {
	// A string indicating the API key's current status, which may be stopped, paused or running. An API key must be running for it to be able to serve requests.
	Status string
	// A timestamp at which the API key was created.
	CreationTime time.Time
	// An integer containing the (estimated) number of remaining true random bits available to the client.
	BitsLeft int
	// An integer containing the (estimated) number of remaining API requests available to the client.
	RequestLeft int
	// An integer containing the number of bits used by this API key since it was created.
	TotalBits int
	// An integer containing the number of requests used by this API key since it was created.
	TotalRequests int
	// Defines if this instance contains all information.
	isComplete bool
}

func (r *Random) parseAndSaveUsage(json map[string]interface{}) {
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
			panic(err)
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
		usage.RequestLeft = int(requestsLeft.(float64))
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
func (r *Random) GetUsage() (Usage, error) {
	params := map[string]interface{}{}

	_, err := r.requestCommand("getUsage", params)
	if err != nil && err != ErrJsonFormat {
		return Usage{}, err
	}

	return r.Usage()
}

// Returns the API usage. This will return a cached version of the last request, if there is one.
func (r *Random) Usage() (Usage, error) {
	if r.usage != nil && r.usage.isComplete {
		return *r.usage, nil
	}

	return r.GetUsage()
}
