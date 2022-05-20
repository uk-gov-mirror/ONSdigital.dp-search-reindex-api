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
	ErrInvalidRequestBody    = errors.New("invalid request body")
	ErrInvalidNumTasks       = errors.New("number of tasks must be a positive integer")
	ErrJobNotFound           = errors.New("failed to find the specified reindex job")
	ErrNewETagSame           = errors.New("no modification made on resource")
	ErrTaskInvalidName       = errors.New("task name is not valid")
	ErrTaskNotFound          = errors.New("failed to find the specified task for the reindex job")
	ErrUnableToParseJSON     = errors.New("failed to parse json body")
	ErrUnableToReadMessage   = errors.New("failed to read message body")
)
