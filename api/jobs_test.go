package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	dpresponse "github.com/ONSdigital/dp-net/v2/handlers/response"
	dpHTTP "github.com/ONSdigital/dp-net/v2/http"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/dp-search-reindex-api/api"
	apiMock "github.com/ONSdigital/dp-search-reindex-api/api/mock"
	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

// Constants for testing
const (
	eTagValidJobID1         = `"e3b461ea19d5a2e345db1f49b7beb076a9c751d3"`
	validJobID1             = "UUID1"
	validJobID2             = "UUID2"
	validJobID3             = "UUID3"
	notFoundJobID           = "UUID4"
	unLockableJobID         = "UUID5"
	expectedServerErrorMsg  = "internal server error"
	validCount              = "3"
	countNotANumber         = "notANumber"
	countNegativeInt        = "-3"
	expectedOffsetErrorMsg  = "invalid offset query parameter"
	expectedLimitErrorMsg   = "invalid limit query parameter"
	expectedLimitOverMaxMsg = "limit query parameter is larger than the maximum allowed"
)

var (
	zeroTime      = time.Time{}.UTC()
	errUnexpected = errors.New("an unexpected error occurred")
)

// expectedJob returns a Job resource that can be used to define and test expected values within it
// If jsonResponse is set to true, this updates the links in the resource to contain the host address and version number
func expectedJob(ctx context.Context, t *testing.T, cfg *config.Config, jsonResponse bool, id, searchIndexName string) models.Job {
	job := models.Job{
		ID:          id,
		LastUpdated: zeroTime,
		Links: &models.JobLinks{
			Tasks: fmt.Sprintf("/jobs/%s/tasks", id),
			Self:  fmt.Sprintf("/jobs/%s", id),
		},
		NumberOfTasks:                0,
		ReindexCompleted:             zeroTime,
		ReindexFailed:                zeroTime,
		ReindexStarted:               zeroTime,
		SearchIndexName:              searchIndexName,
		State:                        models.JobStateCreated,
		TotalSearchDocuments:         0,
		TotalInsertedSearchDocuments: 0,
	}

	jobETag, err := models.GenerateETagForJob(ctx, job)
	if err != nil {
		t.Errorf("failed to generate eTag for expected test job - error: %v", err)
	}
	job.ETag = jobETag

	if jsonResponse {
		job.Links.Tasks = fmt.Sprintf("%s/%s%s", cfg.BindAddr, cfg.LatestVersion, job.Links.Tasks)
		job.Links.Self = fmt.Sprintf("%s/%s%s", cfg.BindAddr, cfg.LatestVersion, job.Links.Self)
	}

	return job
}

func TestCreateJobHandler(t *testing.T) {
	t.Parallel()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	dataStorerMock := &apiMock.DataStorerMock{
		CheckInProgressJobFunc: func(ctx context.Context) error {
			return nil
		},
		ValidateJobIDUniqueFunc: func(ctx context.Context, id string) error {
			return nil
		},
		CreateJobFunc: func(ctx context.Context, job models.Job) error {
			return nil
		},
	}

	indexerMock := &apiMock.IndexerMock{
		CreateIndexFunc: func(ctx context.Context, serviceAuthToken, searchAPISearchURL string, httpClient dpHTTP.Clienter) (*http.Response, error) {
			resp := &http.Response{
				StatusCode: 201,
			}
			return resp, nil
		},
		GetIndexNameFromResponseFunc: func(ctx context.Context, body io.ReadCloser) (string, error) {
			return "ons1638363874110115", nil
		},
		SendReindexRequestedEventFunc: func(cfg *config.Config, jobID string, indexName string) error {
			return nil
		},
	}

	producerMock := &apiMock.ReindexRequestedProducerMock{
		ProduceReindexRequestedFunc: func(ctx context.Context, event models.ReindexRequested) error {
			return nil
		},
	}

	models.NewJobID = func() string {
		return validJobID1
	}

	Convey("Given the Search Reindex Job API can create valid search reindex jobs and store their details in a Data Store", t, func() {
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, dpHTTP.NewClient(), indexerMock, producerMock)

		Convey("When a new reindex job is created and stored", func() {
			req := httptest.NewRequest("POST", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.CreateJobHandler(resp, req)

			Convey("Then the newly created search reindex job is returned with status code 201", func() {
				So(resp.Code, ShouldEqual, http.StatusCreated)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				newJob := models.Job{}
				err = json.Unmarshal(payload, &newJob)
				So(err, ShouldBeNil)

				expectedJob := expectedJob(context.Background(), t, cfg, true, validJobID1, "ons1638363874110115")

				Convey("And the new job resource should contain expected default values", func() {
					So(newJob.ID, ShouldEqual, expectedJob.ID)
					So(newJob.Links, ShouldResemble, expectedJob.Links)
					So(newJob.NumberOfTasks, ShouldEqual, expectedJob.NumberOfTasks)
					So(newJob.ReindexCompleted, ShouldEqual, expectedJob.ReindexCompleted)
					So(newJob.ReindexFailed, ShouldEqual, expectedJob.ReindexFailed)
					So(newJob.ReindexStarted, ShouldEqual, expectedJob.ReindexStarted)
					So(newJob.SearchIndexName, ShouldEqual, expectedJob.SearchIndexName)
					So(newJob.State, ShouldEqual, expectedJob.State)
					So(newJob.TotalSearchDocuments, ShouldEqual, expectedJob.TotalSearchDocuments)
					So(newJob.TotalInsertedSearchDocuments, ShouldEqual, expectedJob.TotalInsertedSearchDocuments)
				})
			})
		})
	})

	Convey("Given an existing reindex job is in progress", t, func() {
		jobInProgressDataStorerMock := &apiMock.DataStorerMock{
			CheckInProgressJobFunc: func(ctx context.Context) error {
				return mongo.ErrExistingJobInProgress
			},
		}

		apiInstance := api.Setup(mux.NewRouter(), jobInProgressDataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, dpHTTP.NewClient(), indexerMock, producerMock)

		Convey("When the POST /jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.CreateJobHandler(resp, req)

			Convey("Then the response should return a 409 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusConflict)

				Convey("And an error message should be returned in the response body", func() {
					errMsg := strings.TrimSpace(resp.Body.String())
					So(errMsg, ShouldEqual, apierrors.ErrExistingJobInProgress.Error())
				})
			})
		})
	})

	Convey("Given an error with the datastore", t, func() {
		dataStorerFailMock := &apiMock.DataStorerMock{
			CheckInProgressJobFunc: func(ctx context.Context) error {
				return errUnexpected
			},
		}

		apiInstance := api.Setup(mux.NewRouter(), dataStorerFailMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, dpHTTP.NewClient(), indexerMock, producerMock)

		Convey("When the POST /jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.CreateJobHandler(resp, req)

			Convey("Then the response should return a 500 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)

				Convey("And an error message should be returned in the response body", func() {
					errMsg := strings.TrimSpace(resp.Body.String())
					So(errMsg, ShouldEqual, apierrors.ErrInternalServer.Error())
				})
			})
		})
	})

	Convey("Given an error with Search API to create index", t, func() {
		createIndexErrMock := &apiMock.IndexerMock{
			CreateIndexFunc: func(ctx context.Context, serviceAuthToken, searchAPISearchURL string, httpClient dpHTTP.Clienter) (*http.Response, error) {
				return nil, errUnexpected
			},
		}

		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, dpHTTP.NewClient(), createIndexErrMock, producerMock)

		Convey("When the POST /jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.CreateJobHandler(resp, req)

			Convey("Then the response should return a 500 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)

				Convey("And an error message should be returned in the response body", func() {
					errMsg := strings.TrimSpace(resp.Body.String())
					So(errMsg, ShouldEqual, apierrors.ErrInternalServer.Error())
				})
			})
		})
	})

	Convey("Given unsuccessful status code returned from Search API to create index", t, func() {
		createIndexFailStatusMock := &apiMock.IndexerMock{
			CreateIndexFunc: func(ctx context.Context, serviceAuthToken, searchAPISearchURL string, httpClient dpHTTP.Clienter) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: 500,
				}
				return resp, nil
			},
		}

		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, dpHTTP.NewClient(), createIndexFailStatusMock, producerMock)

		Convey("When the POST /jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.CreateJobHandler(resp, req)

			Convey("Then the response should return a 500 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)

				Convey("And an error message should be returned in the response body", func() {
					errMsg := strings.TrimSpace(resp.Body.String())
					So(errMsg, ShouldEqual, apierrors.ErrInternalServer.Error())
				})
			})
		})
	})

	Convey("Given an error with Search API to get index name", t, func() {
		GetIndexNameFromResponseErrMock := &apiMock.IndexerMock{
			CreateIndexFunc: func(ctx context.Context, serviceAuthToken, searchAPISearchURL string, httpClient dpHTTP.Clienter) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: 201,
				}
				return resp, nil
			},
			GetIndexNameFromResponseFunc: func(ctx context.Context, body io.ReadCloser) (string, error) {
				return "", errUnexpected
			},
		}

		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, dpHTTP.NewClient(), GetIndexNameFromResponseErrMock, producerMock)

		Convey("When the POST /jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.CreateJobHandler(resp, req)

			Convey("Then the response should return a 500 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)

				Convey("And an error message should be returned in the response body", func() {
					errMsg := strings.TrimSpace(resp.Body.String())
					So(errMsg, ShouldEqual, apierrors.ErrInternalServer.Error())
				})
			})
		})
	})

	Convey("Given the job ID is not unique for the new job", t, func() {
		createJobDataStorerFailMock := &apiMock.DataStorerMock{
			CheckInProgressJobFunc: func(ctx context.Context) error {
				return nil
			},
			ValidateJobIDUniqueFunc: func(ctx context.Context, id string) error {
				return mongo.ErrDuplicateIDProvided
			},
		}

		apiInstance := api.Setup(mux.NewRouter(), createJobDataStorerFailMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, dpHTTP.NewClient(), indexerMock, producerMock)

		Convey("When the POST /jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.CreateJobHandler(resp, req)

			Convey("Then the response should return a 500 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)

				Convey("And an error message should be returned in the response body", func() {
					errMsg := strings.TrimSpace(resp.Body.String())
					So(errMsg, ShouldEqual, apierrors.ErrInternalServer.Error())
				})
			})
		})
	})

	Convey("Given an error to create a reindex job in the datastore", t, func() {
		createJobDataStorerFailMock := &apiMock.DataStorerMock{
			CheckInProgressJobFunc: func(ctx context.Context) error {
				return nil
			},
			ValidateJobIDUniqueFunc: func(ctx context.Context, id string) error {
				return nil
			},
			CreateJobFunc: func(ctx context.Context, job models.Job) error {
				return errUnexpected
			},
		}

		apiInstance := api.Setup(mux.NewRouter(), createJobDataStorerFailMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, dpHTTP.NewClient(), indexerMock, producerMock)

		Convey("When the POST /jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.CreateJobHandler(resp, req)

			Convey("Then the response should return a 500 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)

				Convey("And an error message should be returned in the response body", func() {
					errMsg := strings.TrimSpace(resp.Body.String())
					So(errMsg, ShouldEqual, apierrors.ErrInternalServer.Error())
				})
			})
		})
	})

	Convey("Given an error to send a reindex-requested event", t, func() {
		producerFailMock := &apiMock.ReindexRequestedProducerMock{
			ProduceReindexRequestedFunc: func(ctx context.Context, event models.ReindexRequested) error {
				return errUnexpected
			},
		}

		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, dpHTTP.NewClient(), indexerMock, producerFailMock)

		Convey("When the POST /jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.CreateJobHandler(resp, req)

			Convey("Then the response should return a 500 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)

				Convey("And an error message should be returned in the response body", func() {
					errMsg := strings.TrimSpace(resp.Body.String())
					So(errMsg, ShouldEqual, apierrors.ErrInternalServer.Error())
				})
			})
		})
	})
}

func TestGetJobHandler(t *testing.T) {
	t.Parallel()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	Convey("Given a Search Reindex Job API that returns specific jobs using their id as a key", t, func() {
		dataStorerMock := &apiMock.DataStorerMock{
			GetJobFunc: func(ctx context.Context, id string) (models.Job, error) {
				switch id {
				case validJobID2:
					return expectedJob(ctx, t, cfg, false, id, ""), nil
				case notFoundJobID:
					return models.Job{}, mongo.ErrJobNotFound
				default:
					return models.Job{}, errUnexpected
				}
			},
			AcquireJobLockFunc: func(ctx context.Context, id string) (string, error) {
				switch id {
				case unLockableJobID:
					return "", errors.New("acquiring lock failed")
				default:
					return "", nil
				}
			},
			UnlockJobFunc: func(ctx context.Context, lockID string) {
				// mock UnlockJob to be successful by doing nothing
			},
		}

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to get a specific job that exists in the Data Store", func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost:25700/jobs/%s", validJobID2), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the relevant search reindex job is returned with status code 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				jobReturned := &models.Job{}
				err = json.Unmarshal(payload, jobReturned)
				So(err, ShouldBeNil)

				expectedJob := expectedJob(context.Background(), t, cfg, true, validJobID2, "")

				Convey("And the returned job resource should contain expected values", func() {
					So(jobReturned.ID, ShouldEqual, expectedJob.ID)
					So(jobReturned.Links, ShouldResemble, expectedJob.Links)
					So(jobReturned.NumberOfTasks, ShouldEqual, expectedJob.NumberOfTasks)
					So(jobReturned.ReindexCompleted, ShouldEqual, expectedJob.ReindexCompleted)
					So(jobReturned.ReindexFailed, ShouldEqual, expectedJob.ReindexFailed)
					So(jobReturned.ReindexStarted, ShouldEqual, expectedJob.ReindexStarted)
					So(jobReturned.SearchIndexName, ShouldEqual, expectedJob.SearchIndexName)
					So(jobReturned.State, ShouldEqual, expectedJob.State)
					So(jobReturned.TotalSearchDocuments, ShouldEqual, expectedJob.TotalSearchDocuments)
					So(jobReturned.TotalInsertedSearchDocuments, ShouldEqual, expectedJob.TotalInsertedSearchDocuments)
				})
			})
		})

		Convey("When a request is made to get a specific job that does not exist in the Data Store", func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost:25700/jobs/%s", notFoundJobID), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then job resource was not found returning a status code of 404", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "failed to find the specified reindex job")
			})
		})

		Convey("When a request is made to get a specific job but the Data Store is unable to lock the id", func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost:25700/jobs/%s", unLockableJobID), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an error with status code 500 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedServerErrorMsg)
			})
		})

		Convey("When a request is made to get a specific job but an unexpected error occurs in the Data Store", func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost:25700/jobs/%s", validJobID3), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an error with status code 500 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedServerErrorMsg)
			})
		})
	})
}

func TestGetJobsHandler(t *testing.T) {
	t.Parallel()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	Convey("Given a Search Reindex Job API that returns a list of jobs", t, func() {
		dataStorerMock := &apiMock.DataStorerMock{
			GetJobsFunc: func(ctx context.Context, options mongo.Options) (models.Jobs, error) {
				jobs := models.Jobs{}
				jobsList := make([]models.Job, 2)
				offsetJobsList := make([]models.Job, 1)

				firstJob := expectedJob(ctx, t, cfg, false, validJobID1, "")
				jobsList[0] = firstJob

				secondJob := expectedJob(ctx, t, cfg, false, validJobID2, "")
				jobsList[1] = secondJob
				offsetJobsList[0] = secondJob

				switch {
				case (options.Offset == 0) && (options.Limit > 1):
					jobs.JobList = jobsList
				case (options.Offset == 1) && (options.Limit > 0):
					jobs.JobList = offsetJobsList
				default:
					jobs.JobList = make([]models.Job, 0)
				}

				return jobs, err
			},
		}

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to get a list of all the jobs that exist in the Data Store", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a list of jobs is returned with status code 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				jobsReturned := models.Jobs{}
				err = json.Unmarshal(payload, &jobsReturned)
				So(err, ShouldBeNil)

				ctx := context.Background()

				expectedJob1 := expectedJob(ctx, t, cfg, true, validJobID1, "")
				expectedJob2 := expectedJob(ctx, t, cfg, true, validJobID2, "")

				Convey("And the returned list should contain expected jobs", func() {
					returnedJobList := jobsReturned.JobList
					So(returnedJobList, ShouldHaveLength, 2)
					returnedJob1 := returnedJobList[0]
					So(returnedJob1.ID, ShouldEqual, expectedJob1.ID)
					So(returnedJob1.Links, ShouldResemble, expectedJob1.Links)
					So(returnedJob1.NumberOfTasks, ShouldEqual, expectedJob1.NumberOfTasks)
					So(returnedJob1.ReindexCompleted, ShouldEqual, expectedJob1.ReindexCompleted)
					So(returnedJob1.ReindexFailed, ShouldEqual, expectedJob1.ReindexFailed)
					So(returnedJob1.ReindexStarted, ShouldEqual, expectedJob1.ReindexStarted)
					So(returnedJob1.SearchIndexName, ShouldEqual, expectedJob1.SearchIndexName)
					So(returnedJob1.State, ShouldEqual, expectedJob1.State)
					So(returnedJob1.TotalSearchDocuments, ShouldEqual, expectedJob1.TotalSearchDocuments)
					So(returnedJob1.TotalInsertedSearchDocuments, ShouldEqual, expectedJob1.TotalInsertedSearchDocuments)
					returnedJob2 := returnedJobList[1]
					So(returnedJob2.ID, ShouldEqual, expectedJob2.ID)
					So(returnedJob2.Links, ShouldResemble, expectedJob2.Links)
					So(returnedJob2.NumberOfTasks, ShouldEqual, expectedJob2.NumberOfTasks)
					So(returnedJob2.ReindexCompleted, ShouldEqual, expectedJob2.ReindexCompleted)
					So(returnedJob2.ReindexFailed, ShouldEqual, expectedJob2.ReindexFailed)
					So(returnedJob2.ReindexStarted, ShouldEqual, expectedJob2.ReindexStarted)
					So(returnedJob2.SearchIndexName, ShouldEqual, expectedJob2.SearchIndexName)
					So(returnedJob2.State, ShouldEqual, expectedJob2.State)
					So(returnedJob2.TotalSearchDocuments, ShouldEqual, expectedJob2.TotalSearchDocuments)
					So(returnedJob2.TotalInsertedSearchDocuments, ShouldEqual, expectedJob2.TotalInsertedSearchDocuments)
				})
			})
		})

		Convey("When a request is made to get a list of jobs with an offset of 1 and a limit of 20", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs?offset=1&limit=20", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a list of jobs is returned with status code 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				jobsReturned := models.Jobs{}
				err = json.Unmarshal(payload, &jobsReturned)
				So(err, ShouldBeNil)

				expectedJob := expectedJob(context.Background(), t, cfg, true, validJobID2, "")

				Convey("And the returned list should contain the expected job", func() {
					returnedJobList := jobsReturned.JobList
					So(returnedJobList, ShouldHaveLength, 1)
					returnedJob := returnedJobList[0]
					So(returnedJob.ID, ShouldEqual, expectedJob.ID)
					So(returnedJob.Links, ShouldResemble, expectedJob.Links)
					So(returnedJob.NumberOfTasks, ShouldEqual, expectedJob.NumberOfTasks)
					So(returnedJob.ReindexCompleted, ShouldEqual, expectedJob.ReindexCompleted)
					So(returnedJob.ReindexFailed, ShouldEqual, expectedJob.ReindexFailed)
					So(returnedJob.ReindexStarted, ShouldEqual, expectedJob.ReindexStarted)
					So(returnedJob.SearchIndexName, ShouldEqual, expectedJob.SearchIndexName)
					So(returnedJob.State, ShouldEqual, expectedJob.State)
					So(returnedJob.TotalSearchDocuments, ShouldEqual, expectedJob.TotalSearchDocuments)
					So(returnedJob.TotalInsertedSearchDocuments, ShouldEqual, expectedJob.TotalInsertedSearchDocuments)
				})
			})
		})

		Convey("When a request is made to get a list of jobs with an offset greater than the total number of jobs in the Data Store", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs?offset=10&limit=20", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a list of jobs is returned with status code 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				jobsReturned := models.Jobs{}
				err = json.Unmarshal(payload, &jobsReturned)
				So(err, ShouldBeNil)

				Convey("And the returned list should be empty", func() {
					returnedJobList := jobsReturned.JobList
					So(returnedJobList, ShouldHaveLength, 0)
				})
			})
		})

		Convey("When a request is made to get a list of jobs with an offset that is not numeric", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs?offset=hi&limit=20", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a bad request error is returned with status code 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedOffsetErrorMsg)
			})
		})

		Convey("When a request is made to get a list of jobs with an offset that is negative", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs?offset=-3&limit=20", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a bad request error is returned with status code 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedOffsetErrorMsg)
			})
		})

		Convey("When a request is made to get a list of jobs with a limit that is not numeric", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs?offset=0&limit=sky", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a bad request error is returned with status code 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedLimitErrorMsg)
			})
		})

		Convey("When a request is made to get a list of jobs with a limit that is negative", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs?offset=0&limit=-1", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a bad request error is returned with status code 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedLimitErrorMsg)
			})
		})

		Convey("When a request is made to get a list of jobs with a limit that is greater than the maximum allowed", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs?offset=0&limit=1001", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a bad request error is returned with status code 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedLimitOverMaxMsg)
			})
		})
	})
}

func TestGetJobsHandlerWithEmptyJobStore(t *testing.T) {
	t.Parallel()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	Convey("Given a Search Reindex Job API that returns an empty list of jobs", t, func() {
		dataStorerMock := &apiMock.DataStorerMock{
			GetJobsFunc: func(ctx context.Context, options mongo.Options) (models.Jobs, error) {
				jobs := models.Jobs{}

				return jobs, nil
			},
		}

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to get a list of all the jobs that exist in the jobs collection", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a jobs resource is returned with status code 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				jobsReturned := models.Jobs{}
				err = json.Unmarshal(payload, &jobsReturned)
				So(err, ShouldBeNil)

				Convey("And the returned jobs list should be empty", func() {
					So(jobsReturned.JobList, ShouldHaveLength, 0)
				})
			})
		})
	})
}

func TestGetJobsHandlerWithInternalServerError(t *testing.T) {
	t.Parallel()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	Convey("Given a Search Reindex Job API that that failed to connect to the Data Store", t, func() {
		dataStorerMock := &apiMock.DataStorerMock{
			GetJobsFunc: func(ctx context.Context, options mongo.Options) (models.Jobs, error) {
				jobs := models.Jobs{}

				return jobs, errors.New("something went wrong in the server")
			},
		}

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to get a list of all the jobs that exist in the jobs collection", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an error with status code 500 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedServerErrorMsg)
			})
		})
	})
}

func TestPutNumTasksHandler(t *testing.T) {
	t.Parallel()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	Convey("Given a Search Reindex Job API that updates the number of tasks for specific jobs using their id as a key", t, func() {
		jobStoreMock := &apiMock.DataStorerMock{
			PutNumberOfTasksFunc: func(ctx context.Context, id string, count int) error {
				switch id {
				case validJobID2:
					return nil
				case validJobID3:
					return errors.New("unexpected error updating the number of tasks")
				default:
					return mongo.ErrJobNotFound
				}
			},
			AcquireJobLockFunc: func(ctx context.Context, id string) (string, error) {
				switch id {
				case unLockableJobID:
					return "", errors.New("acquiring lock failed")
				default:
					return "", nil
				}
			},
			UnlockJobFunc: func(ctx context.Context, lockID string) {
				// mock UnlockJob to be successful by doing nothing
			},
		}

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), jobStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of tasks of a specific job that exists in the Data Store", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/jobs/%s/number_of_tasks/%s", validJobID2, validCount), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a status code 200 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When a request is made to update the number of tasks of a specific job that does not exist in the Data Store", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/jobs/%s/number_of_tasks/%s", notFoundJobID, validCount), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then job resource was not found returning a status code of 404", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "failed to find the specified reindex job")
			})
		})

		Convey("When a request is made to update the number of tasks of a specific job and an unexpected error occurs", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/jobs/%s/number_of_tasks/%s", validJobID3, validCount), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response returns a status code of 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "internal server error")
			})
		})

		Convey("When a request is made to update the number of tasks but the path parameter given as the Count is not an integer", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/jobs/%s/number_of_tasks/%s", validJobID2, countNotANumber), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then it is a bad request returning a status code of 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "number of tasks must be a positive integer")
			})
		})

		Convey("When a request is made to update the number of tasks but the path parameter given as the Count is a negative integer", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/jobs/%s/number_of_tasks/%s", validJobID2, countNegativeInt), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then it is a bad request returning a status code of 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "number of tasks must be a positive integer")
			})
		})

		Convey("When a request is made to update the number of tasks but the Data Store is unable to lock the id", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/jobs/%s/number_of_tasks/%s", unLockableJobID, validCount), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an error with status code 500 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedServerErrorMsg)
			})
		})
	})
}

func TestPatchJobStatusHandler(t *testing.T) {
	t.Parallel()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	var etag1, etag2 string

	jobStoreMock := &apiMock.DataStorerMock{
		GetJobFunc: func(ctx context.Context, id string) (models.Job, error) {
			switch id {
			case validJobID1:
				newJob := expectedJob(ctx, t, cfg, false, validJobID1, "")
				etag1 = newJob.ETag
				return newJob, nil
			case validJobID2:
				newJob := expectedJob(ctx, t, cfg, false, validJobID2, "")
				etag2 = newJob.ETag
				return newJob, nil
			case unLockableJobID:
				return expectedJob(ctx, t, cfg, false, unLockableJobID, ""), nil
			case notFoundJobID:
				return models.Job{}, mongo.ErrJobNotFound
			default:
				return models.Job{}, errUnexpected
			}
		},
		AcquireJobLockFunc: func(ctx context.Context, id string) (string, error) {
			switch id {
			case unLockableJobID:
				return "", errors.New("acquiring lock failed")
			default:
				return "", nil
			}
		},
		UnlockJobFunc: func(ctx context.Context, lockID string) {
			// mock UnlockJob to be successful by doing nothing
		},
		UpdateJobFunc: func(ctx context.Context, id string, updates bson.M) error {
			switch id {
			case validJobID2:
				return errUnexpected
			default:
				return nil
			}
		},
	}

	validPatchBody := `[
		{ "op": "replace", "path": "/state", "value": "created" },
		{ "op": "replace", "path": "/total_search_documents", "value": 100 }
	]`

	Convey("Given a Search Reindex Job API that updates state of a job via patch request", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), jobStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a patch request is made with valid job ID and valid patch operations", func() {
			req := httptest.NewRequest("PATCH", fmt.Sprintf("http://localhost:25700/jobs/%s", validJobID1), bytes.NewBufferString(validPatchBody))
			headers.SetIfMatch(req, etag1)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response should return a 204 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusNoContent)

				Convey("And the new eTag of the resource is returned via ETag header", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
				})
			})
		})

		Convey("When a patch request is made with invalid job ID", func() {
			req := httptest.NewRequest("PATCH", fmt.Sprintf("http://localhost:25700/jobs/%s", notFoundJobID), bytes.NewBufferString(validPatchBody))
			headers.SetIfMatch(req, etag1)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response should return a 404 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldNotBeEmpty)
			})
		})

		Convey("When connection to datastore has failed", func() {
			req := httptest.NewRequest("PATCH", fmt.Sprintf("http://localhost:25700/jobs/%s", "invalid"), bytes.NewBufferString(validPatchBody))
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response should return a 500 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldNotBeEmpty)
			})
		})

		Convey("When a patch request is made with no If-Match header", func() {
			req := httptest.NewRequest("PATCH", fmt.Sprintf("http://localhost:25700/jobs/%s", validJobID1), bytes.NewBufferString(validPatchBody))
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response should return a 201 status code as eTag check is ignored", func() {
				So(resp.Code, ShouldEqual, http.StatusNoContent)
			})
		})

		Convey("When a patch request is made with outdated or unknown eTag in If-Match header", func() {
			req := httptest.NewRequest("PATCH", fmt.Sprintf("http://localhost:25700/jobs/%s", validJobID1), bytes.NewBufferString(validPatchBody))
			headers.SetIfMatch(req, "invalidETag")
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response should return a 201 status code as eTag check is ignored", func() {
				So(resp.Code, ShouldEqual, http.StatusConflict)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldNotBeEmpty)
			})
		})

		Convey("When a patch request is made with invalid patch body given", func() {
			req := httptest.NewRequest("PATCH", fmt.Sprintf("http://localhost:25700/jobs/%s", validJobID1), bytes.NewBufferString("{}"))
			headers.SetIfMatch(req, etag1)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response should return a 400 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldNotBeEmpty)
			})
		})

		Convey("When a patch request is made with no patches given in request body", func() {
			req := httptest.NewRequest("PATCH", fmt.Sprintf("http://localhost:25700/jobs/%s", validJobID1), bytes.NewBufferString("[]"))
			headers.SetIfMatch(req, eTagValidJobID1)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response should return a 400 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldNotBeEmpty)
			})
		})

		Convey("When a patch request is made with patch body containing invalid information", func() {
			patchBodyWithInvalidData := `[
				{ "op": "replace", "path": "/total_search_documents", "value": "invalid" }
			]`

			req := httptest.NewRequest("PATCH", fmt.Sprintf("http://localhost:25700/jobs/%s", validJobID1), bytes.NewBufferString(patchBodyWithInvalidData))
			headers.SetIfMatch(req, etag1)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response should return a 400 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldNotBeEmpty)
			})
		})

		Convey("When acquiring job lock to update job has failed", func() {
			req := httptest.NewRequest("PATCH", fmt.Sprintf("http://localhost:25700/jobs/%s", unLockableJobID), bytes.NewBufferString(validPatchBody))
			unLockableJobETag := `"24decf55038de874bc6fa9cf0930adc219f15db1"`
			headers.SetIfMatch(req, unLockableJobETag)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response should return a 500 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldNotBeEmpty)
			})
		})

		Convey("When a patch request is made which results in no modification", func() {
			patchBodyWithNoModification := `[
				{ "op": "replace", "path": "/state", "value": "created" }
			]`

			req := httptest.NewRequest("PATCH", fmt.Sprintf("http://localhost:25700/jobs/%s", validJobID1), bytes.NewBufferString(patchBodyWithNoModification))
			headers.SetIfMatch(req, etag1)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response should return a 304 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusNotModified)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldNotBeEmpty)
			})
		})

		Convey("When the update to job with patches has failed due to failing on UpdateJobWithPatches func", func() {
			req := httptest.NewRequest("PATCH", fmt.Sprintf("http://localhost:25700/jobs/%s", validJobID2), bytes.NewBufferString(validPatchBody))
			headers.SetIfMatch(req, etag2)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response should return a 500 status code", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldNotBeEmpty)
			})
		})
	})
}

func TestPreparePatchUpdatesSuccess(t *testing.T) {
	t.Parallel()

	testCtx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	currentJob := expectedJob(testCtx, t, cfg, false, validJobID1, "")

	Convey("Given valid patches", t, func() {
		validPatches := []dprequest.Patch{
			{
				Op:    dprequest.OpReplace.String(),
				Path:  models.JobTotalSearchDocumentsPath,
				Value: float64(100),
			},
			{
				Op:    dprequest.OpReplace.String(),
				Path:  models.JobNoOfTasksPath,
				Value: float64(2),
			},
		}

		Convey("When preparePatchUpdates is called", func() {
			updatedJob, bsonUpdates, err := api.GetUpdatesFromJobPatches(testCtx, validPatches, currentJob)

			Convey("Then updatedJob should contain updates from the patch", func() {
				So(updatedJob.TotalSearchDocuments, ShouldEqual, 100)
				So(updatedJob.NumberOfTasks, ShouldEqual, 2)

				Convey("And bsonUpdates should contain updates from the patch", func() {
					So(bsonUpdates[models.JobTotalSearchDocumentsKey], ShouldEqual, 100)
					So(bsonUpdates[models.JobNoOfTasksKey], ShouldEqual, 2)

					Convey("And LastUpdated should be updated", func() {
						So(updatedJob.LastUpdated, ShouldNotEqual, currentJob.LastUpdated)
						So(bsonUpdates[models.JobLastUpdatedKey], ShouldNotBeEmpty)

						Convey("And no error should be returned", func() {
							So(err, ShouldBeNil)
						})
					})
				})
			})
		})
	})

	Convey("Given patches which changes state to in-progress", t, func() {
		inProgressStatePatches := []dprequest.Patch{
			{
				Op:    dprequest.OpReplace.String(),
				Path:  models.JobStatePath,
				Value: models.JobStateInProgress,
			},
		}

		Convey("When preparePatchUpdates is called", func() {
			updatedJob, bsonUpdates, err := api.GetUpdatesFromJobPatches(testCtx, inProgressStatePatches, currentJob)

			Convey("Then updatedJob and bsonUpdates should contain updates from the patch", func() {
				So(updatedJob.State, ShouldEqual, models.JobStateInProgress)
				So(bsonUpdates[models.JobStateKey], ShouldEqual, models.JobStateInProgress)

				Convey("And reindex started should be updated", func() {
					So(updatedJob.ReindexStarted, ShouldNotEqual, currentJob.ReindexStarted)
					So(bsonUpdates[models.JobReindexStartedKey], ShouldNotBeEmpty)

					Convey("And LastUpdated should be updated", func() {
						So(updatedJob.LastUpdated, ShouldNotEqual, currentJob.LastUpdated)
						So(bsonUpdates[models.JobLastUpdatedKey], ShouldNotBeEmpty)

						Convey("And no error should be returned", func() {
							So(err, ShouldBeNil)
						})
					})
				})
			})
		})
	})

	Convey("Given patches which changes state to failed", t, func() {
		failedStatePatches := []dprequest.Patch{
			{
				Op:    dprequest.OpReplace.String(),
				Path:  models.JobStatePath,
				Value: models.JobStateFailed,
			},
		}

		Convey("When preparePatchUpdates is called", func() {
			updatedJob, bsonUpdates, err := api.GetUpdatesFromJobPatches(testCtx, failedStatePatches, currentJob)

			Convey("Then updatedJob and bsonUpdates should contain updates from the patch", func() {
				So(updatedJob.State, ShouldEqual, models.JobStateFailed)
				So(bsonUpdates[models.JobStateKey], ShouldEqual, models.JobStateFailed)

				Convey("And ReindexFailed should be updated", func() {
					So(updatedJob.ReindexFailed, ShouldNotEqual, currentJob.ReindexFailed)
					So(bsonUpdates[models.JobReindexFailedKey], ShouldNotBeEmpty)

					Convey("And LastUpdated should be updated", func() {
						So(updatedJob.LastUpdated, ShouldNotEqual, currentJob.LastUpdated)
						So(bsonUpdates[models.JobLastUpdatedKey], ShouldNotBeEmpty)

						Convey("And no error should be returned", func() {
							So(err, ShouldBeNil)
						})
					})
				})
			})
		})
	})

	Convey("Given patches which changes state to completed", t, func() {
		completedStatePatches := []dprequest.Patch{
			{
				Op:    dprequest.OpReplace.String(),
				Path:  models.JobStatePath,
				Value: models.JobStateCompleted,
			},
		}

		Convey("When preparePatchUpdates is called", func() {
			updatedJob, bsonUpdates, err := api.GetUpdatesFromJobPatches(testCtx, completedStatePatches, currentJob)

			Convey("Then updatedJob and bsonUpdates should contain updates from the patch", func() {
				So(updatedJob.State, ShouldEqual, models.JobStateCompleted)
				So(bsonUpdates[models.JobStateKey], ShouldEqual, models.JobStateCompleted)

				Convey("And ReindexCompleted should be updated", func() {
					So(updatedJob.ReindexCompleted, ShouldNotEqual, currentJob.ReindexCompleted)
					So(bsonUpdates[models.JobReindexCompletedKey], ShouldNotBeEmpty)

					Convey("And LastUpdated should be updated", func() {
						So(updatedJob.LastUpdated, ShouldNotEqual, currentJob.LastUpdated)
						So(bsonUpdates[models.JobLastUpdatedKey], ShouldNotBeEmpty)

						Convey("And no error should be returned", func() {
							So(err, ShouldBeNil)
						})
					})
				})
			})
		})
	})
}

func TestPreparePatchUpdatesFail(t *testing.T) {
	t.Parallel()

	testCtx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	currentJob := expectedJob(testCtx, t, cfg, false, validJobID1, "")

	Convey("Given patches with unknown path", t, func() {
		unknownPathPatches := []dprequest.Patch{
			{
				Op:    dprequest.OpReplace.String(),
				Path:  "/unknown",
				Value: "unknown",
			},
		}

		Convey("When preparePatchUpdates is called", func() {
			updatedJob, bsonUpdates, err := api.GetUpdatesFromJobPatches(testCtx, unknownPathPatches, currentJob)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, fmt.Sprintf("provided path '%s' not supported", unknownPathPatches[0].Path))

				So(updatedJob, ShouldResemble, models.Job{})
				So(bsonUpdates, ShouldBeEmpty)
			})
		})
	})

	Convey("Given patches with invalid number of tasks", t, func() {
		invalidNoOfTasksPatches := []dprequest.Patch{
			{
				Op:    dprequest.OpReplace.String(),
				Path:  models.JobNoOfTasksPath,
				Value: "unknown",
			},
		}

		Convey("When preparePatchUpdates is called", func() {
			updatedJob, bsonUpdates, err := api.GetUpdatesFromJobPatches(testCtx, invalidNoOfTasksPatches, currentJob)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, fmt.Sprintf("wrong value type `%s` for `%s`, expected an integer", api.GetValueType(invalidNoOfTasksPatches[0].Value), invalidNoOfTasksPatches[0].Path))

				So(updatedJob, ShouldResemble, models.Job{})
				So(bsonUpdates, ShouldBeEmpty)
			})
		})
	})

	Convey("Given patches with unknown state", t, func() {
		unknownStatePatches := []dprequest.Patch{
			{
				Op:    dprequest.OpReplace.String(),
				Path:  models.JobStatePath,
				Value: "unknown",
			},
		}

		Convey("When preparePatchUpdates is called", func() {
			updatedJob, bsonUpdates, err := api.GetUpdatesFromJobPatches(testCtx, unknownStatePatches, currentJob)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, fmt.Sprintf("invalid job state `%s` for `%s` - expected %v", unknownStatePatches[0].Value, unknownStatePatches[0].Path, models.ValidJobStates))

				So(updatedJob, ShouldResemble, models.Job{})
				So(bsonUpdates, ShouldBeEmpty)
			})
		})
	})

	Convey("Given patches with invalid state", t, func() {
		invalidStatePatches := []dprequest.Patch{
			{
				Op:    dprequest.OpReplace.String(),
				Path:  models.JobStatePath,
				Value: 12,
			},
		}

		Convey("When preparePatchUpdates is called", func() {
			updatedJob, bsonUpdates, err := api.GetUpdatesFromJobPatches(testCtx, invalidStatePatches, currentJob)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, fmt.Sprintf("wrong value type `%s` for `%s`, expected string", api.GetValueType(invalidStatePatches[0].Value), invalidStatePatches[0].Path))

				So(updatedJob, ShouldResemble, models.Job{})
				So(bsonUpdates, ShouldBeEmpty)
			})
		})
	})

	Convey("Given patches with invalid total search documents", t, func() {
		invalidTotalSearchDocsPatches := []dprequest.Patch{
			{
				Op:    dprequest.OpReplace.String(),
				Path:  models.JobTotalSearchDocumentsPath,
				Value: "invalid",
			},
		}

		Convey("When preparePatchUpdates is called", func() {
			updatedJob, bsonUpdates, err := api.GetUpdatesFromJobPatches(testCtx, invalidTotalSearchDocsPatches, currentJob)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, fmt.Sprintf("wrong value type `%s` for `%s`, expected an integer", api.GetValueType(invalidTotalSearchDocsPatches[0].Value), invalidTotalSearchDocsPatches[0].Path))

				So(updatedJob, ShouldResemble, models.Job{})
				So(bsonUpdates, ShouldBeEmpty)
			})
		})
	})
}
