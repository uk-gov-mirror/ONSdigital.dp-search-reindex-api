package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/log.go/v2/log"
)

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}

// GetValueType gets the type of a value and returns the type which is understandable for an API user
func GetValueType(value interface{}) string {
	valueType := fmt.Sprintf("%T", value)

	switch valueType {
	case "bool":
		return "boolean"
	case "float32", "float64", "int", "int8", "int16", "int32", "int64":
		return "integer"
	default:
		return valueType
	}
}

// ReadJSONBody reads the bytes from the provided body, and marshals it to the provided model interface.
func ReadJSONBody(body io.ReadCloser, v interface{}) error {
	defer body.Close()

	// Get Body bytes
	payload, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("%s: %w", apierrors.ErrUnableToReadMessage, err)
	}

	// Unmarshal body bytes to model
	if err := json.Unmarshal(payload, v); err != nil {
		return fmt.Errorf("%s: %w", apierrors.ErrUnableToParseJSON, err)
	}

	return nil
}
