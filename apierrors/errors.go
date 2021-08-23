package apierrors

import (
	"errors"
)

// A list of error messages for JobStore API
var (
	ErrUnableToReadMessage = errors.New("failed to read message body")
	ErrUnableToParseJSON   = errors.New("failed to parse json body")
)
