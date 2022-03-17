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
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string

	/* NOTE: replace the above with the below to run code with for example vscode debugger.
	   BuildTime string = "1601119818"
	   GitCommit string = "6584b786caac36b6214ffe04bf62f058d4021538"
	   Version   string = "v0.1.0"
	*/
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

	// Run the service, providing an error channel for fatal errors
	svcErrors := make(chan error, 1)
	svcList := service.NewServiceList(&service.Init{})

	log.Info(ctx, "dp-search-reindex-api version", log.Data{"version": Version})

	// Read config
	cfg, err := config.Get()
	if err != nil {
		return errors.Wrap(err, "error getting configuration")
	}

	validTaskNames := strings.Split(cfg.TaskNameValues, ",")

	// create map of valid task name values
	taskNames := make(map[string]bool)
	for _, taskName := range validTaskNames {
		taskNames[taskName] = true
	}

	identityClient := clientsidentity.New(cfg.ZebedeeURL)
	searchClient := clientssitesearch.NewClient(cfg.SearchAPIURL)

	// Start service
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
