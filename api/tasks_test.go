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
	emptyTaskName         = ""
	validTaskName1        = "zebedee"
	validTaskName2        = "dataset-api"
	invalidTaskName       = "any-word-not-in-valid-list"
	validLimit            = 2
	validOffset           = 0
	validServiceAuthToken = "Bearer fc4089e2e12937861377629b0cd96cf79298a4c5d329a2ebb96664c88df77b67"
)

// Create Task Payload
var createTaskPayloadFmt = `{
	"task_name": "%s",
	"number_of_documents": 5
}`

func expectedTask(cfg *config.Config, jobID, taskName string, jsonResponse bool, lastUpdated time.Time, numberOfDocuments int) models.Task {
	task := models.Task{
		JobID:       jobID,
		LastUpdated: lastUpdated,
		Links: &models.TaskLinks{
			Job:  fmt.Sprintf("/jobs/%s", jobID),
			Self: fmt.Sprintf("/jobs/%s/tasks/%s", jobID, taskName),
		},
		NumberOfDocuments: numberOfDocuments,
		TaskName:          taskName,
	}

	if jsonResponse {
		task.Links.Job = fmt.Sprintf("%s/%s%s", cfg.BindAddr, cfg.LatestVersion, task.Links.Job)
		task.Links.Self = fmt.Sprintf("%s/%s%s", cfg.BindAddr, cfg.LatestVersion, task.Links.Self)
	}

	return task
}

func TestCreateTaskHandler(t *testing.T) {
	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	dataStorerMock := &apiMock.DataStorerMock{
		GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
			return nil, nil
		},
		UpsertTaskFunc: func(ctx context.Context, jobID, taskName string, task models.Task) error {
			switch taskName {
			case emptyTaskName:
				return apierrors.ErrEmptyTaskNameProvided
			case invalidTaskName:
				return apierrors.ErrTaskInvalidName
			}

			switch jobID {
			case validJobID1:
				return nil
			case invalidJobID:
				return mongo.ErrJobNotFound
			default:
				return errors.New("an unexpected error occurred")
			}
		},
	}

	Convey("Given an API that can create valid search reindex tasks and store their details in a Data Store", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When a new reindex task is created and stored", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/jobs/%s/tasks", validJobID1), bytes.NewBufferString(
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

				expectedTask := expectedTask(cfg, validJobID1, validTaskName1, true, zeroTime, 5)
				So(resp.Header().Get("Etag"), ShouldNotBeEmpty)

				Convey("And the new task resource should contain expected 	values", func() {
					So(newTask.JobID, ShouldEqual, expectedTask.JobID)
					So(newTask.Links, ShouldResemble, expectedTask.Links)
					So(newTask.NumberOfDocuments, ShouldEqual, expectedTask.NumberOfDocuments)
					So(newTask.TaskName, ShouldEqual, expectedTask.TaskName)
				})
			})
		})
	})

	Convey("Given task name is empty", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/jobs/%s/tasks", validJobID1), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, emptyTaskName)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "invalid request body")
				So(resp.Header().Get("Etag"), ShouldBeEmpty)
			})
		})
	})

	Convey("Given task name is invalid", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/jobs/%s/tasks", validJobID1), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, invalidTaskName)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "invalid request body")
				So(resp.Header().Get("Etag"), ShouldBeEmpty)
			})
		})
	})

	Convey("Given job id is invalid", t, func() {
		getJobFailedMock := &apiMock.DataStorerMock{
			GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
				return nil, mongo.ErrJobNotFound
			},
		}

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), getJobFailedMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/jobs/%s/tasks", invalidJobID), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, validTaskName2)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()
			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 404", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "failed to find the specified reindex job")
				So(resp.Header().Get("Etag"), ShouldBeEmpty)
			})
		})
	})

	Convey("Given job id is empty", t, func() {
		getJobFailedMock := &apiMock.DataStorerMock{
			GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
				return nil, mongo.ErrJobNotFound
			},
		}

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), getJobFailedMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", "http://localhost:25700/jobs//tasks", bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, validTaskName1)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then moved permanently redirection 301 status code is returned by gorilla/mux as it cleans url path", func() {
				So(resp.Code, ShouldEqual, http.StatusMovedPermanently)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldBeEmpty)
				So(resp.Header().Get("Etag"), ShouldBeEmpty)
			})
		})
	})

	Convey("Given an unexpected error occurs in the datastore", t, func() {
		unexpectedErrDataStoreMock := &apiMock.DataStorerMock{
			GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
				return nil, nil
			},
			UpsertTaskFunc: func(ctx context.Context, jobID, taskName string, task models.Task) error {
				return errUnexpected
			},
		}

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), unexpectedErrDataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/jobs/%s/tasks", validJobID1), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, validTaskName1)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "internal server error")
				So(resp.Header().Get("Etag"), ShouldBeEmpty)
			})
		})
	})
}

func TestGetTaskHandler(t *testing.T) {
	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	dataStorerMock := &apiMock.DataStorerMock{
		GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
			job := expectedJob(ctx, t, cfg, true, id, "", 1)
			return &job, nil
		},
		GetTaskFunc: func(ctx context.Context, jobID, taskName string) (*models.Task, error) {
			task := expectedTask(cfg, jobID, taskName, false, zeroTime, 1)
			return &task, nil
		},
	}

	Convey("Given a valid job id and task name which exists in the datastore", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks/%s", validJobID1, validTaskName1), nil)
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

				expectedTask := expectedTask(cfg, validJobID1, validTaskName1, true, zeroTime, 1)

				Convey("And the new task resource should contain expected values", func() {
					So(respTask.JobID, ShouldEqual, expectedTask.JobID)
					So(respTask.Links, ShouldResemble, expectedTask.Links)
					So(respTask.NumberOfDocuments, ShouldEqual, expectedTask.NumberOfDocuments)
					So(respTask.TaskName, ShouldEqual, expectedTask.TaskName)
				})
			})
		})
	})

	Convey("Given job id is empty", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs//tasks/%s", validTaskName1), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then moved permanently redirection 301 status code is returned by gorilla/mux as it cleans url path", func() {
				So(resp.Code, ShouldEqual, http.StatusMovedPermanently)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldBeEmpty)
				So(resp.Header().Get("Etag"), ShouldBeEmpty)
			})
		})
	})

	Convey("Given job id is invalid or job does not exist with the given job id", t, func() {
		jobNotFoundDataStoreMock := &apiMock.DataStorerMock{
			GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
				return nil, mongo.ErrJobNotFound
			},
		}

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), jobNotFoundDataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks/%s", invalidJobID, validTaskName1), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then status code 404 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, apierrors.ErrJobNotFound.Error())
			})
		})
	})

	Convey("Given task name is empty", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks//", validJobID1), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then moved permanently redirection 301 status code is returned by gorilla/mux as it cleans url path", func() {
				So(resp.Code, ShouldEqual, http.StatusMovedPermanently)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldBeEmpty)
			})
		})
	})

	Convey("Given task name is invalid or task does not exist with the given task name", t, func() {
		invalidTaskNameDataStoreMock := &apiMock.DataStorerMock{
			GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
				job := expectedJob(ctx, t, cfg, true, id, "", 1)
				return &job, nil
			},
			GetTaskFunc: func(ctx context.Context, jobID, taskName string) (*models.Task, error) {
				return nil, mongo.ErrTaskNotFound
			},
		}

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), invalidTaskNameDataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks/%s", validJobID1, invalidTaskName), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then status code 404 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, apierrors.ErrTaskNotFound.Error())
			})
		})
	})

	Convey("Given an unexpected error occurs in the datastore", t, func() {
		unexpectedErrDataStoreMock := &apiMock.DataStorerMock{
			GetJobFunc: func(ctx context.Context, id string) (*models.Job, error) {
				return nil, errUnexpected
			},
		}

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), unexpectedErrDataStoreMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get task", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks/%s", validJobID1, invalidTaskName), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then status code 500 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, apierrors.ErrInternalServer.Error())
			})
		})
	})
}

func TestGetTasksHandler(t *testing.T) {
	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to retrieve default configuration, error: %v", err)
	}

	dataStorerMock := &apiMock.DataStorerMock{
		AcquireJobLockFunc: func(ctx context.Context, id string) (string, error) {
			switch id {
			case unLockableJobID:
				return "", errUnexpected
			default:
				return "", nil
			}
		},
		UnlockJobFunc: func(ctx context.Context, lockID string) {
			// mock UnlockJob to be successful by doing nothing
		},

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
				expectedTask1 := expectedTask(cfg, jobID, validTaskName1, false, zeroTime, 0)
				expectedTask2 := expectedTask(cfg, jobID, validTaskName2, false, zeroTime, 0)

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

	Convey("Given a valid job id, valid pagination parameters and no If-Match header set", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks?offset=%d&limit=%d", validJobID1, validOffset, validLimit), nil)
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

				expectedTask1 := expectedTask(cfg, validJobID1, validTaskName1, true, zeroTime, 0)
				expectedTask2 := expectedTask(cfg, validJobID1, validTaskName2, true, zeroTime, 0)
				expectedTasks := &models.Tasks{
					Count:      2,
					TaskList:   []models.Task{expectedTask1, expectedTask2},
					Limit:      2,
					Offset:     0,
					TotalCount: 2,
				}

				Convey("And the new task resource should contain expected values", func() {
					So(respTasks.Count, ShouldEqual, expectedTasks.Count)
					So(respTasks.TaskList, ShouldResemble, expectedTasks.TaskList)
					So(respTasks.Limit, ShouldEqual, expectedTasks.Limit)
					So(respTasks.Offset, ShouldEqual, expectedTasks.Offset)
					So(respTasks.TotalCount, ShouldEqual, expectedTasks.TotalCount)

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
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks?offset=%d&limit=%d", validJobID1, validOffset, validLimit), nil)
			headers.SetIfMatch(req, `"024613bed9e9db8b730c94a98a420c2ef7e39a09"`)

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

				expectedTask1 := expectedTask(cfg, validJobID1, validTaskName1, true, zeroTime, 0)
				expectedTask2 := expectedTask(cfg, validJobID1, validTaskName2, true, zeroTime, 0)
				expectedTasks := &models.Tasks{
					Count:      2,
					TaskList:   []models.Task{expectedTask1, expectedTask2},
					Limit:      2,
					Offset:     0,
					TotalCount: 2,
				}

				Convey("And the new task resource should contain expected values", func() {
					So(respTasks.Count, ShouldEqual, expectedTasks.Count)
					So(respTasks.TaskList, ShouldResemble, expectedTasks.TaskList)
					So(respTasks.Limit, ShouldEqual, expectedTasks.Limit)
					So(respTasks.Offset, ShouldEqual, expectedTasks.Offset)
					So(respTasks.TotalCount, ShouldEqual, expectedTasks.TotalCount)

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
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks?offset=%d&limit=%d", validJobID1, validOffset, validLimit), nil)
			headers.SetIfMatch(req, "")

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

				expectedTask1 := expectedTask(cfg, validJobID1, validTaskName1, true, zeroTime, 0)
				expectedTask2 := expectedTask(cfg, validJobID1, validTaskName2, true, zeroTime, 0)
				expectedTasks := &models.Tasks{
					Count:      2,
					TaskList:   []models.Task{expectedTask1, expectedTask2},
					Limit:      2,
					Offset:     0,
					TotalCount: 2,
				}

				Convey("And the new task resource should contain expected values", func() {
					So(respTasks.Count, ShouldEqual, expectedTasks.Count)
					So(respTasks.TaskList, ShouldResemble, expectedTasks.TaskList)
					So(respTasks.Limit, ShouldEqual, expectedTasks.Limit)
					So(respTasks.Offset, ShouldEqual, expectedTasks.Offset)
					So(respTasks.TotalCount, ShouldEqual, expectedTasks.TotalCount)

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
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks?offset=%d&limit=%d", validJobID1, validOffset, validLimit), nil)
			headers.SetIfMatch(req, "*")

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

				expectedTask1 := expectedTask(cfg, validJobID1, validTaskName1, true, zeroTime, 0)
				expectedTask2 := expectedTask(cfg, validJobID1, validTaskName2, true, zeroTime, 0)
				expectedTasks := &models.Tasks{
					Count:      2,
					TaskList:   []models.Task{expectedTask1, expectedTask2},
					Limit:      2,
					Offset:     0,
					TotalCount: 2,
				}

				Convey("And the new task resource should contain expected values", func() {
					So(respTasks.Count, ShouldEqual, expectedTasks.Count)
					So(respTasks.TaskList, ShouldResemble, expectedTasks.TaskList)
					So(respTasks.Limit, ShouldEqual, expectedTasks.Limit)
					So(respTasks.Offset, ShouldEqual, expectedTasks.Offset)
					So(respTasks.TotalCount, ShouldEqual, expectedTasks.TotalCount)

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

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks?offset=%d&limit=%d", validJobID1, validOffset, validLimit), nil)
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

	Convey("Given invalid pagination parameter is given", t, func() {
		invalidLimit := "abc"

		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks?offset=%d&limit=%s", validJobID1, validOffset, invalidLimit), nil)
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
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs//tasks?offset=%d&limit=%d", validOffset, validLimit), nil)
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
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks?offset=%d&limit=%d", invalidJobID, validOffset, validLimit), nil)
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

	Convey("Given datastore is unable to lock job id", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks?offset=%d&limit=%d", unLockableJobID, validOffset, validLimit), nil)
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

	Convey("Given an unexpected error occurs in the datastore", t, func() {
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When request is made to get tasks", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:25700/jobs/%s/tasks?offset=%d&limit=%d", validJobID2, validOffset, validLimit), nil)
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
