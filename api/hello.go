package api

import (
	"context"
	"encoding/json"
	"net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/ONSdigital/dp-search-reindex-api/models"
)

const helloMessage = "Hello, World!"

type HelloResponse struct {
	Message string `json:"message,omitempty"`
}

// JobHandler returns function containing a simple hello world example of an api handler
func JobHandler(ctx context.Context) http.HandlerFunc {
	log.Event(ctx, "api contains example endpoint, remove hello.go as soon as possible", log.INFO)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		response := &models.Job{
			ID: "abc123",
			LastUpdated: "2021-03-15",
			Links: &models.JobLinks{
				Self: "string",
				Tasks: "string",
			},
			NumberOfTasks: 0,
	        ReindexCompleted: "2021-03-15",
			ReindexFailed: "2021-03-15",
			ReindexStarted:	"2021-03-15",
			SearchIndexName: "string",
			State: "CREATED",
			TotalSearchDocuments: 0,
			TotalInsertedSearchDocuments: 0,
		}

		w.Header().Set("Content-Type", "application/json")
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			log.Event(ctx, "marshalling response failed", log.Error(err), log.ERROR)
			http.Error(w, "Failed to marshall json response", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(jsonResponse)
		if err != nil {
			log.Event(ctx, "writing response failed", log.Error(err), log.ERROR)
			http.Error(w, "Failed to write http response", http.StatusInternalServerError)
			return
		}
	}
}
