package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/ONSdigital/log.go/v2/log"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
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
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// Checker calls search-reindex health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// Get returns a response for the requested uri in searchreindexapi
func (c *Client) Get(ctx context.Context, userAccessToken, path string) ([]byte, error) {
	b, _, err := c.get(ctx, userAccessToken, path)
	return b, err
}

// GetWithHeaders returns a response for the requested uri in searchreindexapi, providing the headers too
func (c *Client) GetWithHeaders(ctx context.Context, userAccessToken, path string) ([]byte, http.Header, error) {
	return c.get(ctx, userAccessToken, path)
}

func (c *Client) get(ctx context.Context, userAccessToken, path string) ([]byte, http.Header, error) {
	req, err := http.NewRequest("GET", c.hcCli.URL+path, http.NoBody)
	if err != nil {
		return nil, nil, err
	}

	dprequest.AddFlorenceHeader(req, userAccessToken)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		return nil, nil, ErrInvalidResponse{resp.StatusCode, req.URL.Path}
	}

	b, err := io.ReadAll(resp.Body)
	return b, resp.Header, err
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}
