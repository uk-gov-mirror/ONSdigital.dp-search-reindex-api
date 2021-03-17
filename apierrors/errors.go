package apierrors

import (
	"errors"
)

// A list of error messages for Image API
var (
	ErrJobNotFound                    = errors.New("job not found")

	ErrUnableToReadMessage              = errors.New("failed to read message body")
	ErrUnableToParseJSON                = errors.New("failed to parse json body")
	
)