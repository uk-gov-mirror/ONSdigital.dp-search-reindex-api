package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpkafka "github.com/ONSdigital/dp-kafka/v2"
	dpHTTP "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-reindex-api/api"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/event"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/dp-search-reindex-api/schema"
	"github.com/ONSdigital/log.go/v2/log"
)

// KafkaTLSProtocolFlag informs service to use TLS protocol for kafka
const KafkaTLSProtocolFlag = "TLS"

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	MongoDB       bool
	HealthCheck   bool
	Auth          bool
	KafkaProducer bool
	Init          Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		MongoDB:       false,
		HealthCheck:   false,
		Auth:          false,
		KafkaProducer: false,
		Init:          initialiser,
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
func (e *ExternalServiceList) GetMongoDB(ctx context.Context, cfg *config.Config) (MongoDataStorer, error) {
	mongoDB, err := e.Init.DoGetMongoDB(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.MongoDB = true
	return mongoDB, nil
}

// GetAuthorisationHandlers creates an AuthHandler client and sets the Auth flag to true
func (e *ExternalServiceList) GetAuthorisationHandlers(ctx context.Context, cfg *config.Config) api.AuthHandler {
	e.Auth = true
	return e.Init.DoGetAuthorisationHandlers(ctx, cfg)
}

// GetHealthClient returns a healthClient for the provided URL
func (e *ExternalServiceList) GetHealthClient(name, url string) *health.Client {
	return e.Init.DoGetHealthClient(name, url)
}

// GetHealthCheck creates a healthCheck with versionInfo and sets the HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// GetKafkaProducer creates a Kafka producer and sets the producder flag to true
func (e *ExternalServiceList) GetKafkaProducer(ctx context.Context, cfg *config.Config) (KafkaProducer, error) {
	producer, err := e.Init.DoGetKafkaProducer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.KafkaProducer = true
	return producer, nil
}

func (e *Init) DoGetKafkaProducer(ctx context.Context, cfg *config.Config) (KafkaProducer, error) {
	pChannels := dpkafka.CreateProducerChannels()
	pConfig := &dpkafka.ProducerConfig{
		KafkaVersion: &cfg.KafkaConfig.Version,
	}

	if cfg.KafkaConfig.SecProtocol == KafkaTLSProtocolFlag {
		pConfig.SecurityConfig = dpkafka.GetSecurityConfig(
			cfg.KafkaConfig.SecCACerts,
			cfg.KafkaConfig.SecClientCert,
			cfg.KafkaConfig.SecClientKey,
			cfg.KafkaConfig.SecSkipVerify,
		)
	}
	producer, err := dpkafka.NewProducer(ctx, cfg.KafkaConfig.Brokers, cfg.KafkaConfig.ReindexRequestedTopic, pChannels, pConfig)
	if err != nil {
		log.Fatal(ctx, "kafka producer returned an error", err, log.Data{"topic": cfg.KafkaConfig.ReindexRequestedTopic})
		return nil, err
	}

	reindexRequestedProducer := &event.ReindexRequestedProducer{
		Marshaller: schema.ReindexRequestedEvent,
		Producer:   producer,
	}

	// Kafka error logging go-routine
	producer.Channels().LogErrors(ctx, "kafka producer channels error")

	return reindexRequestedProducer, nil
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
func (e *Init) DoGetMongoDB(ctx context.Context, cfg *config.Config) (MongoDataStorer, error) {
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

func (e *Init) DoGetAuthorisationHandlers(ctx context.Context, cfg *config.Config) api.AuthHandler {
	authClient := auth.NewPermissionsClient(dpHTTP.NewClient())
	authVerifier := auth.DefaultPermissionsVerifier()

	// for checking caller permissions when we only have a user/service token
	permissions := auth.NewHandler(
		auth.NewPermissionsRequestBuilder(cfg.ZebedeeURL),
		authClient,
		authVerifier,
	)

	return permissions
}
