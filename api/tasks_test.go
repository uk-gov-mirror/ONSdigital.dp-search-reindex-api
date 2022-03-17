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

	dpHTTP "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-search-reindex-api/api"
	apiMock "github.com/ONSdigital/dp-search-reindex-api/api/mock"
	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/dp-search-reindex-api/url"
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
	validServiceAuthToken = "Bearer fc4089e2e12937861377629b0cd96cf79298a4c5d329a2ebb96664c88df77b67"
	bindAddress           = "localhost:25700"
)

func TestCreateTaskHandler(t *testing.T) {
	dataStorerMock := &apiMock.DataStorerMock{
		CreateTaskFunc: func(ctx context.Context, jobID string, taskName string, numDocuments int) (models.Task, error) {
			emptyTask := models.Task{}

			switch taskName {
			case emptyTaskName:
				return emptyTask, apierrors.ErrEmptyTaskNameProvided
			case invalidTaskName:
				return emptyTask, apierrors.ErrTaskInvalidName
			}

			switch jobID {
			case validJobID1:
				return models.NewTask(jobID, taskName, numDocuments, bindAddress), nil
			case invalidJobID:
				return emptyTask, mongo.ErrJobNotFound
			default:
				return emptyTask, errors.New("an unexpected error occurred")
			}
		},
	}

	Convey("Given an API that can create valid search reindex tasks and store their details in a Data Store", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
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
				So(err, ShouldBeNil)
				newTask := models.Task{}
				err = json.Unmarshal(payload, &newTask)
				So(err, ShouldBeNil)
				zeroTime := time.Time{}.UTC()
				expectedTask, err := ExpectedTask(validJobID1, zeroTime, 5, validTaskName1)
				So(err, ShouldBeNil)

				Convey("And the new task resource should contain expected 	values", func() {
					So(newTask.JobID, ShouldEqual, expectedTask.JobID)
					So(newTask.Links, ShouldResemble, expectedTask.Links)
					So(newTask.NumberOfDocuments, ShouldEqual, expectedTask.NumberOfDocuments)
					So(newTask.TaskName, ShouldEqual, expectedTask.TaskName)
				})
			})
		})
	})

	Convey("Given an API that can create valid search reindex tasks and store their details in a Data Store", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/jobs/%s/tasks", invalidJobID), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, validTaskName2)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 404 because the job id was invalid", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "Failed to find job that has the specified id")
			})
		})
	})

	Convey("Given an API that can create valid search reindex tasks and store their details in a Data Store", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/jobs/%s/tasks", validJobID1), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, emptyTaskName)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 400 because the task name is empty", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "invalid request body")
			})
		})
	})

	Convey("Given an API that can create valid search reindex tasks and store their details in a Data Store", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		httpClient := dpHTTP.NewClient()
		apiInstance := api.Setup(mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{}, taskNames, cfg, httpClient, &apiMock.IndexerMock{}, &apiMock.ReindexRequestedProducerMock{})

		Convey("When the tasks endpoint is called to create and store a new reindex task", func() {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:25700/jobs/%s/tasks", validJobID1), bytes.NewBufferString(
				fmt.Sprintf(createTaskPayloadFmt, invalidTaskName)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", validServiceAuthToken)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 400 because the task name is invalid", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "invalid request body")
			})
		})
	})
}

// Create Task Payload
var createTaskPayloadFmt = `{
	"task_name": "%s",
	"number_of_documents": 5
}`

func ExpectedTask(jobID string, lastUpdated time.Time, numberOfDocuments int, taskName string) (models.Task, error) {
	cfg, err := config.Get()
	if err != nil {
		return models.Task{}, fmt.Errorf("unable to retrieve service configuration: %w", err)
	}
	urlBuilder := url.NewBuilder("http://" + cfg.BindAddr)
	job := urlBuilder.BuildJobURL(jobID)
	self := urlBuilder.BuildJobTaskURL(jobID, taskName)
	return models.Task{
		JobID:       jobID,
		LastUpdated: lastUpdated,
		Links: &models.TaskLinks{
			Self: self,
			Job:  job,
		},
		NumberOfDocuments: numberOfDocuments,
		TaskName:          taskName,
	}, nil
}
