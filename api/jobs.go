package api

import (
	"context"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"time"
	//"github.com/satori/go.uuid"
)

const helloMessage = "Hello, World!"

type HelloResponse struct {
	Message string `json:"message,omitempty"`
}

// JobHandler returns function containing a simple hello world example of an api handler
func CreateJobHandler(ctx context.Context) http.HandlerFunc {
	log.Event(ctx, "Entering CreateJobHandler function, which generates a new Job resource.", log.INFO)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		var id = func() string {
			return uuid.NewV4().String()
		}
		//id := NewID()

		response := &models.Job{
			ID: id(),
			LastUpdated: time.Date(2021, time.Month(2), 21, 1, 10, 30, 0, time.UTC),
			Links: &models.JobLinks{
				Tasks: "http://localhost:12150/jobs/abc123/tasks",
				Self: "http://localhost:12150/jobs/abc123",
			},
			NumberOfTasks: 0,
	        ReindexCompleted: time.Date(2021, time.Month(2), 21, 1, 10, 30, 0, time.UTC),
			ReindexFailed: time.Date(2021, time.Month(2), 21, 1, 10, 30, 0, time.UTC),
			ReindexStarted:	time.Date(2021, time.Month(2), 21, 1, 10, 30, 0, time.UTC),
			SearchIndexName: "string",
			State: "created",
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




