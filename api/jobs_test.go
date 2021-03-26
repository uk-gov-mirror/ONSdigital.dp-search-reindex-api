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
)

var ctx = context.Background()

func TestCreateJobHandler(t *testing.T) {

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
				createdJob := createdJob()
				So(newJob.ID, ShouldEqual, createdJob.ID)
				So(newJob.Links, ShouldResemble, createdJob.Links)
				So(newJob.NumberOfTasks, ShouldEqual, createdJob.NumberOfTasks)
				So(newJob.ReindexCompleted, ShouldEqual, createdJob.ReindexCompleted)
				So(newJob.ReindexFailed, ShouldEqual, createdJob.ReindexFailed)
				So(newJob.ReindexStarted, ShouldEqual, createdJob.ReindexStarted)
				So(newJob.SearchIndexName, ShouldEqual, createdJob.SearchIndexName)
				So(newJob.State, ShouldEqual, createdJob.State)
				So(newJob.TotalSearchDocuments, ShouldEqual, createdJob.TotalSearchDocuments)
				So(newJob.TotalInsertedSearchDocuments, ShouldEqual, createdJob.TotalInsertedSearchDocuments)
			})
		})
	})
}

// API model corresponding to dbCreatedImage
func createdJob() models.Job {
	created_job := models.Job{
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
	return created_job
}
