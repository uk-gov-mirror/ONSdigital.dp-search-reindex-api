package service

import (
	"context"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	clientssitesearch "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	"github.com/ONSdigital/dp-net/v2/handlers"
	dpHTTP "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-reindex-api/api"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/reindex"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the Search Reindex API
type Service struct {
	config      *config.Config
	server      HTTPServer
	router      *mux.Router
	api         *api.API
	serviceList *ExternalServiceList
	healthCheck HealthChecker
	mongoDB     MongoDataStorer
	producer    KafkaProducer
}

// Run the service
func Run(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error, identityClient *clientsidentity.Client,
	taskNames map[string]bool, searchClient *clientssitesearch.Client) (*Service, error) {
	log.Info(ctx, "running service")

	// Get HTTP Server with collectionID checkHeader middleware
	r := mux.NewRouter()
	middleware := alice.New(handlers.CheckHeader(handlers.CollectionID))
	s := serviceList.GetHTTPServer(cfg.BindAddr, middleware.Then(r))

	// Get MongoDB client
	mongoDB, err := serviceList.GetMongoDB(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise mongo DB", err)
		return nil, err
	}

	var a *api.API

	permissions := serviceList.GetAuthorisationHandlers(ctx, cfg)
	httpClient := dpHTTP.NewClient()
	indexer := &reindex.Reindex{}

	// Get Kafka producer
	producer, err := serviceList.GetKafkaProducer(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise kafka producer", err)
		return nil, err
	}

	// Setup the API
	api.Setup(r, mongoDB, permissions, taskNames, cfg, httpClient, indexer, producer)

	// Get HealthCheck
	hc, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthCheck", err)
		return nil, err
	}
	if err = hc.AddCheck("Mongo DB", mongoDB.Checker); err != nil {
		log.Error(ctx, "error adding check for mongo db", err)
		return nil, err
	}
	if err = hc.AddCheck("Zebedee", identityClient.Checker); err != nil {
		log.Error(ctx, "error adding check for zebedee", err)
		return nil, err
	}
	if err = hc.AddCheck("Search API", searchClient.Checker); err != nil {
		log.Error(ctx, "error adding check for search api", err)
		return nil, err
	}
	if err = hc.AddCheck("Kafka producer", producer.Checker); err != nil {
		log.Error(ctx, "error adding check for Kafka producer", err)
		return nil, err
	}

	r.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
	hc.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return &Service{
		config:      cfg,
		server:      s,
		router:      r,
		api:         a,
		serviceList: serviceList,
		healthCheck: hc,
		mongoDB:     mongoDB,
		producer:    producer,
	}, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.config.GracefulShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)

	// track shutdown gracefully closes up
	var gracefulShutdown bool

	go func() {
		defer cancel()
		var hasShutdownError bool

		// stop healthCheck, as it depends on everything else
		if svc.serviceList.HealthCheck {
			svc.healthCheck.Stop()
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.server.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}

		// close kafka producer
		if svc.serviceList.KafkaProducer {
			if err := svc.producer.Close(ctx); err != nil {
				log.Error(ctx, "error closing kafka producer", err)
				hasShutdownError = true
			}
		}

		// close API
		if err := svc.api.Close(ctx); err != nil {
			log.Error(ctx, "error closing API", err)
			hasShutdownError = true
		}

		// close mongoDB
		if svc.serviceList.MongoDB {
			if err := svc.mongoDB.Close(ctx); err != nil {
				log.Error(ctx, "error closing mongoDB", err)
				hasShutdownError = true
			}
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	if !gracefulShutdown {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}
