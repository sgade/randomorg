package randomorg

import "context"

// Basic commands
// see https://api.random.org/json-rpc/4/basic

// randomData is the "random" payload returned by every generateX method.
type randomData[T any] struct {
	Data           []T    `json:"data"`
	CompletionTime string `json:"completionTime"`
}

// generateResult is the JSON-RPC result payload of every generateX method.
type generateResult[T any] struct {
	Random randomData[T] `json:"random"`
	usageFields
}

// generate invokes method, decodes its data payload into a []T, and merges
// the response's usage-related fields into the client's usage cache.
func generate[T any](ctx context.Context, r *Random, method string, params any) ([]T, error) {
	result, err := invokeRequest[generateResult[T]](ctx, r, method, params)
	if err != nil {
		return nil, err
	}

	r.mergeUsage(result.usageFields)

	if result.Random.Data == nil {
		return nil, ErrJSONFormat
	}

	return result.Random.Data, nil
}

type generateIntegersParams struct {
	baseParams
	N   int   `json:"n"`
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

// GenerateIntegers generates n number of random integers in the range from min to max.
func (r *Random) GenerateIntegers(ctx context.Context, n int, min, max int64) ([]int64, error) {
	if n < 1 || n > 1e4 {
		return nil, ErrParamRange
	}
	if min < -1e9 || min > 1e9 || max < -1e9 || max > 1e9 {
		return nil, ErrParamRange
	}

	params := generateIntegersParams{
		baseParams: baseParams{APIKey: r.apiKey},
		N:          n,
		Min:        min,
		Max:        max,
	}

	return generate[int64](ctx, r, "generateIntegers", params)
}

type generateDecimalFractionsParams struct {
	baseParams
	N             int `json:"n"`
	DecimalPlaces int `json:"decimalPlaces"`
}

// GenerateDecimalFractions generates n number of decimal fractions with decimalPlaces number of decimal places.
func (r *Random) GenerateDecimalFractions(ctx context.Context, n, decimalPlaces int) ([]float64, error) {
	if n < 1 || n > 1e4 {
		return nil, ErrParamRange
	}
	if decimalPlaces < 1 || decimalPlaces > 20 {
		return nil, ErrParamRange
	}

	params := generateDecimalFractionsParams{
		baseParams:    baseParams{APIKey: r.apiKey},
		N:             n,
		DecimalPlaces: decimalPlaces,
	}

	return generate[float64](ctx, r, "generateDecimalFractions", params)
}

type generateGaussiansParams struct {
	baseParams
	N                 int `json:"n"`
	Mean              int `json:"mean"`
	StandardDeviation int `json:"standardDeviation"`
	SignificantDigits int `json:"significantDigits"`
}

// GenerateGaussians generates true random numbers from a Gaussian distribution.
func (r *Random) GenerateGaussians(ctx context.Context, n, mean, standardDeviation, significantDigits int) ([]float64, error) {
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

	params := generateGaussiansParams{
		baseParams:        baseParams{APIKey: r.apiKey},
		N:                 n,
		Mean:              mean,
		StandardDeviation: standardDeviation,
		SignificantDigits: significantDigits,
	}

	return generate[float64](ctx, r, "generateGaussians", params)
}

type generateStringsParams struct {
	baseParams
	N          int    `json:"n"`
	Length     int    `json:"length"`
	Characters string `json:"characters"`
}

// GenerateStrings generates n random strings with the given length composed from the characters.
func (r *Random) GenerateStrings(ctx context.Context, n, length int, characters string) ([]string, error) {
	if n < 1 || n > 1e4 {
		return nil, ErrParamRange
	}
	if length < 1 || length > 20 {
		return nil, ErrParamRange
	}
	if len(characters) < 1 || len(characters) > 80 {
		return nil, ErrParamRange
	}

	params := generateStringsParams{
		baseParams: baseParams{APIKey: r.apiKey},
		N:          n,
		Length:     length,
		Characters: characters,
	}

	return generate[string](ctx, r, "generateStrings", params)
}

type generateUUIDsParams struct {
	baseParams
	N int `json:"n"`
}

// GenerateUUIDs generates n random version 4 Universally Unique Identifiers (see section 4.4 of RFC 4122)
func (r *Random) GenerateUUIDs(ctx context.Context, n int) ([]string, error) {
	if n < 1 || n > 1e3 {
		return nil, ErrParamRange
	}

	params := generateUUIDsParams{
		baseParams: baseParams{APIKey: r.apiKey},
		N:          n,
	}

	return generate[string](ctx, r, "generateUUIDs", params)
}

type generateBlobsParams struct {
	baseParams
	N    int `json:"n"`
	Size int `json:"size"`
}

// GenerateBlobs generates n random blobs of size.
func (r *Random) GenerateBlobs(ctx context.Context, n, size int) ([]string, error) {
	if n < 1 || n > 100 {
		return nil, ErrParamRange
	}
	if size < 1 || size > 1048576 || size%8 != 0 {
		return nil, ErrParamRange
	}

	params := generateBlobsParams{
		baseParams: baseParams{APIKey: r.apiKey},
		N:          n,
		Size:       size,
	}

	return generate[string](ctx, r, "generateBlobs", params)
}
