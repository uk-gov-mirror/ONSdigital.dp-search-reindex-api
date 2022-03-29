package sdk

import (
	"context"
	"fmt"
	"net/http"

	dpclients "github.com/ONSdigital/dp-api-clients-go/v2/headers"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/dp-search-reindex-api/models"
)

//go:generate moq -out ./mocks/client.go -pkg mocks . Client

type Client interface {
	PostJob(ctx context.Context, headers Headers) (models.Job, error)
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
	Operation string
	Path      string
	Value     string
}

type PatchOpsList struct {
	PatchList []PatchOperation
}

func (h *Headers) Add(req *http.Request) {
	if h == nil {
		return
	}
	fmt.Println("the ETag value is: " + h.ETag)
	fmt.Println("the auth value is: " + h.ServiceAuthToken)
	fmt.Println("the IfMatch value is: " + h.IfMatch)
	if h.ETag != "" {
		dpclients.SetETag(req, h.ETag)
	}

	if h.IfMatch != "" {
		dpclients.SetIfMatch(req, h.IfMatch)
	}

	if h.ServiceAuthToken != "" {
		dprequest.AddServiceTokenHeader(req, h.ServiceAuthToken)
	}

	if h.UserAuthToken != "" {
		dprequest.AddFlorenceHeader(req, h.UserAuthToken)
	}
}
