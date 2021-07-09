package apierrors

import (
	"errors"
)

// A list of error messages for Search Reindex API
var (
	ErrJobNotFound         = errors.New("the job id could not be found in the jobs collection")
	ErrEmptyIDProvided     = errors.New("id must not be an empty string")
	ErrDuplicateIDProvided = errors.New("id must be unique")
)
