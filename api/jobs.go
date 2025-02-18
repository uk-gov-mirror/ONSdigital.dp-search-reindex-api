package api

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

//NewID generates a random uuid and returns it as a string.
var NewID = func() string {
	return uuid.NewV4().String()
}

// CreateJobHandler returns a function that generates a new Job resource containing default values in its fields.
func (api *JobStoreAPI) CreateJobHandler(ctx context.Context) http.HandlerFunc {
	log.Event(ctx, "Entering CreateJobHandler function, which generates a new Job resource.", log.INFO)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		id := NewID()

		//Create job in job store
		newJob, err := api.jobStore.CreateJob(ctx, id)
		if err != nil {
			log.Event(ctx, "creating and storing job failed", log.Error(err), log.ERROR)
			http.Error(w, "Failed to create and store job", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonResponse, err := json.Marshal(newJob)
		if err != nil {
			log.Event(ctx, "marshalling response failed", log.Error(err), log.ERROR)
			http.Error(w, "Failed to marshall json response", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(jsonResponse)
		if err != nil {
			log.Event(ctx, "writing response failed", log.Error(err), log.ERROR)
			http.Error(w, "Failed to write http response", http.StatusInternalServerError)
			return
		}
	}
}

//GetJobHandler returns a function that gets an existing Job resource, from the Job Store, that's associated with the id passed in.
func (api *JobStoreAPI) GetJobHandler(ctx context.Context) http.HandlerFunc {
	log.Event(ctx, "Entering GetJobHandler function, which returns an existing Job resource associated with the supplied id.", log.INFO)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		vars := mux.Vars(req)
		id := vars["id"]
		logData := log.Data{"job_id": id}

		// get job from jobStore by id
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
			http.Error(w, "Failed to marshall json response", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonResponse)
		if err != nil {
			log.Event(ctx, "writing response failed", log.Error(err), logData, log.ERROR)
			http.Error(w, "Failed to write http response", http.StatusInternalServerError)
			return
		}
	}
}
