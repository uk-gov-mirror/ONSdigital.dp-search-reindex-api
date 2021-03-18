package api

import (
	"context"
	"encoding/json"
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
	log.Event(ctx, "api contains example endpoint, remove hello.go as soon as possible", log.INFO)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		// id := NewID()

		// generate new image from request, mapping only allowed fields at creation time (model newImage in swagger spec)
		// the image is always created in 'created' state, and it is assigned a newly generated ID
		// newJob := models.Job{
		// 	ID: id,
		// }

		response := &models.Job{
			ID: "abc123",
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

// NewID returns a new UUID
//var NewID = func() string {
//	return uuid.NewV4().String()
//}


