package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-search-reindex-api/models"
)

//go:generate moq -out mock/data_storer_temp.go -pkg mock . DataStorer

// DataStorer is an interface for a type that can store and retrieve jobs
type DataStorer interface {
	CreateJob(ctx context.Context, id string) (job models.Job, err error)
	GetJob(ctx context.Context, id string) (job models.Job, err error)
	GetJobs(ctx context.Context, offsetParam string, limitParam string) (job models.Jobs, err error)
	AcquireJobLock(ctx context.Context, id string) (lockID string, err error)
	UnlockJob(lockID string) error
	PutNumberOfTasks(ctx context.Context, id string, count int) error
	CreateTask(ctx context.Context, jobID string, taskName string, numDocuments int) (task models.Task, err error)
	GetTask(ctx context.Context, jobID string, taskName string) (task models.Task, err error)
	GetTasks(ctx context.Context, offsetParam string, limitParam string, jobID string) (job models.Tasks, err error)
}

// Paginator defines the required methods from the paginator package
type Paginator interface {
	ValidatePaginationParameters(offsetParam string, limitParam string, totalCount int) (offset int, limit int, err error)
}

// AuthHandler provides authorisation checks on requests
type AuthHandler interface {
	Require(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc
}
