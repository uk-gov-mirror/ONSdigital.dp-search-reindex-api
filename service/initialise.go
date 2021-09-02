package service

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpHTTP "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"net/http"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	MongoDB     bool
	HealthCheck bool
	Init        Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		MongoDB:     false,
		HealthCheck: false,
		Init:        initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetMongoDB creates a mongoDB client and sets the Mongo flag to true
func (e *ExternalServiceList) GetMongoDB(ctx context.Context, cfg *config.Config) (MongoJobStorer, error) {
	mongoDB, err := e.Init.DoGetMongoDB(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.MongoDB = true
	return mongoDB, nil
}

// GetHealthClient returns a healthClient for the provided URL
func (e *ExternalServiceList) GetHealthClient(name, url string) *health.Client {
	return e.Init.DoGetHealthClient(name, url)
}

// GetHealthCheck creates a healthCheck with versionInfo and sets teh HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dpHTTP.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetHealthClient creates a new Health Client for the provided name and url
func (e *Init) DoGetHealthClient(name, url string) *health.Client {
	return health.NewClient(name, url)
}

// DoGetHealthCheck creates a healthCheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

// DoGetMongoDB returns a MongoDB
func (e *Init) DoGetMongoDB(ctx context.Context, cfg *config.Config) (MongoJobStorer, error) {
	mongodb := &mongo.JobStore{
		JobsCollection:  cfg.MongoConfig.JobsCollection,
		LocksCollection: cfg.MongoConfig.LocksCollection,
		TasksCollection: cfg.MongoConfig.TasksCollection,
		Database:        cfg.MongoConfig.Database,
		URI:             cfg.MongoConfig.BindAddr,
	}
	if err := mongodb.Init(ctx, cfg); err != nil {
		return nil, err
	}
	return mongodb, nil
}
