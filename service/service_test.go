package service_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/health"
	clientsidentity "github.com/ONSdigital/dp-api-clients-go/identity"
	clientssitesearch "github.com/ONSdigital/dp-api-clients-go/site-search"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-reindex-api/api"
	"github.com/ONSdigital/dp-search-reindex-api/api/mock"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/service"
	serviceMock "github.com/ONSdigital/dp-search-reindex-api/service/mock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "BuildTime"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
	errServer     = errors.New("HTTP Server error")
)

var (
	errMongoDB     = errors.New("mongoDB error")
	errHealthcheck = errors.New("healthCheck error")
)

var funcDoGetMongoDBErr = func(ctx context.Context, cfg *config.Config) (service.MongoDataStorer, error) {
	return nil, errMongoDB
}

var funcDoGetHealthcheckErr = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	return nil, errHealthcheck
}

var funcDoGetHTTPServerNil = func(bindAddr string, router http.Handler) service.HTTPServer {
	return nil
}

var taskName1, taskName2 = "dataset-api", "zebedee"

var taskNames = map[string]bool{
	taskName1: true,
	taskName2: true,
}

func TestRun(t *testing.T) {
	Convey("Having a set of mocked dependencies", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mongoDBMock := &serviceMock.MongoDataStorerMock{}

		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return nil
			},
		}

		failingServerMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return errServer
			},
		}

		authHandlerMock := &mock.AuthHandlerMock{}

		producerMock := &serviceMock.KafkaProducerMock{
			CloseFunc: func(ctx context.Context) error { return nil },
		}

		funcDoGetMongoDBOk := func(ctx context.Context, cfg *config.Config) (service.MongoDataStorer, error) {
			return mongoDBMock, nil
		}

		funcDoGetHealthcheckOk := func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
			return hcMock, nil
		}

		funcDoGetHTTPServer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return serverMock
		}

		funcDoGetFailingHTTPSerer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return failingServerMock
		}

		funcDoGetHealthClientOk := func(name string, url string) *health.Client {
			return &health.Client{
				URL:  url,
				Name: name,
			}
		}

		funcDoGetAuthorisationHandlersOk := func(ctx context.Context, cfg *config.Config) api.AuthHandler {
			return authHandlerMock
		}

		funcDoGetKafkaProducerOk := func(ctx context.Context, cfg *config.Config) (service.KafkaProducer, error) {
			return producerMock, nil
		}

		testIdentityClient := clientsidentity.New(cfg.ZebedeeURL)
		testSearchClient := clientssitesearch.NewClient(cfg.SearchAPIURL)

		Convey("Given that initialising mongoDB returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:            funcDoGetHTTPServerNil,
				DoGetMongoDBFunc:               funcDoGetMongoDBErr,
				DoGetHealthClientFunc:          funcDoGetHealthClientOk,
				DoGetAuthorisationHandlersFunc: funcDoGetAuthorisationHandlersOk,
				DoGetKafkaProducerFunc:         funcDoGetKafkaProducerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors, testIdentityClient, taskNames, testSearchClient)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errMongoDB)
				So(svcList.MongoDB, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising healthcheck returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:            funcDoGetHTTPServerNil,
				DoGetMongoDBFunc:               funcDoGetMongoDBOk,
				DoGetHealthCheckFunc:           funcDoGetHealthcheckErr,
				DoGetHealthClientFunc:          funcDoGetHealthClientOk,
				DoGetAuthorisationHandlersFunc: funcDoGetAuthorisationHandlersOk,
				DoGetKafkaProducerFunc:         funcDoGetKafkaProducerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors, testIdentityClient, taskNames, testSearchClient)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that mongo db Checker cannot be registered", func() {
			errAddCheckFail := errors.New("internal server error")
			hcMockAddFail := &serviceMock.HealthCheckerMock{
				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddCheckFail },
				StartFunc:    func(ctx context.Context) {},
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: funcDoGetHTTPServer,
				DoGetMongoDBFunc:    funcDoGetMongoDBOk,
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMockAddFail, nil
				},
				DoGetHealthClientFunc:          funcDoGetHealthClientOk,
				DoGetAuthorisationHandlersFunc: funcDoGetAuthorisationHandlersOk,
				DoGetKafkaProducerFunc: func(ctx context.Context, cfg *config.Config) (service.KafkaProducer, error) {
					return producerMock, nil
				},
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors, testIdentityClient, taskNames, testSearchClient)

			Convey("Then service Run fails, but the mongo db check tries to register", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, errAddCheckFail.Error())
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
				So(svcList.Auth, ShouldBeTrue)
				So(hcMockAddFail.AddCheckCalls(), ShouldHaveLength, 1)
				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Mongo DB")
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:            funcDoGetHTTPServer,
				DoGetMongoDBFunc:               funcDoGetMongoDBOk,
				DoGetHealthCheckFunc:           funcDoGetHealthcheckOk,
				DoGetHealthClientFunc:          funcDoGetHealthClientOk,
				DoGetAuthorisationHandlersFunc: funcDoGetAuthorisationHandlersOk,
				DoGetKafkaProducerFunc:         funcDoGetKafkaProducerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors, testIdentityClient, taskNames, testSearchClient)

			Convey("Then service Run succeeds and all the flags are set", func() {
				So(err, ShouldBeNil)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
				So(svcList.Auth, ShouldBeTrue)
			})

			Convey("The mongo DB checker is registered and health check and http servers are started", func() {
				So(hcMock.AddCheckCalls(), ShouldHaveLength, 4)
				So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Mongo DB")
				So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Zebedee")
				So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Search API")
				So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "Kafka producer")
				So(initMock.DoGetHTTPServerCalls(), ShouldHaveLength, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, "localhost:25700")
				So(hcMock.StartCalls(), ShouldHaveLength, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(serverMock.ListenAndServeCalls(), ShouldHaveLength, 1)
			})
		})

		Convey("Given that all dependencies are successfully initialised but the http server fails", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:            funcDoGetFailingHTTPSerer,
				DoGetMongoDBFunc:               funcDoGetMongoDBOk,
				DoGetHealthCheckFunc:           funcDoGetHealthcheckOk,
				DoGetHealthClientFunc:          funcDoGetHealthClientOk,
				DoGetAuthorisationHandlersFunc: funcDoGetAuthorisationHandlersOk,
				DoGetKafkaProducerFunc:         funcDoGetKafkaProducerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors, testIdentityClient, taskNames, testSearchClient)
			So(err, ShouldBeNil)

			Convey("Then the error is returned in the error channel", func() {
				sErr := <-svcErrors
				So(sErr.Error(), ShouldResemble, fmt.Sprintf("failure in http listen and serve: %s", errServer.Error()))
				So(failingServerMock.ListenAndServeCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Having a correctly initialised service", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcStopped := false
		serverStopped := false

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				serverStopped = true
				return nil
			},
		}

		// mongoDB Close will fail if healthcheck and http server are not already closed
		mongoDBMock := &serviceMock.MongoDataStorerMock{
			CloseFunc: func(ctx context.Context) error {
				if !hcStopped || !serverStopped {
					return errors.New("MongoDB closed before stopping healthcheck or HTTP server")
				}
				return nil
			},
		}

		producerMock := &serviceMock.KafkaProducerMock{
			CloseFunc: func(ctx context.Context) error { return nil },
		}

		testIdentityClient := clientsidentity.New(cfg.ZebedeeURL)
		testSearchClient := clientssitesearch.NewClient(cfg.SearchAPIURL)

		authHandlerMock := &mock.AuthHandlerMock{}

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return serverMock },
				DoGetMongoDBFunc: func(ctx context.Context, cfg *config.Config) (service.MongoDataStorer, error) {
					return mongoDBMock, nil
				},
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetHealthClientFunc: func(name, url string) *health.Client { return &health.Client{} },
				DoGetAuthorisationHandlersFunc: func(ctx context.Context, cfg *config.Config) api.AuthHandler {
					return authHandlerMock
				},
				DoGetKafkaProducerFunc: func(ctx context.Context, cfg *config.Config) (service.KafkaProducer, error) {
					return producerMock, nil
				},
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors, testIdentityClient, taskNames, testSearchClient)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(hcMock.StopCalls(), ShouldHaveLength, 1)
			So(serverMock.ShutdownCalls(), ShouldHaveLength, 1)
			So(mongoDBMock.CloseCalls(), ShouldHaveLength, 1)
			So(producerMock.CloseCalls(), ShouldHaveLength, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			failingServerMock := &serviceMock.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return failingServerMock },
				DoGetMongoDBFunc: func(ctx context.Context, cfg *config.Config) (service.MongoDataStorer, error) {
					return mongoDBMock, nil
				},
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetHealthClientFunc: func(name, url string) *health.Client { return &health.Client{} },
				DoGetAuthorisationHandlersFunc: func(ctx context.Context, cfg *config.Config) api.AuthHandler {
					return authHandlerMock
				},
				DoGetKafkaProducerFunc: func(ctx context.Context, cfg *config.Config) (service.KafkaProducer, error) {
					return producerMock, nil
				},
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors, testIdentityClient, taskNames, testSearchClient)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(hcMock.StopCalls(), ShouldHaveLength, 1)
			So(failingServerMock.ShutdownCalls(), ShouldHaveLength, 1)
			So(mongoDBMock.CloseCalls(), ShouldHaveLength, 1)
		})
	})
}
