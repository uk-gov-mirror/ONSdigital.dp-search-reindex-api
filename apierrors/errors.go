package apierrors

import (
	"errors"
)

// A list of error messages for JobStore API
var (
	ErrConflictWithJobETag   = errors.New("etag does not match with current state of job resource")
	ErrEmptyTaskNameProvided = errors.New("task name must not be an empty string")
	ErrTaskInvalidName       = errors.New("task name is not valid")
	ErrUnableToParseJSON     = errors.New("failed to parse json body")
	ErrUnableToReadMessage   = errors.New("failed to read message body")
)
