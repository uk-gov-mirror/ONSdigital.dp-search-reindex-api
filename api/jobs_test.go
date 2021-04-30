package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/api/mock"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/store"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

// Constants for testing
const (
	testJobID1 = "UUID1"
	testJobID2 = "UUID2"
	emptyJobID = ""
)

var ctx = context.Background()

func TestCreateJobHandlerWithValidID(t *testing.T) {
	t.Parallel()
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

func TestGetJobHandler(t *testing.T) {
	t.Parallel()
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

			Convey("Then job resource was not found returning a status code of 404", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "Failed to find job in job store")
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
			})
		})
	})
}

func TestGetJobsHandler(t *testing.T) {
	t.Parallel()
	Convey("Given a Search Reindex Job API that returns a list of jobs", t, func() {
		jobStoreMock := &mock.JobStoreMock{
			GetJobsFunc: func(ctx context.Context) (models.Jobs, error) {
				jobs := models.Jobs{}
				jobsList := make([]models.Job, 2)

				jobsList[0] = models.NewJob(testJobID1)
				jobsList[1] = models.NewJob(testJobID2)

				jobs.Job_List = jobsList

				return jobs, nil
			},
		}

		api := Setup(ctx, mux.NewRouter(), jobStoreMock)

		Convey("When a request is made to get a list of all the jobs that exist in the Job Store", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then a list of jobs is returned with status code 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
				payload, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				jobsReturned := models.Jobs{}
				err = json.Unmarshal(payload, &jobsReturned)
				zeroTime := time.Time{}.UTC()
				expectedJob1 := ExpectedJob(testJobID1, zeroTime, 0, zeroTime, zeroTime, zeroTime, "Default Search Index Name", "created", 0, 0)
				expectedJob2 := ExpectedJob(testJobID2, zeroTime, 0, zeroTime, zeroTime, zeroTime, "Default Search Index Name", "created", 0, 0)

				Convey("And the returned list should contain expected jobs", func() {
					returnedJobList := jobsReturned.Job_List
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
	})
}

func TestGetJobsHandlerWithEmptyJobStore(t *testing.T) {
	t.Parallel()
	Convey("Given a Search Reindex Job API that returns an empty list of jobs", t, func() {

		jobStoreMock := &mock.JobStoreMock{
			GetJobsFunc: func(ctx context.Context) (models.Jobs, error) {
				jobs := models.Jobs{}

				return jobs, nil
			},
		}

		api := Setup(ctx, mux.NewRouter(), jobStoreMock)

		Convey("When a request is made to get a list of all the jobs that exist in the Job Store", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then a jobs resource is returned with status code 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
				payload, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				jobsReturned := models.Jobs{}
				err = json.Unmarshal(payload, &jobsReturned)

				Convey("And the returned jobs list should be empty", func() {
					So(jobsReturned.Job_List, ShouldHaveLength, 0)
				})
			})
		})
	})
}

func TestGetJobsHandlerWithInternalServerError(t *testing.T) {
	t.Parallel()
	Convey("Given a Search Reindex Job API that generates an internal server error", t, func() {
		jobStoreMock := &mock.JobStoreMock{
			GetJobsFunc: func(ctx context.Context) (models.Jobs, error) {
				jobs := models.Jobs{}

				return jobs, errors.New("something went wrong in the server")
			},
		}

		api := Setup(ctx, mux.NewRouter(), jobStoreMock)

		Convey("When a request is made to get a list of all the jobs that exist in the Job Store", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then a jobs resource is returned with status code 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "Failed to get list of jobs from job store")
			})
		})
	})
}

//ExpectedJob returns a Job resource that can be used to define and test expected values within it.
func ExpectedJob(id string,
	last_updated time.Time,
	number_of_tasks int,
	reindex_completed time.Time,
	reindex_failed time.Time,
	reindex_started time.Time,
	search_index_name string,
	state string,
	total_search_documents int,
	total_inserted_search_documents int) models.Job {
	return models.Job{
		ID:          id,
		LastUpdated: last_updated,
		Links: &models.JobLinks{
			Tasks: "http://localhost:12150/jobs/" + id + "/tasks",
			Self:  "http://localhost:12150/jobs/" + id,
		},
		NumberOfTasks:                number_of_tasks,
		ReindexCompleted:             reindex_completed,
		ReindexFailed:                reindex_failed,
		ReindexStarted:               reindex_started,
		SearchIndexName:              search_index_name,
		State:                        state,
		TotalSearchDocuments:         total_search_documents,
		TotalInsertedSearchDocuments: total_inserted_search_documents,
	}
}
