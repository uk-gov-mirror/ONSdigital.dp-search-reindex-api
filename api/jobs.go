package api

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// NewID generates a random uuid and returns it as a string.
var NewID = func() string {
	return uuid.NewV4().String()
}

var serverErrorMessage = "internal server error"

// CreateJobHandler returns a function that generates a new Job resource containing default values in its fields.
func (api *JobStoreAPI) CreateJobHandler(ctx context.Context) http.HandlerFunc {
	log.Event(ctx, "Creating handler function, which calls CreateJob and returns a new Job resource.", log.INFO)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		id := NewID()

		newJob, err := api.jobStore.CreateJob(ctx, id)
		if err != nil {
			log.Event(ctx, "creating and storing job failed", log.Error(err), log.ERROR)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonResponse, err := json.Marshal(newJob)
		if err != nil {
			log.Event(ctx, "marshalling response failed", log.Error(err), log.ERROR)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(jsonResponse)
		if err != nil {
			log.Event(ctx, "writing response failed", log.Error(err), log.ERROR)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}
	}
}

// GetJobHandler returns a function that gets an existing Job resource, from the Job Store, that's associated with the id passed in.
func (api *JobStoreAPI) GetJobHandler(ctx context.Context) http.HandlerFunc {
	log.Event(ctx, "Creating handler function, which calls GetJob and returns an existing Job resource associated with the supplied id.", log.INFO)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		vars := mux.Vars(req)
		id := vars["id"]
		logData := log.Data{"job_id": id}

		lockID, err := api.jobStore.AcquireJobLock(ctx, id)
		if err != nil {
			log.Event(ctx, "acquiring lock for job ID failed", log.Error(err), logData, log.ERROR)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}
		defer api.unlockJob(ctx, lockID)

		job, err := api.jobStore.GetJob(req.Context(), id)
		if err != nil {
			log.Event(ctx, "getting job failed", log.Error(err), logData, log.ERROR)
			http.Error(w, "Failed to find job in job store", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonResponse, err := json.Marshal(job)
		if err != nil {
			log.Event(ctx, "marshalling response failed", log.Error(err), logData, log.ERROR)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonResponse)
		if err != nil {
			log.Event(ctx, "writing response failed", log.Error(err), logData, log.ERROR)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}
	}
}

// GetJobsHandler gets a list of existing Job resources, from the Job Store, sorted by their values of
// last_updated time (ascending).
func (api *JobStoreAPI) GetJobsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log.Event(ctx, "Entering handler function, which calls GetJobs and returns a list of existing Job resources held in the JobStore.", log.INFO)

	jobs, err := api.jobStore.GetJobs(ctx)
	if err != nil {
		log.Event(ctx, "getting list of jobs failed", log.Error(err), log.ERROR)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(jobs)
	if err != nil {
		log.Event(ctx, "marshalling response failed", log.Error(err), log.ERROR)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Event(ctx, "writing response failed", log.Error(err), log.ERROR)
		http.Error(w, serverErrorMessage, http.StatusInternalServerError)
		return
	}
}

// unlockJob unlocks the provided job lockID and logs any error with WARN state
func (api *JobStoreAPI) unlockJob(ctx context.Context, lockID string) {
	log.Event(ctx, "Entering unlockJob function, which unlocks the provided job lockID.", log.INFO)
	if err := api.jobStore.UnlockJob(lockID); err != nil {
		log.Event(ctx, "error unlocking lockID for a job resource", log.WARN, log.Data{"lockID": lockID})
	}
}

// PutNumTasksHandler returns a function that updates the number_of_tasks in an existing Job resource, which is associated with the id passed in.
func (api *JobStoreAPI) PutNumTasksHandler(ctx context.Context) http.HandlerFunc {
	log.Event(ctx, "Creating handler function, which calls PutNumberOfTasks to update the number_of_tasks in the job with the supplied id.", log.INFO)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		vars := mux.Vars(req)
		id := vars["id"]
		count := vars["count"]
		logData := log.Data{"id": id, "count": count}

		numTasks, err := strconv.Atoi(count)
		if err != nil {
			log.Event(ctx, "invalid path parameter - failed to convert count to integer", log.Error(err), logData, log.ERROR)
			http.Error(w, "invalid path parameter - failed to convert count to integer", http.StatusBadRequest)
			return
		}

		floatNumTasks := float64(numTasks)
		isNegative := math.Signbit(floatNumTasks)
		if isNegative {
			err = errors.New("the count is negative")
		}
		if err != nil {
			log.Event(ctx, "invalid path parameter - count should be a positive integer", log.Error(err), logData, log.ERROR)
			http.Error(w, "invalid path parameter - count should be a positive integer", http.StatusBadRequest)
			return
		}

		lockID, err := api.jobStore.AcquireJobLock(ctx, id)
		if err != nil {
			log.Event(ctx, "acquiring lock for job ID failed", log.Error(err), logData, log.ERROR)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}
		defer api.unlockJob(ctx, lockID)

		err = api.jobStore.PutNumberOfTasks(req.Context(), id, numTasks)
		if err != nil {
			log.Event(ctx, "putting number of tasks failed", log.Error(err), logData, log.ERROR)
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
}
