package store

import (
	"context"
	"errors"
	models "github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/log"
)

type JobStore interface {
	CreateJob(ctx context.Context, id string) (job models.Job, err error)
	GetJob(ctx context.Context, id string) (job models.Job, err error)
	//GetJobs(ctx context.Context, collectionID string) (images []models.Image, err error)
	//UpdateJob(ctx context.Context, id string, image *models.Image) (didChange bool, err error)
}

//DataStore is a type that contains a map, which can be used for storing Job resources with the keys being string values.
type DataStore struct {
	//JobsMap map[string]models.Job
	Jobs JobStore
}

var JobsMap = make(map[string]models.Job)

// CreateJob creates a new Job resource and stores it in the JobsMap.
func (ds *DataStore) CreateJob(ctx context.Context, id string) (models.Job, error) {

	log.Event(ctx, "creating job", log.Data{"id": id})

	// If an empty id was passed in, return an error with a message.
	if id == "" {
		return models.Job{}, errors.New("id must not be an empty string")
	}

	//Create a Job that's populated with default values of all its attributes
	newJob := models.NewJob(id)

	//Check that the JobsMap does not already contain the id as a key
	if _, idPresent := JobsMap[id]; idPresent {
		return models.Job{}, errors.New("id must be unique")
	}

	JobsMap[id] = newJob
	log.Event(ctx, "adding job to job store", log.Data{"Job details: ": JobsMap[id], "Map length: ": len(JobsMap)})

	return newJob, nil
}

//GetJob gets a Job resource, from the JobsMap, that is associated with the id passed in.
func (ds *DataStore) GetJob(ctx context.Context, id string) (models.Job, error) {

	log.Event(ctx, "getting job", log.Data{"id": id})

	// If an empty id was passed in, return an error with a message.
	if id == "" {
		return models.Job{}, errors.New("id must not be an empty string")
	}

	//If no job store has been created yet, return an error with a message.
	if JobsMap == nil {
		return models.Job{}, errors.New("the job does not exist since there is no job store")
	} else {
		//If JobsMap already exists then check that it contains the id as a key
		if _, idPresent := JobsMap[id]; idPresent == false {
			return models.Job{}, errors.New("the job store does not contain the job id entered")
		}
	}

	job := JobsMap[id]
	log.Event(ctx, "getting job from job store", log.Data{"Job details: ": JobsMap[id]})
	return job, nil
}
