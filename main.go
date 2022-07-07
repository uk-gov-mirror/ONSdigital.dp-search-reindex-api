package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	clientssitesearch "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/service"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

const serviceName = "dp-search-reindex-api"

var (
	// BuildTime represents the time in which the service was built
	//BuildTime = "2022-07-07T09:30:03+01:00"
	BuildTime = "1657189133"
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit = "7b1fcc412086dae938b12a6d8c17317d694eabc9"
	// Version represents the version of the service that is running
	Version = "1.2.3"
)

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(ctx, "fatal runtime error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// run the service, providing an error channel for fatal errors
	svcErrors := make(chan error, 1)
	svcList := service.NewServiceList(&service.Init{})

	log.Info(ctx, "dp-search-reindex-api version", log.Data{"version": Version})

	// read config
	cfg, err := config.Get()
	if err != nil {
		return errors.Wrap(err, "error getting configuration")
	}

	log.Info(ctx, "config on startup", log.Data{"config": cfg, "build_time": BuildTime, "git-commit": GitCommit})

	// prefix token with "Bearer"
	cfg.ServiceAuthToken = "Bearer " + cfg.ServiceAuthToken

	// retrieve a list of tasks
	validTaskNames := strings.Split(cfg.TaskNameValues, ",")

	// create map of valid task name values
	taskNames := make(map[string]bool)
	for _, taskName := range validTaskNames {
		taskNames[taskName] = true
	}

	identityClient := clientsidentity.New(cfg.ZebedeeURL)
	searchClient := clientssitesearch.NewClient(cfg.SearchAPIURL)

	// start service
	svc, err := service.Run(ctx, cfg, svcList, BuildTime, GitCommit, Version, svcErrors, identityClient, taskNames, searchClient)
	if err != nil {
		return errors.Wrap(err, "running service failed")
	}

	// blocks until an os interrupt or a fatal error occurs
	select {
	case err := <-svcErrors:
		// ADD CODE HERE : call svc.Close(ctx) (or something specific)
		//  if there are any service connections like Kafka that you need to shut down
		return errors.Wrap(err, "service error received")
	case sig := <-signals:
		log.Info(ctx, "os signal received", log.Data{"signal": sig})
	}

	return svc.Close(ctx)
}
