package sdk

import (
	"context"
	"net/http"

	dpclients "github.com/ONSdigital/dp-api-clients-go/v2/headers"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out ./mocks/client.go -pkg mocks . Client

type Client interface {
	Checker(ctx context.Context, check *health.CheckState) error
	Health() *healthcheck.Client
	PostJob(ctx context.Context, reqHeaders Headers) (*RespHeaders, *models.Job, error)
	PatchJob(ctx context.Context, reqHeaders Headers, jobID string, patchList []PatchOperation) (*RespHeaders, error)
	PostTask(ctx context.Context, reqHeaders Headers, jobID string, taskToCreate models.TaskToCreate) (*RespHeaders, *models.Task, error)
	GetJob(ctx context.Context, reqheader Headers, jobID string) (*RespHeaders, *models.Job, error)
	GetJobs(ctx context.Context, reqheader Headers, options Options) (*RespHeaders, *models.Jobs, error)
	GetTask(ctx context.Context, reqHeaders Headers, jobID, taskName string) (*RespHeaders, *models.Task, error)
	GetTasks(ctx context.Context, reqHeaders Headers, jobID string) (*RespHeaders, *models.Tasks, error)
	PutJobNumberOfTasks(ctx context.Context, reqHeaders Headers, jobID, numTasks string) (*RespHeaders, error)
	PutTaskNumberOfDocs(ctx context.Context, reqHeaders Headers, jobID, taskName, docCount string) (*RespHeaders, error)
	URL() string
}

type Headers struct {
	IfMatch          string
	ServiceAuthToken string
	UserAuthToken    string
}

type RespHeaders struct {
	ETag string
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

// TaskNames is list of possible tasks associated with a job
var TaskNames = map[string]string{
	"zebedee":     "zebedee",
	"dataset-api": "dataset-api",
}

func (h *Headers) Add(req *http.Request) error {
	ctx := req.Context()

	if h == nil {
		log.Info(ctx, "the Headers struct is nil so there are no headers to add to the request")
		return nil
	}

	if h.IfMatch != "" {
		err := dpclients.SetIfMatch(req, h.IfMatch)
		if err != nil {
			logData := log.Data{"if_match_value": h.IfMatch}
			log.Error(ctx, "failed to set if match in request header", err, logData)
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
