package api

import (
	"context"
	"io"
	"net/http"

	"github.com/ONSdigital/dp-authorisation/auth"
	dpHTTP "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-search-reindex-api/models"
)

//go:generate moq -out mock/data_storer.go -pkg mock . DataStorer
//go:generate moq -out mock/indexer.go -pkg mock . Indexer
//go:generate moq -out mock/reindex_requested_producer.go -pkg mock . ReindexRequestedProducer

// DataStorer is an interface for a type that can store and retrieve jobs
type DataStorer interface {
	CreateJob(ctx context.Context, id string) (job models.Job, err error)
	GetJob(ctx context.Context, id string) (job models.Job, err error)
	GetJobs(ctx context.Context, offset int, limit int) (job models.Jobs, err error)
	AcquireJobLock(ctx context.Context, id string) (lockID string, err error)
	UnlockJob(lockID string)
	PutNumberOfTasks(ctx context.Context, id string, count int) error
	CreateTask(ctx context.Context, jobID string, taskName string, numDocuments int) (task models.Task, err error)
	GetTask(ctx context.Context, jobID string, taskName string) (task models.Task, err error)
	GetTasks(ctx context.Context, offset int, limit int, jobID string) (job models.Tasks, err error)
	UpdateIndexName(indexName string, jobID string) error
	UpdateJobState(state string, jobID string) error
}

// Paginator defines the required methods from the paginator package
type Paginator interface {
	ValidatePaginationParameters(offsetParam string, limitParam string, totalCount int) (offset int, limit int, err error)
}

// AuthHandler provides authorisation checks on requests
type AuthHandler interface {
	Require(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc
}

// Indexer is a type that can create new ElasticSearch indexes
type Indexer interface {
	CreateIndex(ctx context.Context, serviceAuthToken, searchAPISearchURL string, httpClient dpHTTP.Clienter) (*http.Response, error)
	GetIndexNameFromResponse(ctx context.Context, body io.ReadCloser) (string, error)
}

// ReindexRequestedProducer is a type that can produce reindex-requested events
type ReindexRequestedProducer interface {
	ProduceReindexRequested(ctx context.Context, event models.ReindexRequested) error
}
