package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-search-reindex-api/apierrors"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

var update = auth.Permissions{Update: true}

// API provides a struct to wrap the api around
type API struct {
	Router      *mux.Router
	dataStore   DataStorer
	permissions AuthHandler
	taskNames   map[string]bool
	cfg         *config.Config
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, router *mux.Router, dataStore DataStorer, permissions AuthHandler, taskNames map[string]bool, cfg *config.Config) *API {
	api := &API{
		Router:      router,
		dataStore:   dataStore,
		permissions: permissions,
		taskNames:   taskNames,
		cfg:         cfg,
	}

	router.HandleFunc("/jobs", api.CreateJobHandler).Methods("POST")
	router.HandleFunc("/jobs/{id}", api.GetJobHandler).Methods("GET")
	router.HandleFunc("/jobs", api.GetJobsHandler)
	router.HandleFunc("/jobs/{id}/number_of_tasks/{count}", api.PutNumTasksHandler).Methods("PUT")
	taskHandler := permissions.Require(update, api.CreateTaskHandler)
	router.HandleFunc("/jobs/{id}/tasks", taskHandler).Methods("POST")
	router.HandleFunc("/jobs/{id}/tasks/{task_name}", api.GetTaskHandler).Methods("GET")
	router.HandleFunc("/jobs/{id}/tasks", api.GetTasksHandler).Methods("GET")
	return api
}

// Close is called during graceful shutdown to give the API an opportunity to perform any required disposal task
func (*API) Close(ctx context.Context) error {
	log.Info(ctx, "graceful shutdown of api complete")
	return nil
}

// ReadJSONBody reads the bytes from the provided body, and marshals it to the provided model interface.
func ReadJSONBody(ctx context.Context, body io.ReadCloser, v interface{}) error {
	defer body.Close()

	// Get Body bytes
	payload, err := ioutil.ReadAll(body)
	if err != nil {
		return fmt.Errorf("%s: %w", apierrors.ErrUnableToReadMessage, err)
	}

	// Unmarshal body bytes to model
	if err := json.Unmarshal(payload, v); err != nil {
		return fmt.Errorf("%s: %w", apierrors.ErrUnableToParseJSON, err)
	}

	return nil
}
