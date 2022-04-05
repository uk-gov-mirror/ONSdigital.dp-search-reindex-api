package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	client "github.com/ONSdigital/dp-search-reindex-api/sdk"
	apiError "github.com/ONSdigital/dp-search-reindex-api/sdk/errors"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	serviceName   = "test-app"
	serviceToken  = "test-token"
	testHost      = "http://localhost:25700"
	testJobID     = "123"
	ifMatchHeader = "If-Match"
	testETag      = `"56b6890f1321590998d5fd8d293b620581ff3531"`
)

var (
	initialState = healthcheck.CreateCheckState(service)
)

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {
		},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func newSearchReindexClient(t *testing.T, httpClient *dphttp.ClienterMock) *Client {
	healthClient := healthcheck.NewClientWithClienter(serviceName, testHost, httpClient)
	searchReindexClient, err := NewWithHealthClient(serviceToken, healthClient)
	if err != nil {
		t.Errorf("failed to create a search reindex client, error is: %v", err)
	}

	return searchReindexClient
}

func TestClient_HealthChecker(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")
		httpClient := newMockHTTPClient(&http.Response{}, clientError)
		searchReindexClient := newSearchReindexClient(t, httpClient)
		check := initialState

		Convey("when search-reindexClient.Checker is called", func() {
			err := searchReindexClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, health.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Message(), ShouldEqual, clientError.Error())
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given a 500 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		searchReindexClient := newSearchReindexClient(t, httpClient)
		check := initialState

		Convey("when search-reindexClient.Checker is called", func() {
			err := searchReindexClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, health.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Message(), ShouldEqual, service+healthcheck.StatusMessage[health.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func TestClient_PostJob(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	path := "/jobs"

	Convey("Given clienter.Do doesn't return an error", t, func() {
		expectedJob := models.Job{
			ID:          testJobID,
			LastUpdated: time.Now(),
			Links: &models.JobLinks{
				Tasks: "/v1/jobs/" + testJobID + "/tasks",
				Self:  "/v1/jobs/" + testJobID,
			},
			NumberOfTasks:                0,
			ReindexStarted:               time.Now(),
			SearchIndexName:              "ons123456789",
			State:                        "created",
			TotalInsertedSearchDocuments: 0,
			TotalSearchDocuments:         0,
		}

		body, err := json.Marshal(expectedJob)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		searchReindexClient := newSearchReindexClient(t, httpClient)

		Convey("When search-reindexClient.PostJob is called", func() {
			job, err := searchReindexClient.PostJob(ctx, client.Headers{})
			So(err, ShouldBeNil)

			Convey("Then the expected jobid is returned", func() {
				So(job.ID, ShouldEqual, expectedJob.ID)
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("Given a 500 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		searchReindexClient := newSearchReindexClient(t, httpClient)

		Convey("When search-reindexClient.PostJob is called", func() {
			job, err := searchReindexClient.PostJob(ctx, client.Headers{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed as unexpected code from search reindex api: 500")
			So(apiError.ErrorStatus(err), ShouldEqual, http.StatusInternalServerError)

			Convey("Then the expected empty job is returned", func() {
				So(job, ShouldResemble, models.Job{})
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given a 409 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusConflict}, nil)
		searchReindexClient := newSearchReindexClient(t, httpClient)

		Convey("when search-reindexClient.PostJob is called", func() {
			job, err := searchReindexClient.PostJob(ctx, client.Headers{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed as unexpected code from search reindex api: 409")
			So(apiError.ErrorStatus(err), ShouldEqual, http.StatusConflict)

			Convey("then the expected empty job is returned", func() {
				So(job, ShouldResemble, models.Job{})
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func TestClient_PatchJob(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	path := "/jobs/" + testJobID

	patchList := make([]client.PatchOperation, 2)
	statusOperation := client.PatchOperation{
		Op:    "replace",
		Path:  "/state",
		Value: "in-progress",
	}
	patchList[0] = statusOperation
	totalDocsOperation := client.PatchOperation{
		Op:    "replace",
		Path:  "/total_search_documents",
		Value: 100,
	}
	patchList[1] = totalDocsOperation

	invalidPatchList := make([]client.PatchOperation, 1)
	invalidOperation := client.PatchOperation{
		Op:    "invalid operation",
		Path:  "/state",
		Value: "anything will do",
	}
	invalidPatchList[0] = invalidOperation

	headers := client.Headers{
		IfMatch:          "*",
		ServiceAuthToken: serviceToken,
	}

	Convey("Given clienter.Do doesn't return an error", t, func() {
		header := http.Header{}
		header.Add(ETagHeader, testETag)
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: 204,
				Body:       nil,
				Header:     header,
			},
			nil)

		searchReindexClient := newSearchReindexClient(t, httpClient)

		Convey("When search-reindexClient.PatchJob is called", func() {
			respETag, err := searchReindexClient.PatchJob(ctx, headers, testJobID, patchList)
			So(err, ShouldBeNil)

			Convey("Then an ETag is returned", func() {
				So(respETag, ShouldNotBeNil)
				So(respETag, ShouldResemble, testETag)
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				expectedIfMatchHeader := make([]string, 1)
				expectedIfMatchHeader[0] = "*"
				So(doCalls[0].Req.Header[ifMatchHeader], ShouldResemble, expectedIfMatchHeader)
				body, _ := json.Marshal(patchList)
				expectedBody := io.NopCloser(bytes.NewReader(body))
				So(doCalls[0].Req.Body, ShouldResemble, expectedBody)
			})
		})
	})

	Convey("Given a 400 response", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: 400,
				Body:       nil,
				Header:     nil},
			nil)

		searchReindexClient := newSearchReindexClient(t, httpClient)

		Convey("When search-reindexClient.PatchJob is called with an invalid request body", func() {
			respETag, err := searchReindexClient.PatchJob(ctx, headers, testJobID, invalidPatchList)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed as unexpected code from search reindex api: 400")
			So(apiError.ErrorStatus(err), ShouldEqual, 400)

			Convey("Then an empty ETag is returned", func() {
				So(respETag, ShouldNotBeNil)
				So(respETag, ShouldResemble, "")
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				expectedIfMatchHeader := make([]string, 1)
				expectedIfMatchHeader[0] = "*"
				So(doCalls[0].Req.Header[ifMatchHeader], ShouldResemble, expectedIfMatchHeader)
				body, _ := json.Marshal(invalidPatchList)
				expectedBody := io.NopCloser(bytes.NewReader(body))
				So(doCalls[0].Req.Body, ShouldResemble, expectedBody)
			})
		})
	})

	Convey("Given a 500 response", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       nil,
				Header:     nil},
			errors.New("something went wrong in the search reindex api service"))

		searchReindexClient := newSearchReindexClient(t, httpClient)

		Convey("When search-reindexClient.PatchJob is called", func() {
			respETag, err := searchReindexClient.PatchJob(ctx, headers, testJobID, patchList)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed to call search reindex api, error is: something went wrong in the search reindex api service")
			So(apiError.ErrorStatus(err), ShouldEqual, http.StatusInternalServerError)

			Convey("Then an empty ETag is returned", func() {
				So(respETag, ShouldNotBeNil)
				So(respETag, ShouldResemble, "")
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				expectedIfMatchHeader := make([]string, 1)
				expectedIfMatchHeader[0] = "*"
				So(doCalls[0].Req.Header[ifMatchHeader], ShouldResemble, expectedIfMatchHeader)
				body, _ := json.Marshal(patchList)
				expectedBody := io.NopCloser(bytes.NewReader(body))
				So(doCalls[0].Req.Body, ShouldResemble, expectedBody)
			})
		})
	})

	Convey("Given a 409 response", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusConflict,
				Body:       nil,
				Header:     nil},
			nil)

		searchReindexClient := newSearchReindexClient(t, httpClient)

		Convey("When search-reindexClient.PatchJob is called", func() {
			respETag, err := searchReindexClient.PatchJob(ctx, headers, testJobID, patchList)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed as unexpected code from search reindex api: 409")
			So(apiError.ErrorStatus(err), ShouldEqual, http.StatusConflict)

			Convey("Then an empty ETag is returned", func() {
				So(respETag, ShouldNotBeNil)
				So(respETag, ShouldResemble, "")
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				expectedIfMatchHeader := make([]string, 1)
				expectedIfMatchHeader[0] = "*"
				So(doCalls[0].Req.Header[ifMatchHeader], ShouldResemble, expectedIfMatchHeader)
				body, _ := json.Marshal(patchList)
				expectedBody := io.NopCloser(bytes.NewReader(body))
				So(doCalls[0].Req.Body, ShouldResemble, expectedBody)
			})
		})
	})

	Convey("Given a 404 response", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: 404,
				Body:       nil,
				Header:     nil},
			nil)

		searchReindexClient := newSearchReindexClient(t, httpClient)

		Convey("When search-reindexClient.PatchJob is called", func() {
			respETag, err := searchReindexClient.PatchJob(ctx, headers, testJobID, patchList)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed as unexpected code from search reindex api: 404")
			So(apiError.ErrorStatus(err), ShouldEqual, 404)

			Convey("Then an empty ETag is returned", func() {
				So(respETag, ShouldNotBeNil)
				So(respETag, ShouldResemble, "")
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				expectedIfMatchHeader := make([]string, 1)
				expectedIfMatchHeader[0] = "*"
				So(doCalls[0].Req.Header[ifMatchHeader], ShouldResemble, expectedIfMatchHeader)
				body, _ := json.Marshal(patchList)
				expectedBody := io.NopCloser(bytes.NewReader(body))
				So(doCalls[0].Req.Body, ShouldResemble, expectedBody)
			})
		})
	})
}
