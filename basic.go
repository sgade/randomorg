package randomOrg

// Basic commands
// see https://api.random.org/json-rpc/1/basic

// RequestCommand invokes the request and parses all information down to the requested data block.
func (r *RandomOrg) requestCommand(method string, params map[string]interface{}) ([]interface{}, error) {
  result, err := r.invokeRequest(method, params)
  if err != nil {
    return nil, err
  }
  random, err := r.jsonMap(result, "random")
  if err != nil {
    return nil, err
  }

  data := random["data"].([]interface{})

  return data, nil
}

// Generate n number of random integers in the range from min to max.
func (r *RandomOrg) GenerateIntegers(n int, min, max int64) ([]int64, error) {
  if ( n < 1 || n > 1e4 ) {
    return nil, ErrParamRage
  }
  if ( min < -1e9 || min > 1e9 || max < -1e9 || max > 1e9 ) {
    return nil, ErrParamRage
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
