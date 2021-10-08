package apierrors

import (
	"errors"
)

// A list of error messages for JobStore API
var (
	ErrUnableToReadMessage   = errors.New("failed to read message body")
	ErrUnableToParseJSON     = errors.New("failed to parse json body")
	ErrEmptyTaskNameProvided = errors.New("task name must not be an empty string")
	ErrTaskInvalidName       = errors.New("task name is not valid")
)
