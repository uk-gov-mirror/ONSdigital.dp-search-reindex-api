package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

// CreateTaskHandler returns a function that generates a new Task resource containing default values in its fields.
func (api *JobStoreAPI) CreateTaskHandler(ctx context.Context) http.HandlerFunc {
	log.Event(ctx, "Creating handler function, which calls CreateTask and returns a new Task resource.", log.INFO)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		vars := mux.Vars(req)
		jobID := vars["id"]
		logData := log.Data{"job_id": jobID}

		// Unmarshal task to create and validate it
		taskToCreate := &models.TaskToCreate{}
		if err := ReadJSONBody(ctx, req.Body, taskToCreate); err != nil {
			log.Event(ctx, "reading request body failed", log.Error(err), log.ERROR)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}

		newTask, err := api.jobStore.CreateTask(ctx, jobID, taskToCreate.NameOfApi, taskToCreate.NumberOfDocuments)
		if err != nil {
			log.Event(ctx, "creating and storing a task failed", log.Error(err), logData, log.ERROR)
			if err == mongo.ErrJobNotFound {
				http.Error(w, "Failed to find job that has the specified id", http.StatusNotFound)
			} else {
				http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonResponse, err := json.Marshal(newTask)
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
