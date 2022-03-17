package api

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/dp-search-reindex-api/pagination"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

var (
	// NewID generates a random uuid and returns it as a string.
	NewID = func() string {
		return uuid.NewV4().String()
	}

	serverErrorMessage = "internal server error"
)

// CreateJobHandler generates a new Job resource and a new ElasticSearch index associated with it	.
func (api *API) CreateJobHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	id := NewID()

	log.Info(ctx, "creating new job resource in the data store")
	newJob, err := api.dataStore.CreateJob(ctx, id)
	if err != nil {
		log.Error(ctx, "creating and storing job failed", err)
		if err == mongo.ErrExistingJobInProgress {
			http.Error(w, "existing reindex job in progress", http.StatusConflict)
		} else {
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		}
		return
	}
	if newJob == (models.Job{}) {
		log.Info(ctx, "an empty job resource was returned by the data store")
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	log.Info(ctx, "creating new index in ElasticSearch via the Search API")
	serviceAuthToken := "Bearer " + api.cfg.ServiceAuthToken
	searchAPISearchURL := api.cfg.SearchAPIURL + "/search"
	reindexResponse, errCreateIndex := api.reindex.CreateIndex(ctx, serviceAuthToken, searchAPISearchURL, api.httpClient)
	if errCreateIndex != nil {
		log.Error(ctx, "error occurred when connecting to Search API", errCreateIndex)
		if !updateJobStateToFailed(ctx, w, &newJob, api) {
			return
		}
	} else if reindexResponse.StatusCode != 201 {
		log.Info(ctx, "unexpected status returned by the search api", log.Data{"status returned by search api": reindexResponse.Status})
		if !updateJobStateToFailed(ctx, w, &newJob, api) {
			return
		}
	} else {
		newJob, err = api.updateSearchIndexName(ctx, reindexResponse, newJob, id)
		if err != nil {
			log.Error(ctx, "error occurred in updateSearchIndexName function", err)
			if !updateJobStateToFailed(ctx, w, &newJob, api) {
				return
			}
		} else {
			// As the index name was updated successfully we can send a reindex-requested event
			traceID := request.NewRequestID(16)
			reindexReqEvent := models.ReindexRequested{
				JobID:       newJob.ID,
				SearchIndex: newJob.SearchIndexName,
				TraceID:     traceID,
			}

			log.Info(ctx, "sending reindex-requested event", log.Data{"reindexRequestedEvent": reindexReqEvent})

			if err = api.producer.ProduceReindexRequested(ctx, reindexReqEvent); err != nil {
				log.Error(ctx, "error while attempting to send reindex-requested event to producer", err)
				http.Error(w, serverErrorMessage, http.StatusInternalServerError)
				return
			}
			log.Info(ctx, "reindex request has been processed")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(newJob)
	if err != nil {
		log.Error(ctx, "marshalling response failed", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Error(ctx, "writing response failed", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

// updateJobStateToFailed returns true if the job state was successfully updated to failed
func updateJobStateToFailed(ctx context.Context, w http.ResponseWriter, newJob *models.Job, api *API) bool {
	newJob.State = models.JobStateFailed
	log.Info(ctx, "updating job state to failed", log.Data{"job id": newJob.ID})
	setStateErr := api.dataStore.UpdateJobState(models.JobStateFailed, newJob.ID)
	if setStateErr != nil {
		log.Error(ctx, "setting state to failed has failed", setStateErr)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return false
	}
	return true
}

// GetJobHandler returns a function that gets an existing Job resource, from the Job Store, that's associated with the id passed in.
func (api *API) GetJobHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	id := vars["id"]
	logData := log.Data{"job_id": id}

	lockID, err := api.dataStore.AcquireJobLock(ctx, id)
	if err != nil {
		log.Error(ctx, "acquiring lock for job ID failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
	defer api.unlockJob(ctx, lockID)

	job, err := api.dataStore.GetJob(req.Context(), id)
	if err != nil {
		log.Error(ctx, "getting job failed", err, logData)
		if err == mongo.ErrJobNotFound {
			http.Error(w, "Failed to find job in job store", http.StatusNotFound)
		} else {
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(job)
	if err != nil {
		log.Error(ctx, "marshalling response failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Error(ctx, "writing response failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

// GetJobsHandler gets a list of existing Job resources, from the Job Store, sorted by their values of
// last_updated time (ascending).
func (api *API) GetJobsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log.Info(ctx, "Entering handler function, which calls GetJobs and returns a list of existing Job resources held in the JobStore.")
	offsetParam := req.URL.Query().Get("offset")
	limitParam := req.URL.Query().Get("limit")

	offset, limit, err := api.setUpPagination(offsetParam, limitParam)
	if err != nil {
		log.Error(ctx, "pagination validation failed", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jobs, err := api.dataStore.GetJobs(ctx, offset, limit)
	if err != nil {
		log.Error(ctx, "getting list of jobs failed", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(jobs)
	if err != nil {
		log.Error(ctx, "marshalling response failed", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Error(ctx, "writing response failed", err)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

func (api *API) setUpPagination(offsetParam, limitParam string) (offset, limit int, err error) {
	paginator := pagination.NewPaginator(api.cfg.DefaultLimit, api.cfg.DefaultOffset, api.cfg.DefaultMaxLimit)
	return paginator.ValidatePaginationParameters(offsetParam, limitParam)
}

// unlockJob unlocks the provided job lockID
func (api *API) unlockJob(ctx context.Context, lockID string) {
	api.dataStore.UnlockJob(lockID)
	log.Info(ctx, "job lockID has unlocked successfully")
}

// PutNumTasksHandler returns a function that updates the number_of_tasks in an existing Job resource, which is associated with the id passed in.
func (api *API) PutNumTasksHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	id := vars["id"]
	count := vars["count"]
	logData := log.Data{"id": id, "count": count}

	numTasks, err := strconv.Atoi(count)
	if err != nil {
		log.Error(ctx, "invalid path parameter - failed to convert count to integer", err, logData)
		http.Error(w, "invalid path parameter - failed to convert count to integer", http.StatusBadRequest)
		return
	}

	floatNumTasks := float64(numTasks)
	isNegative := math.Signbit(floatNumTasks)
	if isNegative {
		err = errors.New("the count is negative")
	}
	if err != nil {
		log.Error(ctx, "invalid path parameter - count should be a positive integer", err, logData)
		http.Error(w, "invalid path parameter - count should be a positive integer", http.StatusBadRequest)
		return
	}

	lockID, err := api.dataStore.AcquireJobLock(ctx, id)
	if err != nil {
		log.Error(ctx, "acquiring lock for job ID failed", err, logData)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
	defer api.unlockJob(ctx, lockID)

	err = api.dataStore.PutNumberOfTasks(req.Context(), id, numTasks)
	if err != nil {
		log.Error(ctx, "putting number of tasks failed", err, logData)
		if err == mongo.ErrJobNotFound {
			http.Error(w, "Failed to find job in job store", http.StatusNotFound)
		} else {
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}

// updateSearchIndexName calls the GetIndexNameFromResponse function, in the reindex package, to get the index name that was returned by the Search API.
// It then calls the UpdateIndexName function, in the mongo package, to update the search_index_name value in the relevant Job Resource in the data store.
func (api *API) updateSearchIndexName(ctx context.Context, reindexResponse *http.Response, newJob models.Job, id string) (models.Job, error) {
	defer closeResponseBody(ctx, reindexResponse)
	indexName, err := api.reindex.GetIndexNameFromResponse(ctx, reindexResponse.Body)
	if err != nil {
		log.Error(ctx, "failed to get index name from response", err)
		if newJob != (models.Job{}) {
			newJob.State = models.JobStateFailed
			log.Info(ctx, "updating job state to failed", log.Data{"job id": newJob.ID})
			setStateErr := api.dataStore.UpdateJobState(models.JobStateFailed, newJob.ID)
			if setStateErr != nil {
				log.Error(ctx, "setting state to failed has failed", setStateErr)
				return newJob, setStateErr
			}
		}
		return newJob, err
	}

	newJob.SearchIndexName = indexName
	log.Info(ctx, "updating search index name", log.Data{"job id": id, "indexName": indexName})
	err = api.dataStore.UpdateIndexName(indexName, id)

	return newJob, err
}
