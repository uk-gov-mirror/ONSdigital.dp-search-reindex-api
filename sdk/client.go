package sdk

import (
	"context"
	"fmt"
	"net/http"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/dp-search-reindex-api/models"
)

//go:generate moq -out ./mocks/client.go -pkg mocks . Client

type Client interface {
	Checker(ctx context.Context, check *health.CheckState) error
	Health() *healthcheck.Client
	PostJob(ctx context.Context, headers Headers) (models.Job, error)
	PostTasksCount(ctx context.Context, headers Headers, jobID string, payload []byte) (models.Task, error)
	URL() string
}

type Headers struct {
	ETag             string
	IfMatch          string
	ServiceAuthToken string
	UserAuthToken    string
}

type Options struct {
	Offset int
	Limit  int
	Sort   string
}

// TaskNames is list of possible tasks associated with a job
var TaskNames = map[string]string{
	"zebedee":     "zebedee",
	"dataset-api": "dataset-api",
}

func (h *Headers) Add(req *http.Request) {
	if h == nil {
		return
	}

	if h.ETag != "" {
		// TODO Set ETag header
		fmt.Println("currently not handling ETag header")
	}

	if h.IfMatch != "" {
		// TODO Set IfMatch header
		fmt.Println("currently not handling IfMatch header")
	}

	if h.ServiceAuthToken != "" {
		dprequest.AddServiceTokenHeader(req, h.ServiceAuthToken)
	}

	if h.UserAuthToken != "" {
		dprequest.AddFlorenceHeader(req, h.UserAuthToken)
	}
}
