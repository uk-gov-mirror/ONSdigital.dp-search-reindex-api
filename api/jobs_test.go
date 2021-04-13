package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-search-reindex-api/api/mock"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/store"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Constants for testing
const (
	testJobID1 = "UUID1"
	testJobID2 = "UUID2"
	emptyJobID = ""
)

var ctx = context.Background()

func TestCreateJobHandlerWithValidID(t *testing.T) {

	NewID = func() string { return testJobID1 }

	Convey("Given a Search Reindex Job API that can create valid search reindex jobs and store their details in a map", t, func() {

		api := Setup(ctx, mux.NewRouter(), &store.DataStore{})
		createJobHandler := api.CreateJobHandler(ctx)

		Convey("When a new reindex job is created and stored", func() {
			req := httptest.NewRequest("POST", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			createJobHandler.ServeHTTP(resp, req)

			Convey("Then the newly created search reindex job is returned with status code 201", func() {
				So(resp.Code, ShouldEqual, http.StatusCreated)
				payload, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				newJob := models.Job{}
				err = json.Unmarshal(payload, &newJob)
				So(err, ShouldBeNil)
				expectedJob := models.NewJob(testJobID1)

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
}

func TestGetJobHandlerWithValidID(t *testing.T) {

	Convey("Given a Search Reindex Job API that returns specific jobs using their id as a key", t, func() {

		jobStoreMock := &mock.JobStoreMock{
			GetJobFunc: func(ctx context.Context, id string) (models.Job, error) {
				switch id {
				case testJobID2:
					return models.NewJob(testJobID2), nil
				default:
					return models.Job{}, errors.New("the job store does not contain the job id entered")
				}
			},
		}

		api := Setup(ctx, mux.NewRouter(), jobStoreMock)

		Convey("When a request is made to get a specific job that exists in the Job Store", func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost:25700/jobs/%s", testJobID2), nil)
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then the relevant search reindex job is returned with status code 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
				payload, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				jobReturned := models.Job{}
				err = json.Unmarshal(payload, &jobReturned)
				expectedJob := models.NewJob(testJobID2)

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

		Convey("When a request is made to get a specific job that does not exist in the Job Store", func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost:25700/jobs/%s", testJobID1), nil)
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 404", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "Failed to find job in job store")
				payload, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				newJob := models.Job{}
				err = json.Unmarshal(payload, &newJob)
				So(newJob, ShouldResemble, models.Job{})
			})
		})
	})
}

func TestCreateJobHandlerWithInvalidID(t *testing.T) {

	NewID = func() string { return emptyJobID }

	Convey("Given a Search Reindex Job API that can create valid search reindex jobs and store their details in a map", t, func() {
		api := Setup(ctx, mux.NewRouter(), &store.DataStore{})
		createJobHandler := api.CreateJobHandler(ctx)

		Convey("When the jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest("POST", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			createJobHandler.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "Failed to create and store job")
				payload, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				newJob := models.Job{}
				err = json.Unmarshal(payload, &newJob)
				So(newJob, ShouldResemble, models.Job{})
			})
		})
	})
}
