package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/api"
	apiMock "github.com/ONSdigital/dp-search-reindex-api/api/mock"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/dp-search-reindex-api/url"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

// Constants for testing
const (
	validJobID1            = "UUID1"
	validJobID2            = "UUID2"
	validJobID3            = "UUID3"
	notFoundJobID          = "UUID4"
	unLockableJobID        = "UUID5"
	emptyJobID             = ""
	expectedServerErrorMsg = "internal server error"
	validCount             = "3"
	countNotANumber        = "notANumber"
	countNegativeInt       = "-3"
)

var ctx = context.Background()

func TestCreateJobHandler(t *testing.T) {
	t.Parallel()

	dataStorerMock := &apiMock.DataStorerMock{
		CreateJobFunc: func(ctx context.Context, id string) (models.Job, error) {
			switch id {
			case validJobID1:
				return models.NewJob(id)
			case validJobID2:
				return models.Job{}, mongo.ErrExistingJobInProgress
			default:
				return models.Job{}, errors.New("an unexpected error occurred")
			}
		},
	}

	Convey("Given a Search Reindex Job API that can create valid search reindex jobs and store their details in a Job Store", t, func() {
		api.NewID = func() string { return validJobID1 }
		apiInstance := api.Setup(ctx, mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{})
		createJobHandler := apiInstance.CreateJobHandler(ctx)

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
				expectedJob, err := models.NewJob(validJobID1)
				So(err, ShouldBeNil)

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

	Convey("Given a Search Reindex Job API that can create valid search reindex jobs and store their details in a Job Store", t, func() {
		api.NewID = func() string { return validJobID2 }
		apiInstance := api.Setup(ctx, mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{})
		createJobHandler := apiInstance.CreateJobHandler(ctx)

		Convey("When the jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest("POST", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			createJobHandler.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 409 because an existing job is in progress", func() {
				So(resp.Code, ShouldEqual, http.StatusConflict)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "existing reindex job in progress")
			})
		})
	})

	Convey("Given a Search Reindex Job API that generates an empty job ID", t, func() {
		api.NewID = func() string { return emptyJobID }
		apiInstance := api.Setup(ctx, mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{})
		createJobHandler := apiInstance.CreateJobHandler(ctx)

		Convey("When the jobs endpoint is called to create and store a new reindex job", func() {
			req := httptest.NewRequest("POST", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			createJobHandler.ServeHTTP(resp, req)

			Convey("Then an empty search reindex job is returned with status code 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedServerErrorMsg)
			})
		})
	})
}

func TestGetJobHandler(t *testing.T) {
	t.Parallel()
	Convey("Given a Search Reindex Job API that returns specific jobs using their id as a key", t, func() {
		dataStorerMock := &apiMock.DataStorerMock{
			GetJobFunc: func(ctx context.Context, id string) (models.Job, error) {
				switch id {
				case validJobID2:
					return models.NewJob(validJobID2)
				case notFoundJobID:
					return models.Job{}, mongo.ErrJobNotFound
				default:
					return models.Job{}, errors.New("an unexpected error occurred")
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
		}

		apiInstance := api.Setup(ctx, mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{})

		Convey("When a request is made to get a specific job that exists in the Job Store", func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost:25700/jobs/%s", validJobID2), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then the relevant search reindex job is returned with status code 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
				payload, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				jobReturned := models.Job{}
				err = json.Unmarshal(payload, &jobReturned)
				So(err, ShouldBeNil)
				expectedJob, err := models.NewJob(validJobID2)
				So(err, ShouldBeNil)

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
			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost:25700/jobs/%s", notFoundJobID), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then job resource was not found returning a status code of 404", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "Failed to find job in job store")
			})
		})

		Convey("When a request is made to get a specific job but the Job Store is unable to lock the id", func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost:25700/jobs/%s", unLockableJobID), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then an error with status code 500 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, expectedServerErrorMsg)
			})
		})

		Convey("When a request is made to get a specific job but an unexpected error occurs in the Job Store", func() {
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
	Convey("Given a Search Reindex Job API that returns a list of jobs", t, func() {
		dataStorerMock := &apiMock.DataStorerMock{
			GetJobsFunc: func(ctx context.Context, offsetParam string, limitParam string) (models.Jobs, error) {
				jobs := models.Jobs{}
				jobsList := make([]models.Job, 2)

				firstJob, err := models.NewJob(validJobID1)
				So(err, ShouldBeNil)
				jobsList[0] = firstJob

				secondJob, err := models.NewJob(validJobID2)
				So(err, ShouldBeNil)
				jobsList[1] = secondJob

				jobs.JobList = jobsList

				return jobs, nil
			},
		}

		apiInstance := api.Setup(ctx, mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{})

		Convey("When a request is made to get a list of all the jobs that exist in the Job Store", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a list of jobs is returned with status code 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
				payload, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				jobsReturned := models.Jobs{}
				err = json.Unmarshal(payload, &jobsReturned)
				So(err, ShouldBeNil)
				zeroTime := time.Time{}.UTC()
				expectedJob1, err := ExpectedJob(validJobID1, zeroTime, 0, zeroTime, zeroTime, zeroTime, "Default Search Index Name", "created", 0, 0)
				So(err, ShouldBeNil)
				expectedJob2, err := ExpectedJob(validJobID2, zeroTime, 0, zeroTime, zeroTime, zeroTime, "Default Search Index Name", "created", 0, 0)
				So(err, ShouldBeNil)

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
	})
}

func TestGetJobsHandlerWithEmptyJobStore(t *testing.T) {

	Convey("Given a Search Reindex Job API that returns an empty list of jobs", t, func() {

		dataStorerMock := &apiMock.DataStorerMock{
			GetJobsFunc: func(ctx context.Context, offsetParam string, limitParam string) (models.Jobs, error) {
				jobs := models.Jobs{}

				return jobs, nil
			},
		}

		apiInstance := api.Setup(ctx, mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{})

		Convey("When a request is made to get a list of all the jobs that exist in the jobs collection", func() {
			req := httptest.NewRequest("GET", "http://localhost:25700/jobs", nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a jobs resource is returned with status code 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
				payload, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
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
	Convey("Given a Search Reindex Job API that that failed to connect to the Job Store", t, func() {
		dataStorerMock := &apiMock.DataStorerMock{
			GetJobsFunc: func(ctx context.Context, offsetParam string, limitParam string) (models.Jobs, error) {
				jobs := models.Jobs{}

				return jobs, errors.New("something went wrong in the server")
			},
		}

		apiInstance := api.Setup(ctx, mux.NewRouter(), dataStorerMock, &apiMock.AuthHandlerMock{})

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

// ExpectedJob returns a Job resource that can be used to define and test expected values within it.
func ExpectedJob(id string,
	lastUpdated time.Time,
	numberOfTasks int,
	reindexCompleted time.Time,
	reindexFailed time.Time,
	reindexStarted time.Time,
	searchIndexName string,
	state string,
	totalSearchDocuments int,
	totalInsertedSearchDocuments int) (models.Job, error) {
	cfg, err := config.Get()
	if err != nil {
		return models.Job{}, fmt.Errorf("%s: %w", errors.New("unable to retrieve service configuration"), err)
	}
	urlBuilder := url.NewBuilder("http://" + cfg.BindAddr)
	self := urlBuilder.BuildJobURL(id)
	tasks := urlBuilder.BuildJobTasksURL(id)
	return models.Job{
		ID:          id,
		LastUpdated: lastUpdated,
		Links: &models.JobLinks{
			Tasks: tasks,
			Self:  self,
		},
		NumberOfTasks:                numberOfTasks,
		ReindexCompleted:             reindexCompleted,
		ReindexFailed:                reindexFailed,
		ReindexStarted:               reindexStarted,
		SearchIndexName:              searchIndexName,
		State:                        state,
		TotalSearchDocuments:         totalSearchDocuments,
		TotalInsertedSearchDocuments: totalInsertedSearchDocuments,
	}, err
}

func TestPutNumTasksHandler(t *testing.T) {
	t.Parallel()
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
		}

		apiInstance := api.Setup(ctx, mux.NewRouter(), jobStoreMock, &apiMock.AuthHandlerMock{})

		Convey("When a request is made to update the number of tasks of a specific job that exists in the Job Store", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/jobs/%s/number_of_tasks/%s", validJobID2, validCount), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then a status code 200 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When a request is made to update the number of tasks of a specific job that does not exist in the Job Store", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/jobs/%s/number_of_tasks/%s", notFoundJobID, validCount), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then job resource was not found returning a status code of 404", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "Failed to find job in job store")
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
				So(errMsg, ShouldEqual, "invalid path parameter - failed to convert count to integer")
			})
		})

		Convey("When a request is made to update the number of tasks but the path parameter given as the Count is a negative integer", func() {
			req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:25700/jobs/%s/number_of_tasks/%s", validJobID2, countNegativeInt), nil)
			resp := httptest.NewRecorder()

			apiInstance.Router.ServeHTTP(resp, req)

			Convey("Then it is a bad request returning a status code of 400", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				errMsg := strings.TrimSpace(resp.Body.String())
				So(errMsg, ShouldEqual, "invalid path parameter - count should be a positive integer")
			})
		})

		Convey("When a request is made to update the number of tasks but the Job Store is unable to lock the id", func() {
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
