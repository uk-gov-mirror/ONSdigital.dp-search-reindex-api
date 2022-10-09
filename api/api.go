package api

import (
	"context"

	"github.com/ONSdigital/dp-authorisation/auth"
	dpHTTP "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

var update = auth.Permissions{Update: true}

const v1 = "v1"

// API provides a struct to wrap the api around
type API struct {
	Router      *mux.Router
	cfg         *config.Config
	dataStore   DataStorer
	httpClient  dpHTTP.Clienter
	permissions AuthHandler
	producer    ReindexRequestedProducer
	reindex     Indexer
	taskNames   map[string]bool
}

// Setup function sets up the api and returns an api
func Setup(router *mux.Router,
	dataStore DataStorer,
	permissions AuthHandler,
	taskNames map[string]bool,
	cfg *config.Config,
	httpClient dpHTTP.Clienter,
	reindex Indexer,
	producer ReindexRequestedProducer) *API {
	api := &API{
		Router:      router,
		cfg:         cfg,
		dataStore:   dataStore,
		permissions: permissions,
		taskNames:   taskNames,
		httpClient:  httpClient,
		reindex:     reindex,
		producer:    producer,
	}

	// These routes should always use the latest API version
	api.latestAPIRoutes(permissions)

	v1 := router.PathPrefix("/{version:v1}").Subrouter()
	v1.HandleFunc("/search-reindex-jobs", api.GetJobsHandler).Methods("GET")
	v1.HandleFunc("/search-reindex-jobs", api.CreateJobHandler).Methods("POST")
	v1.HandleFunc("/search-reindex-jobs/{job_id}/tasks/{task_name}/number-of-documents/{count}", permissions.Require(update, api.PutTaskNumOfDocsHandler)).Methods("PUT")
	v1.HandleFunc("/search-reindex-jobs/{id}", api.GetJobHandler).Methods("GET")
	v1.HandleFunc("/search-reindex-jobs/{id}", permissions.Require(update, api.PatchJobStatusHandler)).Methods("PATCH")
	v1.HandleFunc("/search-reindex-jobs/{id}/number-of-tasks/{count}", permissions.Require(update, api.PutNumTasksHandler)).Methods("PUT")
	v1.HandleFunc("/search-reindex-jobs/{id}/tasks", api.GetTasksHandler).Methods("GET")
	v1.HandleFunc("/search-reindex-jobs/{id}/tasks", permissions.Require(update, api.CreateTaskHandler)).Methods("POST")
	v1.HandleFunc("/search-reindex-jobs/{id}/tasks/{task_name}", api.GetTaskHandler).Methods("GET")

	return api
}

// Close is called during graceful shutdown to give the API an opportunity to perform any required disposal task
func (*API) Close(ctx context.Context) error {
	log.Info(ctx, "graceful shutdown of api complete")
	return nil
}

func (api *API) latestAPIRoutes(permissions AuthHandler) {
	switch api.cfg.LatestVersion {
	case "v1":
		api.v1RoutesAsLatest(permissions)
	default:
		// Set default in case app configuration is set to non-existant version
		api.v1RoutesAsLatest(permissions)
	}
}

func (api *API) v1RoutesAsLatest(permissions AuthHandler) {
	api.Router.HandleFunc("/search-reindex-jobs", api.GetJobsHandler).Methods("GET")
	api.Router.HandleFunc("/search-reindex-jobs", api.CreateJobHandler).Methods("POST")
	api.Router.HandleFunc("/search-reindex-jobs/{id}", api.GetJobHandler).Methods("GET")
	api.Router.HandleFunc("/search-reindex-jobs/{id}", permissions.Require(update, api.PatchJobStatusHandler)).Methods("PATCH")
	api.Router.HandleFunc("/search-reindex-jobs/{job_id}/tasks/{task_name}/number-of-documents/{count}", permissions.Require(update, api.PutTaskNumOfDocsHandler)).Methods("PUT")
	api.Router.HandleFunc("/search-reindex-jobs/{id}/number-of-tasks/{count}", permissions.Require(update, api.PutNumTasksHandler)).Methods("PUT")
	api.Router.HandleFunc("/search-reindex-jobs/{id}/tasks", api.GetTasksHandler).Methods("GET")
	api.Router.HandleFunc("/search-reindex-jobs/{id}/tasks", permissions.Require(update, api.CreateTaskHandler)).Methods("POST")
	api.Router.HandleFunc("/search-reindex-jobs/{id}/tasks/{task_name}", api.GetTaskHandler).Methods("GET")
}
