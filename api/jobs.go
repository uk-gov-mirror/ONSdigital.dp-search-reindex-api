package api

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"

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

// CreateJobHandler generates a new Job resource containing default values in its fields.
func (api *API) CreateJobHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	id := NewID()

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

func (api *API) setUpPagination(offsetParam string, limitParam string) (int, int, error) {
	paginator := pagination.NewPaginator(api.cfg.DefaultLimit, api.cfg.DefaultOffset, api.cfg.DefaultMaxLimit)
	return paginator.ValidatePaginationParameters(offsetParam, limitParam)
}

// unlockJob unlocks the provided job lockID and logs any error with WARN state
func (api *API) unlockJob(ctx context.Context, lockID string) {
	log.Info(ctx, "Entering unlockJob function, which unlocks the provided job lockID.")
	if err := api.dataStore.UnlockJob(lockID); err != nil {
		log.Warn(ctx, "error unlocking lockID for a job resource", log.Data{"lockID": lockID})
	}
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
