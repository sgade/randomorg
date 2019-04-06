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

// Basic commands
// see https://api.random.org/json-rpc/2/basic

// GenerateIntegers generates n number of random integers in the range from min to max.
func (r *Random) GenerateIntegers(n int, min, max int64) ([]int64, error) {
	if n < 1 || n > 1e4 {
		return nil, ErrParamRange
	}
	if min < -1e9 || min > 1e9 || max < -1e9 || max > 1e9 {
		return nil, ErrParamRange
	}

	params := map[string]interface{}{
		"n":   n,
		"min": min,
		"max": max,
	}

	values, err := r.requestCommand("generateIntegers", params)
	if err != nil {
		return nil, err
	}

	ints := make([]int64, len(values))
	for i, value := range values {
		f := value.(float64)
		ints[i] = int64(f)
	}

	return ints, nil
}

// GenerateDecimalFractions generates n number of decimal fractions with decimalPlaces number of decimal places.
func (r *Random) GenerateDecimalFractions(n, decimalPlaces int) ([]float64, error) {
	if n < 1 || n > 1e4 {
		return nil, ErrParamRange
	}
	if decimalPlaces < 1 || decimalPlaces > 20 {
		return nil, ErrParamRange
	}

	params := map[string]interface{}{
		"n":             n,
		"decimalPlaces": decimalPlaces,
	}

	values, err := r.requestCommand("generateDecimalFractions", params)
	if err != nil {
		return nil, err
	}

	decimals := make([]float64, len(values))
	for i, value := range values {
		decimals[i] = value.(float64)
	}

	return decimals, nil
}

// GenerateGaussians generates true random numbers from a Gaussian distribution.
func (r *Random) GenerateGaussians(n, mean, standardDeviation, significantDigits int) ([]float64, error) {
	if n < 1 || n > 1e4 {
		return nil, ErrParamRange
	}
	if mean < -1e6 || mean > 1e6 {
		return nil, ErrParamRange
	}
	if standardDeviation < -1e6 || standardDeviation > 1e6 {
		return nil, ErrParamRange
	}
	if significantDigits < 2 || significantDigits > 20 {
		return nil, ErrParamRange
	}

	params := map[string]interface{}{
		"n":                 n,
		"mean":              mean,
		"standardDeviation": standardDeviation,
		"significantDigits": significantDigits,
	}

	values, err := r.requestCommand("generateGaussians", params)
	if err != nil {
		return nil, err
	}

	gaussians := make([]float64, len(values))
	for i, value := range values {
		gaussians[i] = value.(float64)
	}

	return gaussians, nil
}

// GenerateStrings generates n random strings with the given length composed from the characters.
func (r *Random) GenerateStrings(n, length int, characters string) ([]string, error) {
	if n < 1 || n > 1e4 {
		return nil, ErrParamRange
	}
	if length < 1 || length > 20 {
		return nil, ErrParamRange
	}
	if len(characters) < 1 || len(characters) > 80 {
		return nil, ErrParamRange
	}

	params := map[string]interface{}{
		"n":          n,
		"length":     length,
		"characters": characters,
	}

	values, err := r.requestCommand("generateStrings", params)
	if err != nil {
		return nil, err
	}

	strings := make([]string, len(values))
	for i, value := range values {
		strings[i] = value.(string)
	}

	return strings, nil
}

// GenerateUUIDs generates n random version 4 Universally Unique Identifiers (see section 4.4 of RFC 4122)
func (r *Random) GenerateUUIDs(n int) ([]string, error) {
	if n < 1 || n > 1e3 {
		return nil, ErrParamRange
	}

	params := map[string]interface{}{
		"n": n,
	}

	values, err := r.requestCommand("generateUUIDs", params)
	if err != nil {
		return nil, err
	}

	uuids := make([]string, len(values))
	for i, value := range values {
		uuids[i] = value.(string)
	}

	return uuids, nil
}

// GenerateBlobs generates n random blobs of size.
func (r *Random) GenerateBlobs(n, size int) ([]string, error) {
	if n < 1 || n > 100 {
		return nil, ErrParamRange
	}
	if size < 1 || size > 1048576 || size%8 != 0 {
		return nil, ErrParamRange
	}

	params := map[string]interface{}{
		"n":    n,
		"size": size,
	}

	values, err := r.requestCommand("generateBlobs", params)
	if err != nil {
		return nil, err
	}

	blobs := make([]string, len(values))
	for i, value := range values {
		blobs[i] = value.(string)
	}

	return blobs, nil
}
