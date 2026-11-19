package auth

import (
	"net/http"
	"strconv"

	"github.com/txix-open/isp-kit/errors"
)

const (
	// ApplicationIdHeader is the header key for application identity.
	ApplicationIdHeader = "x-application-identity"
	// ApplicationNameHeader is the header key for application name.
	ApplicationNameHeader = "x-application-name"
	// UserIdHeader is the header key for user identity.
	UserIdHeader = "x-user-identity"
	// DeviceIdHeader is the header key for device identity.
	DeviceIdHeader = "x-device-identity"
	// ServiceIdHeader is the header key for service identity.
	ServiceIdHeader = "x-service-identity"
	// DomainIdHeader is the header key for domain identity.
	DomainIdHeader = "x-domain-identity"
	// SystemIdHeader is the header key for system identity.
	SystemIdHeader = "x-system-identity"
	// UserTokenHeader is the header key for user token.
	UserTokenHeader = "x-user-token"
	// DeviceTokenHeader is the header key for device token.
	DeviceTokenHeader = "x-device-token"
)

// AuthData wraps http.Header to provide structured access to authentication
// and authorization information from http request header.
// All methods return an error if the key is not found or cannot be parsed.
type AuthData http.Header

// SystemId extracts the system identity from header as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) SystemId() (int, error) {
	return IntFromHeader(SystemIdHeader, http.Header(i))
}

// DomainId extracts the domain identity from header as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) DomainId() (int, error) {
	return IntFromHeader(DomainIdHeader, http.Header(i))
}

// ServiceId extracts the service identity from header as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) ServiceId() (int, error) {
	return IntFromHeader(ServiceIdHeader, http.Header(i))
}

// ApplicationId extracts the application identity from header as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) ApplicationId() (int, error) {
	return IntFromHeader(ApplicationIdHeader, http.Header(i))
}

// ApplicationName extracts the application name from header as a string.
// Returns an error if the key is not found.
func (i AuthData) ApplicationName() (string, error) {
	return StringFromHeader(ApplicationNameHeader, http.Header(i))
}

// UserId extracts the user identity from header as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) UserId() (int, error) {
	return IntFromHeader(UserIdHeader, http.Header(i))
}

// DeviceId extracts the device identity from header as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) DeviceId() (int, error) {
	return IntFromHeader(DeviceIdHeader, http.Header(i))
}

// UserToken extracts the user token from header as a string.
// Returns an error if the key is not found.
func (i AuthData) UserToken() (string, error) {
	return StringFromHeader(UserTokenHeader, http.Header(i))
}

// DeviceToken extracts the device token from header as a string.
// Returns an error if the key is not found.
func (i AuthData) DeviceToken() (string, error) {
	return StringFromHeader(DeviceTokenHeader, http.Header(i))
}

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
