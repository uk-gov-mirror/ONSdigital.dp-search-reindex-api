package store

import (
	"context"
	"fmt"
	models "github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/log"
	uuid "github.com/satori/go.uuid"
)

type JobStorer struct {
	JobsMap	map[uuid.UUID]models.Job
}

// CreateJob creates a new Job resource and stores it in the JobsMap
func (js *JobStorer) CreateJob(ctx context.Context, id uuid.UUID, job_details models.Job) error{

	log.Event(ctx, "creating job", log.Data{"id": id})
	js.JobsMap = make(map[uuid.UUID]models.Job)
	fmt.Println(js.JobsMap)
	fmt.Println(job_details)

	js.JobsMap[id] = job_details
	fmt.Println(js.JobsMap[id])

	return nil
}
