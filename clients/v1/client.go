package v1

import (
	"context"
	"fmt"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
)

const service = "dp-search-reindex-api"

type Client struct {
	hcCli *healthcheck.Client
}

// ErrInvalidResponse is returned when dp-search-reindex-api does not respond
// with a valid status
type ErrInvalidResponse struct {
	ActualCode int
	URI        string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidResponse) Error() string {
	return fmt.Sprintf("invalid response from dp-search-reindex-api: %d, path: %s",
		e.ActualCode,
		e.URI,
	)
}

// NewWithSetTimeoutAndMaxRetry creates a new SearchReindexApi Client, with a configurable timeout and maximum number of retries
func NewClientWithClienter(searchReindexURL string, clienter dphttp.Clienter) *Client {
	hcClient := healthcheck.NewClientWithClienter(service, searchReindexURL, clienter)

	return &Client{
		hcClient,
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewClientWithHealthcheck(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// Checker calls search-reindex health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}
