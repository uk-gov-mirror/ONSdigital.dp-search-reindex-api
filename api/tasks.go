package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// CreateTaskHandler returns a function that generates a new TaskName resource containing default values in its fields.
func (api *JobStoreAPI) CreateTaskHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		vars := mux.Vars(req)
		jobID := vars["id"]

		// Unmarshal task to create and validate it
		taskToCreate := &models.TaskToCreate{}
		if err := ReadJSONBody(ctx, req.Body, taskToCreate); err != nil {
			log.Error(ctx, "reading request body failed", err)
			http.Error(w, serverErrorMessage, http.StatusBadRequest)
			return
		}

		newTask, err := api.jobStore.CreateTask(ctx, jobID, taskToCreate.TaskName, taskToCreate.NumberOfDocuments)
		if err != nil {
			log.Error(ctx, "creating and storing a task failed", err, log.Data{"job id": jobID})
			if err == mongo.ErrJobNotFound {
				http.Error(w, "Failed to find job that has the specified id", http.StatusNotFound)
			} else {
				http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			}
			return
		}

		jsonResponse, err := json.Marshal(newTask)
		if err != nil {
			log.Error(ctx, "marshalling response failed", err)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(jsonResponse)
		if err != nil {
			log.Error(ctx, "writing response failed", err)
			http.Error(w, serverErrorMessage, http.StatusInternalServerError)
			return
		}
	}
}
