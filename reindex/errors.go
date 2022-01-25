package reindex

import (
	"errors"
)

// A list of error messages for creating a new index
var (
	ErrReadBodyFailed      = errors.New("failed to read response body")
	ErrUnmarshalBodyFailed = errors.New("failed to unmarshal response body")
	ErrSettingServiceAuth  = errors.New("error setting service auth token")
	ErrResponseBodyEmpty   = errors.New("response body empty")
)
