package api

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

// JobStoreAPI provides a struct to wrap the api around
type JobStoreAPI struct {
	Router   *mux.Router
	jobStore JobStorer
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, router *mux.Router, jobStorer JobStorer) *JobStoreAPI {
	api := &JobStoreAPI{
		Router:   router,
		jobStore: jobStorer,
	}

	router.HandleFunc("/jobs", api.CreateJobHandler(ctx)).Methods("POST")
	router.HandleFunc("/jobs/{id}", api.GetJobHandler(ctx)).Methods("GET")
	router.HandleFunc("/jobs", api.GetJobsHandler)
	router.HandleFunc("/jobs/{id}/number_of_tasks/{count}", api.PutNumTasksHandler(ctx)).Methods("PUT")
	router.HandleFunc("/jobs/{id}/tasks", api.CreateTaskHandler(ctx)).Methods("POST")
	return api
}

// Close is called during graceful shutdown to give the API an opportunity to perform any required disposal task
func (*JobStoreAPI) Close(ctx context.Context) error {
	log.Event(ctx, "graceful shutdown of api complete", log.INFO)
	return nil
}

// ReadJSONBody reads the bytes from the provided body, and marshals it to the provided model interface.
func ReadJSONBody(ctx context.Context, body io.ReadCloser, v interface{}) error {
	defer body.Close()

	// Get Body bytes
	payload, err := ioutil.ReadAll(body)
	if err != nil {
		return apierrors.ErrUnableToReadMessage
	}

	// Unmarshal body bytes to model
	if err := json.Unmarshal(payload, v); err != nil {
		return apierrors.ErrUnableToParseJSON
	}

	return nil
}
