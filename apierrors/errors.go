package apierrors

import (
	"errors"
)

// A list of general error messages
var (
	ErrConflictWithETag    = errors.New("etag does not match with current state of resource")
	ErrInternalServer      = errors.New("internal server error")
	ErrInvalidRequestBody  = errors.New("invalid request body")
	ErrNewETagSame         = errors.New("no modification made on resource")
	ErrUnableToParseJSON   = errors.New("failed to parse json body")
	ErrUnableToReadMessage = errors.New("failed to read message body")
)

// A list of job error messages
var (
	ErrEmptyJobID            = errors.New("job id must not be an empty string")
	ErrExistingJobInProgress = errors.New("existing reindex job in progress")
	ErrInvalidNumTasks       = errors.New("number of tasks must be a positive integer")
	ErrInvalidNumDocs        = errors.New("number of docs must be a positive integer")
	ErrJobNotFound           = errors.New("failed to find the specified reindex job")
)

// A list of task error messages
var (
	ErrEmptyTaskNameProvided = errors.New("task name must not be an empty string")
	ErrTaskInvalidName       = errors.New("task name is not valid")
	ErrTaskNotFound          = errors.New("failed to find the specified task for the reindex job")
)

// A list of pagnination error messages
var (
	ErrInvalidOffsetParameter = errors.New("invalid offset query parameter")
	ErrInvalidLimitParameter  = errors.New("invalid limit query parameter")
	ErrLimitOverMax           = errors.New("limit query parameter is larger than the maximum allowed")
)
