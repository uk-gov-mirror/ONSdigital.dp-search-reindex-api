package api

import (
	"context"
	"github.com/ONSdigital/dp-search-reindex-api/models"
)

// JobStore defines the required methods from jobStore
type JobStore interface {
	CreateJob(ctx context.Context, id string, job *models.Job) (err error)
	GetJob(ctx context.Context, id string) (job *models.Job, err error)
	//GetJobs(ctx context.Context, collectionID string) (images []models.Image, err error)
	//UpdateJob(ctx context.Context, id string, image *models.Image) (didChange bool, err error)
}
