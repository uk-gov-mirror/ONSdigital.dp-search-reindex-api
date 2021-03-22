package api

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/log"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"time"
)

// CreateJobHandler returns a function that generates a new Job resource containing default values in its fields.
func CreateJobHandler(ctx context.Context) http.HandlerFunc {
	log.Event(ctx, "Entering CreateJobHandler function, which generates a new Job resource.", log.INFO)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		var NewID = func() string {
			return uuid.NewV4().String()
		}
		id := NewID()

		response := &models.Job{
			ID: id,
			LastUpdated: time.Now().UTC(),
			Links: &models.JobLinks{
				Tasks: "http://localhost:12150/jobs/" + id + "/tasks",
				Self: "http://localhost:12150/jobs/" + id,
			},
			NumberOfTasks: 0,
	        ReindexCompleted: time.Time{}.UTC(),
			ReindexFailed: time.Time{}.UTC(),
			ReindexStarted:time.Time{}.UTC(),
			SearchIndexName: "Default Search Index Name",
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




