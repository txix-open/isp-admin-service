package auth

import (
	"net/http"
	"strconv"

	"github.com/txix-open/isp-kit/errors"
)

// StringFromHeader extracts a string value from header by key.
// Returns an error if header is nil or the key is not found.
func StringFromHeader(key string, h http.Header) (string, error) {
	if h == nil {
		return "", errors.New("header is nil")
	}
	values := h.Values(key)
	if len(values) == 0 {
		return "", errors.Errorf("'%s' is expected in header", key)
	}
	return values[0], nil
}

// IntFromHeader extracts an integer value from header by key.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func IntFromHeader(key string, h http.Header) (int, error) {
	value, err := StringFromHeader(key, h)
	if err != nil {
		return 0, err
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.WithMessagef(err, "parse '%s' to int", key)
	}
	return intValue, nil
}
