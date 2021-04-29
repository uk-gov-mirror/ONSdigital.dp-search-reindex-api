package api

import (
	"context"

	store "github.com/ONSdigital/dp-search-reindex-api/store"
	"github.com/gorilla/mux"
)

//JobStoreAPI provides a struct to wrap the api around
type JobStoreAPI struct {
	Router   *mux.Router
	jobStore store.JobStore
}

//Setup function sets up the api and returns an api
func Setup(ctx context.Context, router *mux.Router, jobStore store.JobStore) *JobStoreAPI {
	api := &JobStoreAPI{
		Router:   router,
		jobStore: jobStore,
	}

	router.HandleFunc("/jobs", api.CreateJobHandler(ctx)).Methods("POST")
	router.HandleFunc("/jobs/{id}", api.GetJobHandler(ctx)).Methods("GET")
	router.HandleFunc("/jobs", api.GetJobsHandler(ctx)).Methods("GET")
	return api
}
