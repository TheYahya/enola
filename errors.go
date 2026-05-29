package enola

import "errors"

var (
	ErrDataFileIsNotAValidJson = errors.New("the data file cannot be read due to invalid JSON format")
	ErrSiteNotFound            = errors.New("the requested site is not supported")
	ErrUnknownErrorType        = errors.New("unknown error type")
)
