package api

import (
	"context"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

//JobStoreAPI provides a struct to wrap the api around
type JobStoreAPI struct {
	Router  *mux.Router
	jobStore JobStorer
}

//Setup function sets up the api and returns an api
func Setup(ctx context.Context, router *mux.Router, jobStorer JobStorer) *JobStoreAPI {
	api := &JobStoreAPI{
		Router:  router,
		jobStore: jobStorer,
	}

	router.HandleFunc("/jobs", api.CreateJobHandler(ctx)).Methods("POST")
	router.HandleFunc("/jobs/{id}", api.GetJobHandler(ctx)).Methods("GET")
	router.HandleFunc("/jobs", api.GetJobsHandler)
	return api
}

//Close is called during graceful shutdown to give the API an opportunity to perform any required disposal task
func (*JobStoreAPI) Close(ctx context.Context) error {
	log.Event(ctx, "graceful shutdown of api complete", log.INFO)
	return nil
}
