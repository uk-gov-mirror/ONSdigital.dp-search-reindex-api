package api

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-reindex-api/models"
)

//go:generate moq -out mock/mongo.go -pkg mock . MongoServer

// MongoServer defines the required methods from MongoDB
type MongoServer interface {
	Close(ctx context.Context) error
	Checker(ctx context.Context, state *healthcheck.CheckState) (err error)
	GetJobs(ctx context.Context, collectionID string) (images []models.Jobs, err error)
	GetJob(ctx context.Context, id string) (image *models.Job, err error)
	UpdateJob(ctx context.Context, id string, image *models.Job) (didChange bool, err error)
	UpsertJob(ctx context.Context, id string, image *models.Job) (err error)
	AcquireJobLock(ctx context.Context, id string) (lockID string, err error)
	UnlockJob(lockID string) error
}
