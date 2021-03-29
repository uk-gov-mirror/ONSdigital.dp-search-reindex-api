package store

import (
	"context"
	models "github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/log"
)

//JobStorer is a type that contains a map, which can be used for storing Job resources with the keys being string values.
type JobStorer struct {
	JobsMap map[string]models.Job
}

// CreateJob creates a new Job resource and stores it in the JobsMap
func (js *JobStorer) CreateJob(ctx context.Context, id string) (models.Job, error) {

	log.Event(ctx, "creating job", log.Data{"id": id})
	//Create a Job that's populated with default values of all its attributes
	newJob := models.NewJob(id)

	//Only create a new JobsMap if one does not exist already
	if js.JobsMap == nil {
		js.JobsMap = make(map[string]models.Job)
	}

	js.JobsMap[id] = newJob
	log.Event(ctx, "adding job to map", log.Data{"Job details: ": js.JobsMap[id], "Map length: ": len(js.JobsMap)})

	return newJob, nil
}
