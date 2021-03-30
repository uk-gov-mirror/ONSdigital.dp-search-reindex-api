package api

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/store"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Constants for testing
const (
	testJobID1 = "UUID1"
	emptyJobID = ""
)

var ctx = context.Background()

func TestCreateJobHandlerWithValidID(t *testing.T) {

	NewID = func() string { return testJobID1 }

	Convey("Given a Search Reindex Job API that can create valid search reindex jobs and store their details in a map", t, func() {

		api := Setup(ctx, mux.NewRouter(), store.JobStorer{})
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
				expectedJob := expectedJob()

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

func expectedJob() models.Job {
	return models.Job{
		ID:          testJobID1,
		LastUpdated: time.Now().UTC(),
		Links: &models.JobLinks{
			Tasks: "http://localhost:12150/jobs/" + testJobID1 + "/tasks",
			Self:  "http://localhost:12150/jobs/" + testJobID1,
		},
		NumberOfTasks:                0,
		ReindexCompleted:             time.Time{}.UTC(),
		ReindexFailed:                time.Time{}.UTC(),
		ReindexStarted:               time.Time{}.UTC(),
		SearchIndexName:              "Default Search Index Name",
		State:                        "created",
		TotalSearchDocuments:         0,
		TotalInsertedSearchDocuments: 0,
	}
}

func TestCreateJobHandlerWithInvalidID(t *testing.T) {

	NewID = func() string { return emptyJobID }

	Convey("Given a Search Reindex Job API that can create valid search reindex jobs and store their details in a map", t, func() {
		api := Setup(ctx, mux.NewRouter(), store.JobStorer{})
		createJobHandler := api.CreateJobHandler(ctx)

		Convey("When the jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest("POST", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			createJobHandler.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				payload, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				newJob := models.Job{}
				err = json.Unmarshal(payload, &newJob)
				So(newJob, ShouldResemble, models.Job{})
			})
		})
	})
}