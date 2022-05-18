package pagination

import (
	"errors"
	"strconv"
)

var (
	// ErrInvalidOffsetParameter represents an error case where an invalid offset value is provided
	ErrInvalidOffsetParameter = errors.New("invalid offset query parameter")

	// ErrInvalidLimitParameter represents an error case where an invalid limit value is provided
	ErrInvalidLimitParameter = errors.New("invalid limit query parameter")

	// ErrLimitOverMax represents an error case where the given limit value is larger than the maximum allowed
	ErrLimitOverMax = errors.New("limit query parameter is larger than the maximum allowed")
)

// Paginator is a type to hold pagination related defaults, and provides helper functions using the defaults if needed
type Paginator struct {
	DefaultLimit    int
	DefaultOffset   int
	DefaultMaxLimit int
}

// PaginatedResponse represents the pagination related values that go into list based response
type PaginatedResponse struct {
	Count      int `json:"count"`
	Offset     int `json:"offset"`
	Limit      int `json:"limit"`
	TotalCount int `json:"total_count"`
}

// NewPaginator creates a new instance
func NewPaginator(defaultLimit, defaultOffset, defaultMaxLimit int) *Paginator {
	return &Paginator{
		DefaultLimit:    defaultLimit,
		DefaultOffset:   defaultOffset,
		DefaultMaxLimit: defaultMaxLimit,
	}
}

// ValidateParameters returns pagination related values based on the given request
func (p *Paginator) ValidateParameters(offsetParameter, limitParameter string) (offset, limit int, err error) {
	offset = p.DefaultOffset
	limit = p.DefaultLimit

	if offsetParameter != "" {
		offset, err = strconv.Atoi(offsetParameter)
		if err != nil || offset < 0 {
			return 0, 0, ErrInvalidOffsetParameter
		}
	}

	if limitParameter != "" {
		limit, err = strconv.Atoi(limitParameter)
		if err != nil || limit < 0 {
			return 0, 0, ErrInvalidLimitParameter
		}
	}

	if limit > p.DefaultMaxLimit {
		return 0, 0, ErrLimitOverMax
	}

	return
}
