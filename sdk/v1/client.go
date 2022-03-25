package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	client "github.com/ONSdigital/dp-search-reindex-api/sdk"
	apiError "github.com/ONSdigital/dp-search-reindex-api/sdk/errors"
)

const (
	service    = "dp-search-reindex-api"
	apiVersion = "v1"

	jobsEndpoint = "/jobs"
)

type Client struct {
	apiVersion   string
	hcCli        *healthcheck.Client
	serviceToken string
}

// NewWithSetTimeoutAndMaxRetry creates a new SearchReindexApi Client, with a configurable timeout and maximum number of retries
func NewClientWithClienter(serviceToken, searchReindexURL string, clienter dphttp.Clienter) *Client {
	hcClient := healthcheck.NewClientWithClienter(service, searchReindexURL, clienter)

	return &Client{
		apiVersion:   apiVersion,
		hcCli:        hcClient,
		serviceToken: serviceToken,
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewClientWithHealthcheck(serviceToken string, hcCli *healthcheck.Client) *Client {
	return &Client{
		apiVersion:   apiVersion,
		hcCli:        healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
		serviceToken: serviceToken,
	}
}

// Checker calls search-reindex health endpoint and returns a check object to the caller.
func (cli *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return cli.hcCli.Checker(ctx, check)
}

// PostJob creates a new reindex job for processing
func (cli *Client) PostJob(ctx context.Context, headers client.Headers) (models.Job, error) {
	var job models.Job
	if headers.ServiceAuthToken == "" {
		headers.ServiceAuthToken = cli.serviceToken
	}

    path := cli.hcCli.URL + jobsEndpoint
	b, err := cli.callReindexAPI(ctx, path, http.MethodPost, headers, nil)
	if err != nil {
		return job, err
	}

	if err = json.Unmarshal(b, &job); err != nil {
		return job, apiError.StatusError{
			Err:  fmt.Errorf("failed to unmarshal bytes into reindex job, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	return job, nil
}

func (cli *Client) callReindexAPI(ctx context.Context, path, method string, headers client.Headers, payload []byte) ([]byte, error) {
	URL, err := url.Parse(path)
	if err != nil {
		return nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to parse path: \"%v\" error is: %v", path, err),
			Code: http.StatusInternalServerError,
		}
	}

	path = URL.String()

	var req *http.Request

	if payload != nil {
		req, err = http.NewRequest(method, path, bytes.NewReader(payload))
		req.Header.Add("Content-type", "application/json")
	} else {
		req, err = http.NewRequest(method, path, http.NoBody)
	}

	// check req, above, didn't error
	if err != nil {
		return nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to create request for call to search reindex api, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	headers.Add(req)

	resp, err := cli.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to call search reindex api, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}
	defer func() {
		err = closeResponseBody(ctx, resp)
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 400 {
		return nil, apiError.StatusError{
			Err:  fmt.Errorf("failed as unexpected code from search reindex api: %v", resp.StatusCode),
			Code: resp.StatusCode,
		}
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to read response body from call to search reindex api, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	return b, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) error {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			return apiError.StatusError{
				Err:  fmt.Errorf("error closing http response body from call to search reindex api, error is: %v", err),
				Code: http.StatusInternalServerError,
			}
		}
	}

	return nil
}
