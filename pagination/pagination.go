package pagination

import (
	"strconv"

	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/dp-search-reindex-api/config"
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

// ValidateParameters returns pagination related values based on the given request.
func (p *Paginator) ValidateParameters(offsetParameter, limitParameter string) (offset, limit int, err error) {
	offset = p.DefaultOffset
	limit = p.DefaultLimit

	if offsetParameter != "" {
		offset, err = strconv.Atoi(offsetParameter)
		if err != nil || offset < 0 {
			return 0, 0, apierrors.ErrInvalidOffsetParameter
		}
	}

	if limitParameter != "" {
		limit, err = strconv.Atoi(limitParameter)
		if err != nil || limit < 0 {
			return 0, 0, apierrors.ErrInvalidLimitParameter
		}
	}

	if limit > p.DefaultMaxLimit {
		return 0, 0, apierrors.ErrLimitOverMax
	}

	return
}

// InitialisePagination creates a Paginator and uses it to validate, and set, the offset and limit parameters.
func InitialisePagination(cfg *config.Config, offsetParam, limitParam string) (offset, limit int, err error) {
	paginator := NewPaginator(cfg.DefaultLimit, cfg.DefaultOffset, cfg.DefaultMaxLimit)
	return paginator.ValidateParameters(offsetParam, limitParam)
}
