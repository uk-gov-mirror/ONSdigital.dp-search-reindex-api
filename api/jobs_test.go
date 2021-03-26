package api

import (
	"context"
	"github.com/ONSdigital/dp-search-reindex-api/store"
	"github.com/gorilla/mux"
	"time"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"testing"
)

// Constants for testing
const (
	testJobID1           = "UUID1"
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
				So(newJob, ShouldResemble, createdJob())
			})
		})
	})
}

// API model corresponding to dbCreatedImage
func createdJob() models.Job {
	created_job := models.Job {
		ID: testJobID1,
		LastUpdated: time.Now().UTC(),
		Links: &models.JobLinks{
			Tasks: "http://localhost:12150/jobs/" + testJobID1 + "/tasks",
			Self: "http://localhost:12150/jobs/" + testJobID1,
		},
		NumberOfTasks: 0,
		ReindexCompleted: time.Time{}.UTC(),
		ReindexFailed: time.Time{}.UTC(),
		ReindexStarted:time.Time{}.UTC(),
		SearchIndexName: "Default Search Index Name",
		State: "created",
		TotalSearchDocuments: 0,
		TotalInsertedSearchDocuments: 0,
	}
	return created_job
}

// GetAPIWithMocks also used in other tests, so exported
//func GetAPIWithMocks() *JobStorerAPI {
//
//	return Setup(ctx, mux.NewRouter(), store.JobStorer{})
//}

//func TestHelloHandler(t *testing.T) {
//
//	Convey("Given a Hello handler ", t, func() {
//		helloHandler := HelloHandler(ctx)
//
//		Convey("when a good response is returned", func() {
//			req := httptest.NewRequest("GET", "http://localhost:8080/hello", nil)
//			resp := httptest.NewRecorder()
//
//			helloHandler.ServeHTTP(resp, req)
//
//			So(resp.Code, ShouldEqual, http.StatusOK)
//			So(resp.Body.String(), ShouldResemble, `{"message":"Hello, World!"}`)
//		})
//
//	})
//}
