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
	"github.com/ONSdigital/dp-search-reindex-api/api"
	apiMock "github.com/ONSdigital/dp-search-reindex-api/api/mock"
	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

// Constants for testing
const (
	invalidJobID          = "UUID3"
	validTaskName         = "test-task"
	emptyTaskName         = ""
	invalidTaskName       = "any-word-not-in-valid-list"
	invalidTaskRetrieve   = "non-retrievable-task"
	validLimit            = 2
	validOffset           = 0
	validServiceAuthToken = "Bearer fc4089e2e12937861377629b0cd96cf79298a4c5d329a2ebb96664c88df77b67"
	validTaskName1        = "zebedee"
	validTaskName2        = "dataset-api"
)

// Create Task Payload
var createTaskPayloadFmt = `{
	"task_name": "%s",
	"number_of_documents": 5
}`

func expectedTask(ctx context.Context, cfg *config.Config, t *testing.T, jobID, taskName string, jsonResponse bool, lastUpdated time.Time, numberOfDocuments int) models.Task {
	task := models.Task{
		JobID:       jobID,
		LastUpdated: lastUpdated,
		Links: &models.TaskLinks{
			Job:  fmt.Sprintf("/search-reindex-jobs/%s", jobID),
			Self: fmt.Sprintf("/search-reindex-jobs/%s/tasks/%s", jobID, taskName),
		},
		NumberOfDocuments: numberOfDocuments,
		TaskName:          taskName,
	}

	taskETag, err := models.GenerateETagForTask(ctx, task)
	if err != nil {
		t.Errorf("failed to generate eTag for expected test task - error: %v", err)
	}
	task.ETag = taskETag

	if jsonResponse {
		task.Links.Job = fmt.Sprintf("%s/%s%s", cfg.BindAddr, cfg.LatestVersion, task.Links.Job)
		task.Links.Self = fmt.Sprintf("%s/%s%s", cfg.BindAddr, cfg.LatestVersion, task.Links.Self)
	}

	return task
}

func expectedTasks(ctx context.Context, t *testing.T, cfg *config.Config, jsonResponse bool, jobID string, limit, offset int) models.Tasks {
	var firstTask, secondTask models.Task

	tasks := models.Tasks{
		Limit:      limit,
		Offset:     offset,
		TotalCount: 2,
	}

	firstTask = expectedTask(ctx, cfg, t, jobID, validTaskName1, jsonResponse, zeroTime, 0)
	secondTask = expectedTask(ctx, cfg, t, jobID, validTaskName2, jsonResponse, zeroTime, 0)

	if (offset == 0) && (limit > 1) {
		tasks.Count = 2
		tasks.TaskList = []models.Task{firstTask, secondTask}
	}

	if (offset == 1) && (limit > 0) {
		tasks.Count = 1
		tasks.TaskList = []models.Task{secondTask}
	}

	return tasks
}

func TestCreateTaskHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	dataStorerMock := &apiMock.DataStorerMock{
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
		GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
			switch id {
			case invalidJobID:
				return nil, mongo.ErrJobNotFound
			default:
				job := expectedJob(ctx, t, cfg, false, id, "", 0)
				return &job, nil
			}
		},
		UpsertTaskFunc: func(ctx context.Context, jobID, taskName string, task models.Task) error {
			switch taskName {
			case emptyTaskName:
				return apierrors.ErrEmptyTaskNameProvided
			case invalidTaskName:
				return apierrors.ErrTaskInvalidName
			}

			switch jobID {
			case validJobID2:
				return errUnexpected
			default:
				return nil
			}
		},
	}

	Convey("Given an API that can create valid search reindex tasks and store their details in a Data Store", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a new reindex task is created and stored", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks", validJobID1), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, validTaskName1)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the newly created search reindex task is returned with status code 201", func() {
				So(resp.Code, ShouldEqual, http.StatusCreated)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				newTask := models.Task{}
				err = json.Unmarshal(payload, &newTask)
				So(err, ShouldBeNil)

				expectedTask := expectedTask(ctx, cfg, t, validJobID1, validTaskName1, true, zeroTime, 5)

				Convey("And the new task resource should contain expected values", func() {
					So(newTask.JobID, ShouldEqual, expectedTask.JobID)
					So(newTask.Links, ShouldResemble, expectedTask.Links)
					So(newTask.NumberOfDocuments, ShouldEqual, expectedTask.NumberOfDocuments)
					So(newTask.TaskName, ShouldEqual, expectedTask.TaskName)

					Convey("And the etag of the resource should be returned via the ETag header", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
					})
				})
			})
		})
	})

	Convey("Given task name is empty", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks", validJobID1), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, emptyTaskName)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an error is returned with status code 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "invalid request body")

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given task name is invalid", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks", validJobID1), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, invalidTaskName)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an error is returned with status code 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "invalid request body")

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given job id is invalid", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks", invalidJobID), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, validTaskName2)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an error is returned with status code 404", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "failed to find the specified reindex job")

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given job id is empty", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", "http://localhost:25700/search-reindex-jobs//tasks", bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, validTaskName1)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then moved permanently redirection 301 status code is returned by gorilla/mux as it cleans url path", func() {
				So(resp.Code, ShouldEqual, http.StatusMovedPermanently)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldBeEmpty)

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given datastore failed to lock job id", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks", unLockableJobID), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, validTaskName1)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an error is returned with status code 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "internal server error")

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given an unexpected error occurs in the datastore", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks", validJobID2), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, validTaskName1)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an error is returned with status code 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "internal server error")

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})
}

func TestGetTaskHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	dataStorerMock := &apiMock.DataStorerMock{
		GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
			switch id {
			case validJobID2:
				return nil, errUnexpected
			case invalidJobID:
				return nil, mongo.ErrJobNotFound
			default:
				job := expectedJob(ctx, t, cfg, true, id, "", 1)
				return &job, nil
			}
		},
		GetTaskFunc: func(ctx context.Context, jobID, taskName string) (*models.Task, error) {
			switch taskName {
			case invalidTaskName:
				return nil, mongo.ErrTaskNotFound
			default:
				task := expectedTask(ctx, cfg, t, jobID, taskName, false, zeroTime, 1)
				return &task, nil
			}
		},
	}

	Convey("Given a valid job id and task name and no If-Match header set", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s", validJobID1, validTaskName1), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then 200 status code should be returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				respTask := models.Task{}
				err = json.Unmarshal(payload, &respTask)
				So(err, ShouldBeNil)

				expectedTask := expectedTask(ctx, cfg, t, validJobID1, validTaskName1, true, zeroTime, 1)

				Convey("And the new task resource should contain expected values", func() {
					So(respTask.JobID, ShouldEqual, expectedTask.JobID)
					So(respTask.Links, ShouldResemble, expectedTask.Links)
					So(respTask.NumberOfDocuments, ShouldEqual, expectedTask.NumberOfDocuments)
					So(respTask.TaskName, ShouldEqual, expectedTask.TaskName)

					Convey("And the etag of the resource should be returned via the ETag header", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
					})
				})
			})
		})
	})

	Convey("Given If-Match header set to *", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s", validJobID1, validTaskName1), nil)
			headers.SetIfMatch(req, "*")

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then 200 status code should be returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				respTask := models.Task{}
				err = json.Unmarshal(payload, &respTask)
				So(err, ShouldBeNil)

				expectedTask := expectedTask(ctx, cfg, t, validJobID1, validTaskName1, true, zeroTime, 1)

				Convey("And the new task resource should contain expected values", func() {
					So(respTask.JobID, ShouldEqual, expectedTask.JobID)
					So(respTask.Links, ShouldResemble, expectedTask.Links)
					So(respTask.NumberOfDocuments, ShouldEqual, expectedTask.NumberOfDocuments)
					So(respTask.TaskName, ShouldEqual, expectedTask.TaskName)

					Convey("And the etag of the resource should be returned via the ETag header", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
					})
				})
			})
		})
	})

	Convey("Given valid etag", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s", validJobID1, validTaskName1), nil)
			headers.SetIfMatch(req, `"41d158dea3ce9221021648969926a61018cecd35"`)

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then 200 status code should be returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				respTask := models.Task{}
				err = json.Unmarshal(payload, &respTask)
				So(err, ShouldBeNil)

				expectedTask := expectedTask(ctx, cfg, t, validJobID1, validTaskName1, true, zeroTime, 1)

				Convey("And the new task resource should contain expected values", func() {
					So(respTask.JobID, ShouldEqual, expectedTask.JobID)
					So(respTask.Links, ShouldResemble, expectedTask.Links)
					So(respTask.NumberOfDocuments, ShouldEqual, expectedTask.NumberOfDocuments)
					So(respTask.TaskName, ShouldEqual, expectedTask.TaskName)

					Convey("And the etag of the resource should be returned via the ETag header", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
					})
				})
			})
		})
	})

	Convey("Given empty etag", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s", validJobID1, validTaskName1), nil)
			headers.SetIfMatch(req, "")

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then 200 status code should be returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				respTask := models.Task{}
				err = json.Unmarshal(payload, &respTask)
				So(err, ShouldBeNil)

				expectedTask := expectedTask(ctx, cfg, t, validJobID1, validTaskName1, true, zeroTime, 1)

				Convey("And the new task resource should contain expected values", func() {
					So(respTask.JobID, ShouldEqual, expectedTask.JobID)
					So(respTask.Links, ShouldResemble, expectedTask.Links)
					So(respTask.NumberOfDocuments, ShouldEqual, expectedTask.NumberOfDocuments)
					So(respTask.TaskName, ShouldEqual, expectedTask.TaskName)

					Convey("And the etag of the resource should be returned via the ETag header", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
					})
				})
			})
		})
	})

	Convey("Given outdated or invalid etag", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s", validJobID1, validTaskName1), nil)
			headers.SetIfMatch(req, "invalid")

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a conflict with etag error is returned with status code 409", func() {
				So(resp.Code, ShouldEqual, http.StatusConflict)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, apierrors.ErrConflictWithETag.Error())

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given job id is empty", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs//tasks/%s", validTaskName1), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then moved permanently redirection 301 status code is returned by gorilla/mux as it cleans url path", func() {
				So(resp.Code, ShouldEqual, http.StatusMovedPermanently)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldBeEmpty)

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given job id is invalid or job does not exist with the given job id", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s", invalidJobID, validTaskName1), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then status code 404 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, apierrors.ErrJobNotFound.Error())

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given task name is empty", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks//", validJobID1), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then moved permanently redirection 301 status code is returned by gorilla/mux as it cleans url path", func() {
				So(resp.Code, ShouldEqual, http.StatusMovedPermanently)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldBeEmpty)

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given task name is invalid or task does not exist with the given task name", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s", validJobID1, invalidTaskName), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then status code 404 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, apierrors.ErrTaskNotFound.Error())

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given an unexpected error occurs in the datastore", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s", validJobID2, validTaskName1), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then status code 500 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, apierrors.ErrInternalServer.Error())

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})
}

func TestGetTasksHandlerSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	dataStorerMock := &apiMock.DataStorerMock{
		GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
			job := expectedJob(ctx, t, cfg, true, id, "", 2)
			return &job, nil
		},

		GetTasksFunc: func(ctx context.Context, jobID string, options mongo.Options) (*models.Tasks, error) {
			expectedTask1 := expectedTask(ctx, cfg, t, jobID, validTaskName1, false, zeroTime, 0)
			expectedTask2 := expectedTask(ctx, cfg, t, jobID, validTaskName2, false, zeroTime, 0)

			tasks := &models.Tasks{
				Count:      2,
				TaskList:   []models.Task{expectedTask1, expectedTask2},
				Limit:      options.Limit,
				Offset:     options.Offset,
				TotalCount: 2,
			}
			return tasks, nil
		},
	}

	Convey("Given a valid job id, valid pagination parameters and no If-Match header set", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks?offset=%d&limit=%d", validJobID1, validOffset, validLimit), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then 200 status code should be returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				respTasks := models.Tasks{}
				err = json.Unmarshal(payload, &respTasks)
				So(err, ShouldBeNil)

				expectedTasks := expectedTasks(ctx, t, cfg, true, validJobID1, validLimit, validOffset)
				expectedTask1 := expectedTasks.TaskList[0]
				expectedTask2 := expectedTasks.TaskList[1]

				Convey("And the new task resource should contain expected values", func() {
					So(respTasks.Count, ShouldEqual, expectedTasks.Count)
					So(respTasks.Limit, ShouldEqual, expectedTasks.Limit)
					So(respTasks.Offset, ShouldEqual, expectedTasks.Offset)
					So(respTasks.TotalCount, ShouldEqual, expectedTasks.TotalCount)

					respTaskList := respTasks.TaskList
					So(respTaskList, ShouldHaveLength, 2)

					respTask1 := respTaskList[0]
					So(respTask1.ETag, ShouldBeEmpty)
					So(respTask1.JobID, ShouldEqual, expectedTask1.JobID)
					So(respTask1.LastUpdated, ShouldEqual, expectedTask1.LastUpdated)
					So(respTask1.Links.Job, ShouldEqual, expectedTask1.Links.Job)
					So(respTask1.Links.Self, ShouldEqual, expectedTask1.Links.Self)
					So(respTask1.NumberOfDocuments, ShouldEqual, expectedTask1.NumberOfDocuments)
					So(respTask1.TaskName, ShouldEqual, expectedTask1.TaskName)

					respTask2 := respTaskList[1]
					So(respTask2.ETag, ShouldBeEmpty)
					So(respTask2.JobID, ShouldEqual, expectedTask2.JobID)
					So(respTask2.LastUpdated, ShouldEqual, expectedTask2.LastUpdated)
					So(respTask2.Links.Job, ShouldEqual, expectedTask2.Links.Job)
					So(respTask2.Links.Self, ShouldEqual, expectedTask2.Links.Self)
					So(respTask2.NumberOfDocuments, ShouldEqual, expectedTask2.NumberOfDocuments)
					So(respTask2.TaskName, ShouldEqual, expectedTask2.TaskName)

					Convey("And the etag of the resource should be returned via the ETag header", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
					})
				})
			})
		})
	})

	Convey("Given valid etag", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks?offset=%d&limit=%d", validJobID1, validOffset, validLimit), nil)

			err := headers.SetIfMatch(req, `"d9bc1d82addbd6ff0ab9259dde8eeb6c2324d42b"`)
			if err != nil {
				t.Errorf("failed to set if-match header, error: %v", err)
			}

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then 200 status code should be returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				respTasks := models.Tasks{}
				err = json.Unmarshal(payload, &respTasks)
				So(err, ShouldBeNil)

				expectedTasks := expectedTasks(ctx, t, cfg, true, validJobID1, validLimit, validOffset)
				expectedTask1 := expectedTasks.TaskList[0]
				expectedTask2 := expectedTasks.TaskList[1]

				Convey("And the new task resource should contain expected values", func() {
					So(respTasks.Count, ShouldEqual, expectedTasks.Count)
					So(respTasks.Limit, ShouldEqual, expectedTasks.Limit)
					So(respTasks.Offset, ShouldEqual, expectedTasks.Offset)
					So(respTasks.TotalCount, ShouldEqual, expectedTasks.TotalCount)

					respTaskList := respTasks.TaskList
					So(respTaskList, ShouldHaveLength, 2)

					respTask1 := respTaskList[0]
					So(respTask1.ETag, ShouldBeEmpty)
					So(respTask1.JobID, ShouldEqual, expectedTask1.JobID)
					So(respTask1.LastUpdated, ShouldEqual, expectedTask1.LastUpdated)
					So(respTask1.Links.Job, ShouldEqual, expectedTask1.Links.Job)
					So(respTask1.Links.Self, ShouldEqual, expectedTask1.Links.Self)
					So(respTask1.NumberOfDocuments, ShouldEqual, expectedTask1.NumberOfDocuments)
					So(respTask1.TaskName, ShouldEqual, expectedTask1.TaskName)

					respTask2 := respTaskList[1]
					So(respTask2.ETag, ShouldBeEmpty)
					So(respTask2.JobID, ShouldEqual, expectedTask2.JobID)
					So(respTask2.LastUpdated, ShouldEqual, expectedTask2.LastUpdated)
					So(respTask2.Links.Job, ShouldEqual, expectedTask2.Links.Job)
					So(respTask2.Links.Self, ShouldEqual, expectedTask2.Links.Self)
					So(respTask2.NumberOfDocuments, ShouldEqual, expectedTask2.NumberOfDocuments)
					So(respTask2.TaskName, ShouldEqual, expectedTask2.TaskName)

					Convey("And the etag of the resource should be returned via the ETag header", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
					})
				})
			})
		})
	})

	Convey("Given empty etag", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks?offset=%d&limit=%d", validJobID1, validOffset, validLimit), nil)

			err := headers.SetIfMatch(req, "")
			if err != nil {
				t.Errorf("failed to set if-match header, error: %v", err)
			}

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then 200 status code should be returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				respTasks := models.Tasks{}
				err = json.Unmarshal(payload, &respTasks)
				So(err, ShouldBeNil)

				expectedTasks := expectedTasks(ctx, t, cfg, true, validJobID1, validLimit, validOffset)
				expectedTask1 := expectedTasks.TaskList[0]
				expectedTask2 := expectedTasks.TaskList[1]

				Convey("And the new task resource should contain expected values", func() {
					So(respTasks.Count, ShouldEqual, expectedTasks.Count)
					So(respTasks.Limit, ShouldEqual, expectedTasks.Limit)
					So(respTasks.Offset, ShouldEqual, expectedTasks.Offset)
					So(respTasks.TotalCount, ShouldEqual, expectedTasks.TotalCount)

					respTaskList := respTasks.TaskList
					So(respTaskList, ShouldHaveLength, 2)

					respTask1 := respTaskList[0]
					So(respTask1.ETag, ShouldBeEmpty)
					So(respTask1.JobID, ShouldEqual, expectedTask1.JobID)
					So(respTask1.LastUpdated, ShouldEqual, expectedTask1.LastUpdated)
					So(respTask1.Links.Job, ShouldEqual, expectedTask1.Links.Job)
					So(respTask1.Links.Self, ShouldEqual, expectedTask1.Links.Self)
					So(respTask1.NumberOfDocuments, ShouldEqual, expectedTask1.NumberOfDocuments)
					So(respTask1.TaskName, ShouldEqual, expectedTask1.TaskName)

					respTask2 := respTaskList[1]
					So(respTask2.ETag, ShouldBeEmpty)
					So(respTask2.JobID, ShouldEqual, expectedTask2.JobID)
					So(respTask2.LastUpdated, ShouldEqual, expectedTask2.LastUpdated)
					So(respTask2.Links.Job, ShouldEqual, expectedTask2.Links.Job)
					So(respTask2.Links.Self, ShouldEqual, expectedTask2.Links.Self)
					So(respTask2.NumberOfDocuments, ShouldEqual, expectedTask2.NumberOfDocuments)
					So(respTask2.TaskName, ShouldEqual, expectedTask2.TaskName)

					Convey("And the etag of the resource should be returned via the ETag header", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
					})
				})
			})
		})
	})

	Convey("Given If-Match header set to *", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks?offset=%d&limit=%d", validJobID1, validOffset, validLimit), nil)

			err := headers.SetIfMatch(req, "*")
			if err != nil {
				t.Errorf("failed to set if-match header, error: %v", err)
			}

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then 200 status code should be returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				payload, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("failed to read payload with io.ReadAll, error: %v", err)
				}

				respTasks := models.Tasks{}
				err = json.Unmarshal(payload, &respTasks)
				So(err, ShouldBeNil)

				expectedTasks := expectedTasks(ctx, t, cfg, true, validJobID1, validLimit, validOffset)
				expectedTask1 := expectedTasks.TaskList[0]
				expectedTask2 := expectedTasks.TaskList[1]

				Convey("And the new task resource should contain expected values", func() {
					So(respTasks.Count, ShouldEqual, expectedTasks.Count)
					So(respTasks.Limit, ShouldEqual, expectedTasks.Limit)
					So(respTasks.Offset, ShouldEqual, expectedTasks.Offset)
					So(respTasks.TotalCount, ShouldEqual, expectedTasks.TotalCount)

					respTaskList := respTasks.TaskList
					So(respTaskList, ShouldHaveLength, 2)

					respTask1 := respTaskList[0]
					So(respTask1.ETag, ShouldBeEmpty)
					So(respTask1.JobID, ShouldEqual, expectedTask1.JobID)
					So(respTask1.LastUpdated, ShouldEqual, expectedTask1.LastUpdated)
					So(respTask1.Links.Job, ShouldEqual, expectedTask1.Links.Job)
					So(respTask1.Links.Self, ShouldEqual, expectedTask1.Links.Self)
					So(respTask1.NumberOfDocuments, ShouldEqual, expectedTask1.NumberOfDocuments)
					So(respTask1.TaskName, ShouldEqual, expectedTask1.TaskName)

					respTask2 := respTaskList[1]
					So(respTask2.ETag, ShouldBeEmpty)
					So(respTask2.JobID, ShouldEqual, expectedTask2.JobID)
					So(respTask2.LastUpdated, ShouldEqual, expectedTask2.LastUpdated)
					So(respTask2.Links.Job, ShouldEqual, expectedTask2.Links.Job)
					So(respTask2.Links.Self, ShouldEqual, expectedTask2.Links.Self)
					So(respTask2.NumberOfDocuments, ShouldEqual, expectedTask2.NumberOfDocuments)
					So(respTask2.TaskName, ShouldEqual, expectedTask2.TaskName)
					Convey("And the etag of the resource should be returned via the ETag header", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
					})
				})
			})
		})
	})
}

func TestPutNumDocCountHandler(t *testing.T) {
	t.Parallel()

	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	dataStoreMock := &apiMock.DataStorerMock{
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
		GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
			switch id {
			case notFoundJobID:
				return nil, mongo.ErrJobNotFound
			default:
				jobs := expectedJob(ctx, t, cfg, false, id, "", 0)
				return &jobs, nil
			}
		},
		GetTaskFunc: func(ctx context.Context, jobID, taskName string) (*models.Task, error) {
			switch taskName {
			case invalidTaskName:
				return nil, mongo.ErrTaskNotFound
			case invalidTaskRetrieve:
				return nil, errors.New("something bad happened")
			default:
				task := expectedTask(ctx, cfg, t, jobID, taskName, false, time.Now(), 4)
				return &task, nil
			}
		},
		UpsertTaskFunc: func(ctx context.Context, jobID, taskName string, task models.Task) error {
			switch taskName {
			case emptyTaskName:
				return apierrors.ErrEmptyTaskNameProvided
			case invalidTaskRetrieve:
				return apierrors.ErrTaskInvalidName
			}

			switch jobID {
			case validJobID2:
				return nil
			default:
				return errors.New("something bad happened")
			}
		},
	}

	Convey("Given valid job id, valid value for number of tasks and no If-Match header is set", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of tasks of a specific job", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", validJobID2, validTaskName, "1"), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a status code 204 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusNoContent)

				Convey("And the etag of the response job resource should be returned via the ETag header", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
				})
			})
		})
	})

	Convey("Given If-Match header is set to *", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of tasks of a specific job", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", validJobID2, validTaskName, "1"), nil)

			err := headers.SetIfMatch(req, "*")
			if err != nil {
				t.Errorf("failed to set if-match header, error: %v", err)
			}

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a status code 204 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusNoContent)

				Convey("And the etag of the response job resource should be returned via the ETag header", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
				})
			})
		})
	})

	Convey("Given a valid etag is set in the If-Match header", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of documents of a specific task", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", validJobID2, validTaskName, "14"), nil)
			currentETag := `"8e219bfd578dd529849647bf73cabe3ef1463941"`

			err := headers.SetIfMatch(req, currentETag)
			if err != nil {
				t.Errorf("failed to set if-match header, error: %v", err)
			}

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a status code 204 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusNoContent)

				Convey("And the etag of the response job resource should be returned via the ETag header", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotEqual, currentETag)
				})
			})
		})
	})

	Convey("Given an empty etag is set in the If-Match header", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of tasks of a specific job", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", validJobID2, validTaskName, "14"), nil)

			err := headers.SetIfMatch(req, "")
			if err != nil {
				t.Errorf("failed to set if-match header, error: %v", err)
			}

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a status code 204 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusNoContent)

				Convey("And the etag of the response job resource should be returned via the ETag header", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
				})
			})
		})
	})

	Convey("Given an outdated or invalid etag is set in the If-Match header", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of tasks of a specific job", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", validJobID2, validTaskName, "14"), nil)

			err := headers.SetIfMatch(req, "invalid")
			if err != nil {
				t.Errorf("failed to set if-match header, error: %v", err)
			}

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a conflict with etag error is returned with status code 409", func() {
				So(resp.Code, ShouldEqual, http.StatusConflict)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, apierrors.ErrConflictWithETag.Error())

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given a specific job does not exist in the Data Store", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of tasks of a specific job", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", validJobID2, invalidTaskName, "14"), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then job resource was not found returning a status code of 404", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "failed to find the specified task for the reindex job")

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given the value of number of tasks is not an integer", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of docs of a specific task", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", validJobID2, invalidTaskName, "-1"), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then it is a bad request returning a status code of 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "number of docs must be a positive integer")

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given the value of number of tasks is a negative integer", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of tasks of a specific job", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", validJobID2, invalidTaskName, "-1"), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then it is a bad request returning a status code of 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "number of docs must be a positive integer")

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given the Data Store is unable to lock the job id", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of tasks of a specific job", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", unLockableJobID, validTaskName, "3"), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an error with status code 500 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedServerErrorMsg)

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given the request results in no modifications to the job resource", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of tasks of a specific job", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", validJobID2, validTaskName, "4"), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a status code 304 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusNotModified)

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldNotBeEmpty)
				})
			})
		})
	})

	Convey("Given an unexpected error occurs in the Data Store", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a request is made to update the number of tasks of a specific job", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks/%s/number-of-documents/%s", validJobID2, invalidTaskRetrieve, "14"), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the response returns a status code of 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "internal server error")

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})
}

func TestGetTasksHandlerFail(t *testing.T) {
	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	dataStorerMock := &apiMock.DataStorerMock{
		GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
			switch id {
			case validJobID1, validJobID2:
				job := expectedJob(ctx, t, cfg, true, id, "", 2)
				return &job, nil
			case invalidJobID:
				return nil, mongo.ErrJobNotFound
			default:
				return nil, errUnexpected
			}
		},

		GetTasksFunc: func(ctx context.Context, jobID string, options mongo.Options) (*models.Tasks, error) {
			switch jobID {
			case validJobID2:
				return nil, errUnexpected
			default:
				expectedTask1 := expectedTask(ctx, cfg, t, jobID, validTaskName1, false, zeroTime, 0)
				expectedTask2 := expectedTask(ctx, cfg, t, jobID, validTaskName2, false, zeroTime, 0)

				tasks := &models.Tasks{
					Count:      2,
					TaskList:   []models.Task{expectedTask1, expectedTask2},
					Limit:      options.Limit,
					Offset:     options.Offset,
					TotalCount: 2,
				}
				return tasks, nil
			}
		},
	}

	Convey("Given outdated or invalid etag", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks?offset=%d&limit=%d", validJobID1, validOffset, validLimit), nil)

			err := headers.SetIfMatch(req, "invalid")
			if err != nil {
				t.Errorf("failed to set if-match header, error: %v", err)
			}

			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a conflict with etag error is returned with status code 409", func() {
				So(resp.Code, ShouldEqual, http.StatusConflict)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, apierrors.ErrConflictWithETag.Error())

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given invalid pagination parameter is given", t, func() {
		invalidLimit := "abc"

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks?offset=%d&limit=%s", validJobID1, validOffset, invalidLimit), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then 400 status code should be returned", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)

				Convey("And an error message should be returned in the response body", func() {
					errMsg := strings.TrimSpace(resp.Body.String())
					So(errMsg, ShouldEqual, apierrors.ErrInvalidLimitParameter.Error())

					Convey("And the response ETag header should be empty", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
					})
				})
			})
		})
	})

	Convey("Given job id is empty", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs//tasks?offset=%d&limit=%d", validOffset, validLimit), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then moved permanently redirection 301 status code is returned by gorilla/mux as it cleans url path", func() {
				So(resp.Code, ShouldEqual, http.StatusMovedPermanently)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldBeEmpty)

				Convey("And the response ETag header should be empty", func() {
					So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given job id is invalid or job does not exist with the given job id", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks?offset=%d&limit=%d", invalidJobID, validOffset, validLimit), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then status code 404 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)

				Convey("And an error message should be returned in the response body", func() {
					errMsg := strings.TrimSpace(resp.Body.String())
					So(errMsg, ShouldEqual, apierrors.ErrJobNotFound.Error())

					Convey("And the response ETag header should be empty", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
					})
				})
			})
		})
	})

	Convey("Given an unexpected error occurs in the datastore", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/search-reindex-jobs/%s/tasks?offset=%d&limit=%d", validJobID2, validOffset, validLimit), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then status code 500 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)

				Convey("And an error message should be returned in the response body", func() {
					errMsg := strings.TrimSpace(resp.Body.String())
					So(errMsg, ShouldEqual, apierrors.ErrInternalServer.Error())

					Convey("And the response ETag header should be empty", func() {
						So(resp.Header().Get(dpresponse.ETagHeader), ShouldBeEmpty)
					})
				})
			})
		})
	})
}
