package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	client "github.com/ONSdigital/dp-search-reindex-api/sdk"
	apiError "github.com/ONSdigital/dp-search-reindex-api/sdk/errors"
)

const (
	service      = "dp-search-reindex-api"
	apiVersion   = "v1"
	jobsEndpoint = "/jobs"
	ETagHeader   = "ETag"
)

type Client struct {
	apiVersion   string
	hcCli        *healthcheck.Client
	serviceToken string
}

// New creates a new instance of Client with a given search reindex api url
func New(searchReindexAPIURL, serviceToken string) *Client {
	return &Client{
		apiVersion:   apiVersion,
		hcCli:        healthcheck.NewClient(service, searchReindexAPIURL),
		serviceToken: serviceToken,
	}
}

// NewWithHealthClient creates a new instance of SearchReindexAPI Client,
// reusing the URL and Clienter from the provided healthcheck client
func NewWithHealthClient(serviceToken string, hcCli *healthcheck.Client) (*Client, error) {
	if hcCli == nil {
		return nil, errors.New("health client is nil")
	}
	return &Client{
		apiVersion:   apiVersion,
		hcCli:        healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
		serviceToken: serviceToken,
	}, nil
}

// URL returns the URL used by this client
func (cli *Client) URL() string {
	return cli.hcCli.URL
}

// HealthClient returns the underlying Healthcheck Client for this search reindex API client
func (cli *Client) Health() *healthcheck.Client {
	return cli.hcCli
}

// Checker calls search reindex api health endpoint and returns a check object to the caller
func (cli *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return cli.hcCli.Checker(ctx, check)
}

// PostJob creates a new reindex job for processing
func (cli *Client) PostJob(ctx context.Context, headers client.Headers) (models.Job, error) {
	var job models.Job
	if headers.ServiceAuthToken == "" {
		headers.ServiceAuthToken = cli.serviceToken
	}

	path := cli.hcCli.URL + "/" + cli.apiVersion + jobsEndpoint
	_, b, err := cli.callReindexAPI(ctx, path, http.MethodPost, headers, nil)
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

// PostTasksCount Updates tasks count for processing
func (cli *Client) PostTasksCount(ctx context.Context, headers client.Headers, jobID string, payload []byte) (string, models.Task, error) {
	var task models.Task

	if headers.ServiceAuthToken == "" {
		headers.ServiceAuthToken = cli.serviceToken
	}

	path := fmt.Sprintf("%s/jobs/%s/tasks", cli.apiVersion, jobID)

	respHeader, b, err := cli.callReindexAPI(ctx, path, http.MethodPost, headers, payload)
	if err != nil {
		return "", task, err
	}

	if err = json.Unmarshal(b, &task); err != nil {
		return "", task, apiError.StatusError{
			Err:  fmt.Errorf("failed to unmarshal bytes into reindex job, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	respETag := respHeader.Get(ETagHeader)

	return respETag, task, nil
}

// PatchJob applies the patch operations, provided in the body, to the job with id = jobID
// It returns the ETag from the response header
func (cli *Client) PatchJob(ctx context.Context, headers client.Headers, jobID string, patchList []client.PatchOperation) (string, error) {
	if headers.ServiceAuthToken == "" {
		headers.ServiceAuthToken = cli.serviceToken
	}

	path := cli.hcCli.URL + "/" + cli.apiVersion + jobsEndpoint + "/" + jobID
	payload, _ := json.Marshal(patchList)

	respHeader, _, err := cli.callReindexAPI(ctx, path, http.MethodPatch, headers, payload)
	if err != nil {
		return "", err
	}

	respETag := respHeader.Get(ETagHeader)

	return respETag, nil
}

// GetTask Get a specific task for a given reindex job
func (cli *Client) GetTask(ctx context.Context, headers client.Headers, jobID, taskName string) (models.Task, error) {
	var task models.Task

	if headers.ServiceAuthToken == "" {
		headers.ServiceAuthToken = cli.serviceToken
	}

	path := fmt.Sprintf("%s/jobs/%s/tasks/%s", cli.apiVersion, jobID, taskName)

	respHeader, b, err := cli.callReindexAPI(ctx, path, http.MethodGet, headers, nil)
	if err != nil {
		return task, err
	}

	if err = json.Unmarshal(b, &task); err != nil {
		return task, apiError.StatusError{
			Err:  fmt.Errorf("failed to unmarshal bytes into reindex job, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	respETag := respHeader.Get(ETagHeader)
	task.MetaData.RespETag = respETag

	return task, nil
}

// callReindexAPI calls the Search Reindex endpoint given by path for the provided REST method, request headers, and body payload.
// It returns the response body and any error that occurred.
func (cli *Client) callReindexAPI(ctx context.Context, path, method string, headers client.Headers, payload []byte) (http.Header, []byte, error) {
	URL, err := url.Parse(path)
	if err != nil {
		return nil, nil, apiError.StatusError{
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
		return nil, nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to create request for call to search reindex api, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	err = headers.Add(req)
	if err != nil {
		return nil, nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to add headers to request, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	resp, err := cli.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to call search reindex api, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}
	defer func() {
		err = closeResponseBody(ctx, resp)
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 400 {
		return nil, nil, apiError.StatusError{
			Err:  fmt.Errorf("failed as unexpected code from search reindex api: %v", resp.StatusCode),
			Code: resp.StatusCode,
		}
	}

	if resp.Body == nil {
		return resp.Header, nil, nil
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.Header, nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to read response body from call to search reindex api, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	return resp.Header, b, nil
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
