package httpx

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrForbidden       = errors.New("forbidden")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidInput    = errors.New("invalid input")
	ErrConflict        = errors.New("conflict")
	ErrNotImplemented  = errors.New("not implemented")
	ErrSourceDisabled  = errors.New("source disabled")
	ErrFeedFetchFailed = errors.New("feed fetch failed")
)
