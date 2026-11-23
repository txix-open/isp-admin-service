package domain

import (
	"fmt"

	"github.com/txix-open/isp-kit/errors"
)

const (
	ErrCodeInvalidPassword = 1001
)

var (
	ErrNotFound             = errors.New("not found")
	ErrUnauthenticated      = errors.New("authentication failure")
	ErrOauthAuthorization   = errors.New("user is authorized only with oauth")
	ErrOauthIsMissed        = errors.New("oauth is not configured on the server")
	ErrInvalid              = errors.New("entity is invalid")
	ErrAlreadyExists        = errors.New("already exists")
	ErrTokenExpired         = errors.New("token expired")
	ErrTokenNotFound        = errors.New("token not found")
	ErrTooManyLoginRequests = errors.New("too many login requests")
	ErrUserIsBlocked        = errors.New("user is blocked")
	ErrNoActionRequired     = errors.New("no action required")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrUnknownClient        = errors.New("unknown client")
)

const (
	UnknownAuditEventErrorCode  = 601
	ErrOauthAuthorizationCode   = 602
	ErrUnauthenticatedCode      = 603
	ErrTooManyLoginRequestsCode = 604
	ErrOauthIsMissedCode        = 605
	ErrAlreadyExistsCode        = 606
	ErrNotFoundCode             = 607
	ErrUnknownClientCode        = 608
	ErrInvalidCode              = 609
	ErrInvalidPasswordCode      = 610
)

type UnknownAuditEventError struct {
	Event string
}

func (e UnknownAuditEventError) Error() string {
	return fmt.Sprintf("unknown audit event: %s", e.Event)
}
