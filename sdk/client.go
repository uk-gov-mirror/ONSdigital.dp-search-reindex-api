package sdk

import (
	"context"
	"net/http"

	dpclients "github.com/ONSdigital/dp-api-clients-go/v2/headers"
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
	PatchJob(ctx context.Context, headers Headers, jobID string, body []PatchOperation) error
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

type PatchOperation struct {
	Op    string
	Path  string
	Value interface{}
}

func (h *Headers) Add(req *http.Request) error {
	if h == nil {
		return nil
	}

	if h.ETag != "" {
		err := dpclients.SetETag(req, h.ETag)
		if err != nil {
			return err
		}
	}

	if h.IfMatch != "" {
		err := dpclients.SetIfMatch(req, h.IfMatch)
		if err != nil {
			return err
		}
	}

	if h.ServiceAuthToken != "" {
		dprequest.AddServiceTokenHeader(req, h.ServiceAuthToken)
	}

	if h.UserAuthToken != "" {
		dprequest.AddFlorenceHeader(req, h.UserAuthToken)
	}

	return nil
}
