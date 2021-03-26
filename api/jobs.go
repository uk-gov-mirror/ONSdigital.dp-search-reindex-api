package api

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/log.go/log"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

// CreateJobHandler returns a function that generates a new Job resource containing default values in its fields.
func (api *JobStorerAPI)CreateJobHandler(ctx context.Context) http.HandlerFunc {
	log.Event(ctx, "Entering CreateJobHandler function, which generates a new Job resource.", log.INFO)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		id := NewID()

		//Create job in job store
		newJob, err := api.jobStore.CreateJob(ctx, id)
		if err != nil {
			log.Event(ctx, "creating and storing job failed", log.Error(err), log.ERROR)
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

//NewID generates a random uuid and returns it as a string.
var NewID = func() string {
	return uuid.NewV4().String()
}