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
	"strconv"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	client "github.com/ONSdigital/dp-search-reindex-api/sdk"
	apiError "github.com/ONSdigital/dp-search-reindex-api/sdk/errors"
)

const (
	service      = "dp-search-reindex-api"
	apiVersion   = "v1"
	jobsEndpoint = "/search-reindex-jobs"
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

// Health returns the underlying Healthcheck Client for this search reindex API client
func (cli *Client) Health() *healthcheck.Client {
	return cli.hcCli
}

// Checker calls search reindex api health endpoint and returns a check object to the caller
func (cli *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return cli.hcCli.Checker(ctx, check)
}

// PostJob creates a new reindex job for processing
func (cli *Client) PostJob(ctx context.Context, reqHeaders client.Headers) (*client.RespHeaders, *models.Job, error) {
	var job models.Job
	if reqHeaders.ServiceAuthToken == "" {
		reqHeaders.ServiceAuthToken = cli.serviceToken
	}

	path := fmt.Sprintf("%s/%s"+jobsEndpoint, cli.hcCli.URL, cli.apiVersion)
	respHeader, b, err := cli.callReindexAPI(ctx, path, http.MethodPost, reqHeaders, nil)
	if err != nil {
		return nil, nil, err
	}

	if err = json.Unmarshal(b, &job); err != nil {
		return nil, nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to unmarshal bytes into reindex job, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	respHeaders := client.RespHeaders{
		ETag: respHeader.Get(ETagHeader),
	}
	return &respHeaders, &job, nil
}

// PostTask creates or updates a task, for the job with id = jobID, containing the number of documents to be processed
func (cli *Client) PostTask(ctx context.Context, reqHeaders client.Headers, jobID string, taskToCreate models.TaskToCreate) (*client.RespHeaders, *models.Task, error) {
	if reqHeaders.ServiceAuthToken == "" {
		reqHeaders.ServiceAuthToken = cli.serviceToken
	}

	path := fmt.Sprintf("%s/%s"+jobsEndpoint+"/%s/tasks", cli.hcCli.URL, cli.apiVersion, jobID)
	payload, errMarshal := json.Marshal(taskToCreate)
	if errMarshal != nil {
		return nil, nil, errMarshal
	}

	respHeader, b, err := cli.callReindexAPI(ctx, path, http.MethodPost, reqHeaders, payload)
	if err != nil {
		return nil, nil, err
	}

	var task models.Task
	if err = json.Unmarshal(b, &task); err != nil {
		return nil, nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to unmarshal bytes into reindex job, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	respHeaders := client.RespHeaders{
		ETag: respHeader.Get(ETagHeader),
	}
	return &respHeaders, &task, nil
}

// PatchJob applies the patch operations, provided in the body, to the job with id = jobID
// It returns the ETag from the response header
func (cli *Client) PatchJob(ctx context.Context, reqHeaders client.Headers, jobID string, patchList []client.PatchOperation) (*client.RespHeaders, error) {
	if reqHeaders.ServiceAuthToken == "" {
		reqHeaders.ServiceAuthToken = cli.serviceToken
	}

	path := fmt.Sprintf("%s/%s"+jobsEndpoint+"/%s", cli.hcCli.URL, cli.apiVersion, jobID)
	payload, _ := json.Marshal(patchList)

	respHeader, _, err := cli.callReindexAPI(ctx, path, http.MethodPatch, reqHeaders, payload)
	if err != nil {
		return nil, err
	}

	respHeaders := client.RespHeaders{
		ETag: respHeader.Get(ETagHeader),
	}

	return &respHeaders, nil
}

// GetTask Get a specific task for a given reindex job
func (cli *Client) GetTask(ctx context.Context, reqHeaders client.Headers, jobID, taskName string) (*client.RespHeaders, *models.Task, error) {
	if reqHeaders.ServiceAuthToken == "" {
		reqHeaders.ServiceAuthToken = cli.serviceToken
	}

	path := fmt.Sprintf("%s/%s/search-reindex-jobs/%s/tasks/%s", cli.hcCli.URL, cli.apiVersion, jobID, taskName)

	respHeader, b, err := cli.callReindexAPI(ctx, path, http.MethodGet, reqHeaders, nil)
	if err != nil {
		return nil, nil, err
	}

	var task models.Task

	if err = json.Unmarshal(b, &task); err != nil {
		return nil, nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to unmarshal bytes into reindex task, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	respHeaders := client.RespHeaders{
		ETag: respHeader.Get(ETagHeader),
	}

	return &respHeaders, &task, nil
}

// GetTasks Get all tasks for a given reindex job
func (cli *Client) GetTasks(ctx context.Context, reqHeaders client.Headers, jobID string) (*client.RespHeaders, *models.Tasks, error) {
	if reqHeaders.ServiceAuthToken == "" {
		reqHeaders.ServiceAuthToken = cli.serviceToken
	}

	path := fmt.Sprintf("%s/%s/search-reindex-jobs/%s/tasks", cli.hcCli.URL, cli.apiVersion, jobID)

	respHeader, b, err := cli.callReindexAPI(ctx, path, http.MethodGet, reqHeaders, nil)
	if err != nil {
		return nil, nil, err
	}

	var tasks models.Tasks

	if err = json.Unmarshal(b, &tasks); err != nil {
		return nil, nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to unmarshal bytes into reindex tasks, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	respHeaders := client.RespHeaders{
		ETag: respHeader.Get(ETagHeader),
	}

	return &respHeaders, &tasks, nil
}

// GetJob Get the specific search reindex job that has the id given in the path.
func (cli *Client) GetJob(ctx context.Context, reqheader client.Headers, jobID string) (*client.RespHeaders, *models.Job, error) {
	if reqheader.ServiceAuthToken == "" {
		reqheader.ServiceAuthToken = cli.serviceToken
	}

	path := fmt.Sprintf("%s/%s/search-reindex-jobs/%s", cli.hcCli.URL, cli.apiVersion, jobID)

	respHeader, b, err := cli.callReindexAPI(ctx, path, http.MethodGet, reqheader, nil)
	if err != nil {
		return nil, nil, err
	}

	var job models.Job

	if err = json.Unmarshal(b, &job); err != nil {
		return nil, nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to unmarshal bytes into reindex job, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	respHeaders := client.RespHeaders{
		ETag: respHeader.Get(ETagHeader),
	}

	return &respHeaders, &job, nil
}

func (cli *Client) GetJobs(ctx context.Context, reqheader client.Headers, options client.Options) (*client.RespHeaders, *models.Jobs, error) {
	if reqheader.ServiceAuthToken == "" {
		reqheader.ServiceAuthToken = cli.serviceToken
	}

	validOffset, err := cli.ValidateOptions(options.Offset)
	if err != nil {
		return nil, nil, err
	}

	validLimit, err := cli.ValidateOptions(options.Limit)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("%s/%s/search-reindex-jobs?offset=%s&limit=%s&sort=last_updated", cli.hcCli.URL, cli.apiVersion, validOffset, validLimit)

	respHeader, b, err := cli.callReindexAPI(ctx, path, http.MethodGet, reqheader, nil)
	if err != nil {
		return nil, nil, err
	}

	var jobs models.Jobs

	if err = json.Unmarshal(b, &jobs); err != nil {
		return nil, nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to unmarshal bytes into reindex jobs, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}

	respHeaders := client.RespHeaders{
		ETag: respHeader.Get(ETagHeader),
	}

	return &respHeaders, &jobs, nil
}

// PutTaskNumberOfDocs updates the number of documents, with the provided count, for a task associated with the job specified by the jobID.
func (cli *Client) PutTaskNumberOfDocs(ctx context.Context, reqHeaders client.Headers, jobID, taskName, docCount string) (*client.RespHeaders, error) {
	if reqHeaders.ServiceAuthToken == "" {
		reqHeaders.ServiceAuthToken = cli.serviceToken
	}

	path := fmt.Sprintf("%s/%s"+jobsEndpoint+"/%s/tasks/%s/number-of-documents/%s", cli.hcCli.URL, cli.apiVersion, jobID, taskName, docCount)

	respHeader, _, err := cli.callReindexAPI(ctx, path, http.MethodPut, reqHeaders, nil)
	if err != nil {
		return nil, err
	}

	respHeaders := client.RespHeaders{
		ETag: respHeader.Get(ETagHeader),
	}
	return &respHeaders, nil
}

// PutJobNumberOfTasks updates the number of tasks field, with the provided count, for the job specified by the provided jobID.
func (cli *Client) PutJobNumberOfTasks(ctx context.Context, reqHeaders client.Headers, jobID, numTasks string) (*client.RespHeaders, error) {
	if reqHeaders.ServiceAuthToken == "" {
		reqHeaders.ServiceAuthToken = cli.serviceToken
	}

	path := fmt.Sprintf("%s/%s"+jobsEndpoint+"/%s/number-of-tasks/%s", cli.hcCli.URL, cli.apiVersion, jobID, numTasks)

	respHeader, _, err := cli.callReindexAPI(ctx, path, http.MethodPut, reqHeaders, nil)
	if err != nil {
		return nil, err
	}

	respHeaders := client.RespHeaders{
		ETag: respHeader.Get(ETagHeader),
	}
	return &respHeaders, nil
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

func (cli *Client) ValidateOptions(option int) (validOption string, err error) {
	if option < 0 {
		return "", apiError.StatusError{
			Err: fmt.Errorf("failed to validate option: %v", err),
		}
	}
	return strconv.Itoa(option), nil
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
