package api

import (
	"context"
	"github.com/ONSdigital/dp-search-reindex-api/store"
	"github.com/gorilla/mux"
)

//JobstorerAPI provides a struct to wrap the api around
type JobStorerAPI struct {
	Router   *mux.Router
	jobStore store.JobStorer
}

//Setup function sets up the api and returns an api
func Setup(ctx context.Context, router *mux.Router, jobStore store.JobStorer) *JobStorerAPI {
	api := &JobStorerAPI{
		Router:   router,
		jobStore: jobStore,
	}

	router.HandleFunc("/jobs", api.CreateJobHandler(ctx)).Methods("POST")
	return api
}
