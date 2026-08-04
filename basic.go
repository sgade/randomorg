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

// PregeneratedRandomization selects historical, pregenerated randomness for
// a generate call instead of fresh, one-time randomness. Construct one with
// PregeneratedRandomizationByDate or PregeneratedRandomizationByID. A nil
// *PregeneratedRandomization (the default on every Options struct) requests
// fresh randomness, which RANDOM.ORG discards immediately after use.
//
// The non-nil forms turn RANDOM.ORG into a deterministic pseudo-random
// number generator: the same request replays the same values, which is
// useful for reproducing a draw or letting multiple parties derive the same
// values independently.
type PregeneratedRandomization struct {
	Date string `json:"date,omitempty"`
	ID   string `json:"id,omitempty"`
}

// PregeneratedRandomizationByDate uses the historical true randomness
// RANDOM.ORG generated on date (an ISO 8601 "YYYY-MM-DD" string), which must
// be today or in the past in UTC.
func PregeneratedRandomizationByDate(date string) *PregeneratedRandomization {
	return &PregeneratedRandomization{Date: date}
}

// PregeneratedRandomizationByID uses historical true randomness derived
// deterministically from id, a persistent identifier of length [1, 64].
// Requesting the same id again reproduces the same values.
func PregeneratedRandomizationByID(id string) *PregeneratedRandomization {
	return &PregeneratedRandomization{ID: id}
}

// validatePregeneratedRandomization checks the constraints the API
// documentation places on PregeneratedRandomization.ID; Date is validated
// server-side since it must be compared against the current UTC date.
func validatePregeneratedRandomization(p *PregeneratedRandomization) error {
	if p == nil {
		return nil
	}
	if p.ID != "" && len(p.ID) > 64 {
		return ErrParamRange
	}
	return nil
}

// validateSequenceParams validates the parameters shared by
// GenerateIntegerSequences and GenerateSignedIntegerSequences: n sequences,
// each with its own length, lower bound and upper bound.
func validateSequenceParams(n int, length []int, min, max []int64) error {
	if n < 1 || n > 1_000 {
		return ErrParamRange
	}
	if len(length) != n || len(min) != n || len(max) != n {
		return ErrParamRange
	}

	sum := 0
	for i := range length {
		if length[i] < 1 || length[i] > 10_000 {
			return ErrParamRange
		}
		sum += length[i]

		if min[i] < -1_000_000_000 || min[i] > 1_000_000_000 {
			return ErrParamRange
		}
		if max[i] < -1_000_000_000 || max[i] > 1_000_000_000 {
			return ErrParamRange
		}
	}
	if sum < 1 || sum > 10_000 {
		return ErrParamRange
	}

	return nil
}

// Blob encoding formats accepted by GenerateBlobs and GenerateSignedBlobs.
const (
	BlobFormatBase64 = "base64"
	BlobFormatHex    = "hex"
)

// GenerateIntegersOptions holds optional parameters for GenerateIntegers.
type GenerateIntegersOptions struct {
	// Replacement specifies whether the numbers are picked with
	// replacement. nil (the default) behaves like true: the result may
	// contain duplicate values. Set to Bool(false) to draw unique values.
	Replacement *bool
	// PregeneratedRandomization selects historical randomness instead of
	// fresh, on-the-fly randomness. nil (the default) uses fresh randomness.
	PregeneratedRandomization *PregeneratedRandomization
}

type generateIntegersParams struct {
	baseParams
	N                         int                        `json:"n"`
	Min                       int64                      `json:"min"`
	Max                       int64                      `json:"max"`
	Replacement               *bool                      `json:"replacement,omitempty"`
	PregeneratedRandomization *PregeneratedRandomization `json:"pregeneratedRandomization,omitempty"`
}

// GenerateIntegers generates n number of random integers in the range from min to max.
func (r *Random) GenerateIntegers(ctx context.Context, n int, min, max int64, opts ...GenerateIntegersOptions) ([]int64, error) {
	if n < 1 || n > 10_000 {
		return nil, ErrParamRange
	}
	if min < -1_000_000_000 || min > 1_000_000_000 || max < -1_000_000_000 || max > 1_000_000_000 {
		return nil, ErrParamRange
	}

	o := resolveOptions(opts)
	if err := validatePregeneratedRandomization(o.PregeneratedRandomization); err != nil {
		return nil, err
	}

	params := generateIntegersParams{
		baseParams:                baseParams{APIKey: r.apiKey},
		N:                         n,
		Min:                       min,
		Max:                       max,
		Replacement:               o.Replacement,
		PregeneratedRandomization: o.PregeneratedRandomization,
	}

	return generate[int64](ctx, r, "generateIntegers", params)
}

// GenerateIntegerSequencesOptions holds optional parameters for GenerateIntegerSequences.
type GenerateIntegerSequencesOptions struct {
	// Replacement specifies, per sequence, whether its numbers are picked
	// with replacement. If non-nil, it must have exactly n elements. nil
	// (the default) behaves as if every sequence used replacement.
	Replacement []bool
	// PregeneratedRandomization selects historical randomness instead of
	// fresh, on-the-fly randomness, for every sequence. nil (the default)
	// uses fresh randomness.
	PregeneratedRandomization *PregeneratedRandomization
}

type generateIntegerSequencesParams struct {
	baseParams
	N                         int                        `json:"n"`
	Length                    []int                      `json:"length"`
	Min                       []int64                    `json:"min"`
	Max                       []int64                    `json:"max"`
	Replacement               []bool                     `json:"replacement,omitempty"`
	PregeneratedRandomization *PregeneratedRandomization `json:"pregeneratedRandomization,omitempty"`
}

// GenerateIntegerSequences generates n sequences of random integers, where
// sequence i has length[i] elements drawn from the range [min[i], max[i]].
// length, min and max must each have exactly n elements; to request
// identical sequences, repeat the same length/min/max in every slot.
func (r *Random) GenerateIntegerSequences(ctx context.Context, n int, length []int, min, max []int64, opts ...GenerateIntegerSequencesOptions) ([][]int64, error) {
	if err := validateSequenceParams(n, length, min, max); err != nil {
		return nil, err
	}

	o := resolveOptions(opts)
	if o.Replacement != nil && len(o.Replacement) != n {
		return nil, ErrParamRange
	}
	if err := validatePregeneratedRandomization(o.PregeneratedRandomization); err != nil {
		return nil, err
	}

	params := generateIntegerSequencesParams{
		baseParams:                baseParams{APIKey: r.apiKey},
		N:                         n,
		Length:                    length,
		Min:                       min,
		Max:                       max,
		Replacement:               o.Replacement,
		PregeneratedRandomization: o.PregeneratedRandomization,
	}

	return generate[[]int64](ctx, r, "generateIntegerSequences", params)
}

// GenerateDecimalFractionsOptions holds optional parameters for GenerateDecimalFractions.
type GenerateDecimalFractionsOptions struct {
	// Replacement specifies whether the numbers are picked with
	// replacement. nil (the default) behaves like true.
	Replacement *bool
	// PregeneratedRandomization selects historical randomness instead of
	// fresh, on-the-fly randomness. nil (the default) uses fresh randomness.
	PregeneratedRandomization *PregeneratedRandomization
}

type generateDecimalFractionsParams struct {
	baseParams
	N                         int                        `json:"n"`
	DecimalPlaces             int                        `json:"decimalPlaces"`
	Replacement               *bool                      `json:"replacement,omitempty"`
	PregeneratedRandomization *PregeneratedRandomization `json:"pregeneratedRandomization,omitempty"`
}

// GenerateDecimalFractions generates n number of decimal fractions with decimalPlaces number of decimal places.
func (r *Random) GenerateDecimalFractions(ctx context.Context, n, decimalPlaces int, opts ...GenerateDecimalFractionsOptions) ([]float64, error) {
	if n < 1 || n > 10_000 {
		return nil, ErrParamRange
	}
	if decimalPlaces < 1 || decimalPlaces > 14 {
		return nil, ErrParamRange
	}

	o := resolveOptions(opts)
	if err := validatePregeneratedRandomization(o.PregeneratedRandomization); err != nil {
		return nil, err
	}

	params := generateDecimalFractionsParams{
		baseParams:                baseParams{APIKey: r.apiKey},
		N:                         n,
		DecimalPlaces:             decimalPlaces,
		Replacement:               o.Replacement,
		PregeneratedRandomization: o.PregeneratedRandomization,
	}

	return generate[float64](ctx, r, "generateDecimalFractions", params)
}

// GenerateGaussiansOptions holds optional parameters for GenerateGaussians.
type GenerateGaussiansOptions struct {
	// PregeneratedRandomization selects historical randomness instead of
	// fresh, on-the-fly randomness. nil (the default) uses fresh randomness.
	PregeneratedRandomization *PregeneratedRandomization
}

type generateGaussiansParams struct {
	baseParams
	N                         int                        `json:"n"`
	Mean                      float64                    `json:"mean"`
	StandardDeviation         float64                    `json:"standardDeviation"`
	SignificantDigits         int                        `json:"significantDigits"`
	PregeneratedRandomization *PregeneratedRandomization `json:"pregeneratedRandomization,omitempty"`
}

// GenerateGaussians generates true random numbers from a Gaussian distribution.
func (r *Random) GenerateGaussians(ctx context.Context, n int, mean, standardDeviation float64, significantDigits int, opts ...GenerateGaussiansOptions) ([]float64, error) {
	if n < 1 || n > 10_000 {
		return nil, ErrParamRange
	}
	if mean < -1_000_000 || mean > 1_000_000 {
		return nil, ErrParamRange
	}
	if standardDeviation < -1_000_000 || standardDeviation > 1_000_000 {
		return nil, ErrParamRange
	}
	if significantDigits < 2 || significantDigits > 14 {
		return nil, ErrParamRange
	}

	o := resolveOptions(opts)
	if err := validatePregeneratedRandomization(o.PregeneratedRandomization); err != nil {
		return nil, err
	}

	params := generateGaussiansParams{
		baseParams:                baseParams{APIKey: r.apiKey},
		N:                         n,
		Mean:                      mean,
		StandardDeviation:         standardDeviation,
		SignificantDigits:         significantDigits,
		PregeneratedRandomization: o.PregeneratedRandomization,
	}

	return generate[float64](ctx, r, "generateGaussians", params)
}

// GenerateStringsOptions holds optional parameters for GenerateStrings.
type GenerateStringsOptions struct {
	// Replacement specifies whether the strings are picked with
	// replacement. nil (the default) behaves like true.
	Replacement *bool
	// PregeneratedRandomization selects historical randomness instead of
	// fresh, on-the-fly randomness. nil (the default) uses fresh randomness.
	PregeneratedRandomization *PregeneratedRandomization
}

type generateStringsParams struct {
	baseParams
	N                         int                        `json:"n"`
	Length                    int                        `json:"length"`
	Characters                string                     `json:"characters"`
	Replacement               *bool                      `json:"replacement,omitempty"`
	PregeneratedRandomization *PregeneratedRandomization `json:"pregeneratedRandomization,omitempty"`
}

// GenerateStrings generates n random strings with the given length composed from the characters.
func (r *Random) GenerateStrings(ctx context.Context, n, length int, characters string, opts ...GenerateStringsOptions) ([]string, error) {
	if n < 1 || n > 10_000 {
		return nil, ErrParamRange
	}
	if length < 1 || length > 32 {
		return nil, ErrParamRange
	}
	if len(characters) < 1 || len(characters) > 128 {
		return nil, ErrParamRange
	}

	o := resolveOptions(opts)
	if err := validatePregeneratedRandomization(o.PregeneratedRandomization); err != nil {
		return nil, err
	}

	params := generateStringsParams{
		baseParams:                baseParams{APIKey: r.apiKey},
		N:                         n,
		Length:                    length,
		Characters:                characters,
		Replacement:               o.Replacement,
		PregeneratedRandomization: o.PregeneratedRandomization,
	}

	return generate[string](ctx, r, "generateStrings", params)
}

// GenerateUUIDsOptions holds optional parameters for GenerateUUIDs.
type GenerateUUIDsOptions struct {
	// PregeneratedRandomization selects historical randomness instead of
	// fresh, on-the-fly randomness. nil (the default) uses fresh randomness.
	PregeneratedRandomization *PregeneratedRandomization
}

type generateUUIDsParams struct {
	baseParams
	N                         int                        `json:"n"`
	PregeneratedRandomization *PregeneratedRandomization `json:"pregeneratedRandomization,omitempty"`
}

// GenerateUUIDs generates n random version 4 Universally Unique Identifiers (see section 4.4 of RFC 4122)
func (r *Random) GenerateUUIDs(ctx context.Context, n int, opts ...GenerateUUIDsOptions) ([]string, error) {
	if n < 1 || n > 1_000 {
		return nil, ErrParamRange
	}

	o := resolveOptions(opts)
	if err := validatePregeneratedRandomization(o.PregeneratedRandomization); err != nil {
		return nil, err
	}

	params := generateUUIDsParams{
		baseParams:                baseParams{APIKey: r.apiKey},
		N:                         n,
		PregeneratedRandomization: o.PregeneratedRandomization,
	}

	return generate[string](ctx, r, "generateUUIDs", params)
}

// GenerateBlobsOptions holds optional parameters for GenerateBlobs.
type GenerateBlobsOptions struct {
	// Format specifies the encoding used for the returned blobs:
	// BlobFormatBase64 (the default) or BlobFormatHex.
	Format string
	// PregeneratedRandomization selects historical randomness instead of
	// fresh, on-the-fly randomness. nil (the default) uses fresh randomness.
	PregeneratedRandomization *PregeneratedRandomization
}

type generateBlobsParams struct {
	baseParams
	N                         int                        `json:"n"`
	Size                      int                        `json:"size"`
	Format                    string                     `json:"format,omitempty"`
	PregeneratedRandomization *PregeneratedRandomization `json:"pregeneratedRandomization,omitempty"`
}

// GenerateBlobs generates n random blobs of size (in bits, must be divisible by 8).
// The total size of all blobs requested (n*size) must not exceed 1,048,576 bits.
func (r *Random) GenerateBlobs(ctx context.Context, n, size int, opts ...GenerateBlobsOptions) ([]string, error) {
	if n < 1 || n > 100 {
		return nil, ErrParamRange
	}
	if size < 1 || size > 1_048_576 || size%8 != 0 {
		return nil, ErrParamRange
	}
	if n*size > 1_048_576 {
		return nil, ErrParamRange
	}

	o := resolveOptions(opts)
	if o.Format != "" && o.Format != BlobFormatBase64 && o.Format != BlobFormatHex {
		return nil, ErrParamRange
	}
	if err := validatePregeneratedRandomization(o.PregeneratedRandomization); err != nil {
		return nil, err
	}

	params := generateBlobsParams{
		baseParams:                baseParams{APIKey: r.apiKey},
		N:                         n,
		Size:                      size,
		Format:                    o.Format,
		PregeneratedRandomization: o.PregeneratedRandomization,
	}

	return generate[string](ctx, r, "generateBlobs", params)
}
