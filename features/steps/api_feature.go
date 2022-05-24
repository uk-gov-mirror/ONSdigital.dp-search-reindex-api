// Package steps is used to define the steps that are used in the component test, which is written in godog (Go's version of cucumber).
package steps

import (
	"context"
	"fmt"
	"net/http"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	clientssitesearch "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	"github.com/ONSdigital/dp-authorisation/auth"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-component-test/utils"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	dpHTTP "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-reindex-api/api"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/event"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/dp-search-reindex-api/schema"
	"github.com/ONSdigital/dp-search-reindex-api/service"
	serviceMock "github.com/ONSdigital/dp-search-reindex-api/service/mock"
)

// collection names
const jobsCol = "jobs"
const tasksCol = "tasks"

var (
	// create map of valid task_name values for testing (in place of TaskNameValues in config)
	taskName1, taskName2, taskName3, taskName4 = "dataset-api", "zebedee", "another-task-name3", "another-task-name4"

	taskNames = map[string]bool{
		taskName1: true,
		taskName2: true,
		taskName3: true,
		taskName4: true,
	}
)

// SearchReindexAPIFeature is a type that contains all the requirements for running a godog (cucumber) feature that tests the SearchReindexAPIFeature endpoints.
type SearchReindexAPIFeature struct {
	apiVersion              string
	APIFeature              *componentTest.APIFeature
	AuthFeature             *componentTest.AuthorizationFeature
	Config                  *config.Config
	createdJob              *models.Job
	errorChan               chan error
	ErrorFeature            componentTest.ErrorFeature
	HTTPServer              *http.Server
	KafkaProducer           service.KafkaProducer
	kafkaProducerOutputData chan []byte
	KafkaMessageProducer    kafka.IProducer
	MongoClient             *mongo.JobStore
	MongoFeature            *componentTest.MongoFeature
	quitReadingOutput       chan bool
	responseBody            []byte
	fakeSearchAPI           *FakeAPI
	ServiceRunning          bool
	svc                     *service.Service
}

// NewSearchReindexAPIFeature returns a pointer to a new SearchReindexAPIFeature, which can then be used for testing the SearchReindexAPIFeature endpoints.
func NewSearchReindexAPIFeature(mongoFeature *componentTest.MongoFeature,
	authFeature *componentTest.AuthorizationFeature,
	fakeSearchAPI *FakeAPI) (*SearchReindexAPIFeature, error) {
	f := &SearchReindexAPIFeature{
		HTTPServer:     &http.Server{},
		errorChan:      make(chan error),
		ServiceRunning: false,
	}
	svcErrors := make(chan error, 1)
	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	mongodb := &mongo.JobStore{
		JobsCollection:  jobsCol,
		TasksCollection: tasksCol,
		Database:        utils.RandomDatabase(),
		URI:             mongoFeature.Server.URI(),
	}

	ctx := context.Background()
	if dbErr := mongodb.Init(ctx, cfg); dbErr != nil {
		return nil, fmt.Errorf("failed to initialise mongo DB: %w", dbErr)
	}

	f.MongoClient = mongodb

	f.AuthFeature = authFeature
	cfg.ZebedeeURL = f.AuthFeature.FakeAuthService.ResolveURL("")

	f.fakeSearchAPI = fakeSearchAPI
	cfg.SearchAPIURL = f.fakeSearchAPI.fakeHTTP.ResolveURL("")

	messageProducer := kafkatest.NewMessageProducer(true)
	messageProducer.CheckerFunc = DoGetKafkaProducerChecker
	f.KafkaMessageProducer = messageProducer

	kafkaProducer := &event.ReindexRequestedProducer{
		Marshaller: schema.ReindexRequestedEvent,
		Producer:   messageProducer,
	}
	f.KafkaProducer = kafkaProducer

	f.kafkaProducerOutputData = make(chan []byte, 9999)
	f.quitReadingOutput = make(chan bool)
	f.readOutputMessages()

	err = runSearchReindexAPIFeature(ctx, f, cfg, svcErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to run SearchReindexAPIFeature service: %w", err)
	}

	return f, nil
}

// runSearchReindexAPIFeature uses the InitialiserMock to create the mock services that are required for testing the SearchReindexAPIFeature
// It then uses these to create the external serviceList, which it passes into the service.Run function along with the fake identity and search clients.
func runSearchReindexAPIFeature(ctx context.Context, f *SearchReindexAPIFeature, cfg *config.Config, svcErrors chan error) error {
	var err error
	initFunctions := &serviceMock.InitialiserMock{
		DoGetHealthCheckFunc:           f.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:            f.DoGetHTTPServer,
		DoGetMongoDBFunc:               f.DoGetMongoDB,
		DoGetAuthorisationHandlersFunc: f.DoGetAuthorisationHandlers,
		DoGetKafkaProducerFunc:         f.DoGetKafkaProducer,
	}

	serviceList := service.NewServiceList(initFunctions)
	testIdentityClient := clientsidentity.New(cfg.ZebedeeURL)
	testSearchClient := clientssitesearch.NewClient(cfg.SearchAPIURL)

	f.svc, err = service.Run(ctx, cfg, serviceList, "1", "", "", svcErrors, testIdentityClient, taskNames, testSearchClient)
	return err
}

// InitAPIFeature initialises the APIFeature that's contained within a specific SearchReindexAPIFeature.
func (f *SearchReindexAPIFeature) InitAPIFeature() *componentTest.APIFeature {
	f.APIFeature = componentTest.NewAPIFeature(f.InitialiseService)

	return f.APIFeature
}

// Reset sets the resources within a specific SearchReindexAPIFeature back to their default values.
func (f *SearchReindexAPIFeature) Reset(mongoFail bool) error {
	if mongoFail {
		f.MongoClient.Database = "lost database connection"
	} else {
		f.MongoClient.Database = utils.RandomDatabase()
	}

	if f.Config == nil {
		cfg, err := config.Get()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}
		f.Config = cfg
	}

	return nil
}

// Close stops the *service.Service, which is pointed to from within the specific SearchReindexAPIFeature, from running.
func (f *SearchReindexAPIFeature) Close() error {
	close(f.quitReadingOutput)
	if f.svc != nil && f.ServiceRunning {
		err := f.svc.Close(context.Background())
		if err != nil {
			return fmt.Errorf("failed to close SearchReindexAPIFeature service: %w", err)
		}
		f.ServiceRunning = false
	}

	return nil
}

// InitialiseService returns the http.Handler that's contained within a specific SearchReindexAPIFeature.
func (f *SearchReindexAPIFeature) InitialiseService() (http.Handler, error) {
	return f.HTTPServer.Handler, nil
}

// DoGetHTTPServer takes a bind Address (string) and a router (http.Handler), which are used to set up an HTTPServer.
// The HTTPServer is in a specific SearchReindexAPIFeature and is returned.
func (f *SearchReindexAPIFeature) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	f.HTTPServer.Addr = bindAddr
	f.HTTPServer.Handler = router
	return f.HTTPServer
}

// DoGetHealthcheckOk returns a mock HealthChecker service for a specific SearchReindexAPIFeature.
func (f *SearchReindexAPIFeature) DoGetHealthcheckOk(cfg *config.Config, curTime, commit, version string) (service.HealthChecker, error) {
	return &serviceMock.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

// DoGetMongoDB returns a MongoDB, for the component test, which has a random database name and different URI to the one used by the API under test.
func (f *SearchReindexAPIFeature) DoGetMongoDB(ctx context.Context, cfg *config.Config) (service.MongoDataStorer, error) {
	return f.MongoClient, nil
}

// DoGetAuthorisationHandlers returns the mock AuthHandler that was created in the NewSearchReindexAPIFeature function.
func (f *SearchReindexAPIFeature) DoGetAuthorisationHandlers(ctx context.Context, cfg *config.Config) api.AuthHandler {
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

// DoGetKafkaProducer returns a mock kafka producer.
func (f *SearchReindexAPIFeature) DoGetKafkaProducer(ctx context.Context, cfg *config.Config) (service.KafkaProducer, error) {
	return f.KafkaProducer, nil
}

// DoGetKafkaProducerChecker is used to mock the kafka producer CheckerFunc
func DoGetKafkaProducerChecker(ctx context.Context, state *healthcheck.CheckState) error {
	return nil
}
