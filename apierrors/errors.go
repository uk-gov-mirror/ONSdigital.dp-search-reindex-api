package apierrors

import (
	"errors"
)

// A list of error messages for JobStore API
var (
	ErrConflictWithJobETag   = errors.New("etag does not match with current state of job resource")
	ErrEmptyTaskNameProvided = errors.New("task name must not be an empty string")
	ErrExistingJobInProgress = errors.New("existing reindex job in progress")
	ErrInternalServer        = errors.New("internal server error")
	ErrInvalidNumTasks       = errors.New("number of tasks must be a positive integer")
	ErrJobNotFound           = errors.New("failed to find the specified reindex job")
	ErrNewETagSame           = errors.New("new etag is same as existing etag")
	ErrTaskInvalidName       = errors.New("task name is not valid")
	ErrUnableToParseJSON     = errors.New("failed to parse json body")
	ErrUnableToReadMessage   = errors.New("failed to read message body")
)
