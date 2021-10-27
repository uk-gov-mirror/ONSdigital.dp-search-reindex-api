package service

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-reindex-api/api"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"net/http"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthCheck.go -pkg mock . HealthChecker
//go:generate moq -out mock/mongo_data_storer.go -pkg mock . MongoDataStorer

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetHealthClient(name, url string) *health.Client
	DoGetMongoDB(ctx context.Context, cfg *config.Config) (MongoDataStorer, error)
	DoGetAuthorisationHandlers(ctx context.Context, cfg *config.Config) api.AuthHandler
}

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from HealthCheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

// MongoDataStorer defines the required methods for a Mongo DB Jobstore
type MongoDataStorer interface {
	api.DataStorer
	Close(ctx context.Context) error
	Checker(ctx context.Context, state *healthcheck.CheckState) (err error)
}
