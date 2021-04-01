package store

import (
	"context"
	"errors"
	models "github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/log"
)

//JobStorer is a type that contains a map, which can be used for storing Job resources with the keys being string values.
type JobStorer struct {
	JobsMap map[string]models.Job
}

// CreateJob creates a new Job resource and stores it in the JobsMap.
func (js *JobStorer) CreateJob(ctx context.Context, id string) (models.Job, error) {

	log.Event(ctx, "creating job", log.Data{"id": id})

	// If an empty id was passed in, return an error with a message.
	if id == "" {
		return models.Job{}, errors.New("id must not be an empty string")
	}

	//Create a Job that's populated with default values of all its attributes
	newJob := models.NewJob(id)

	//Only create a new JobsMap if one does not exist already
	if js.JobsMap == nil {
		js.JobsMap = make(map[string]models.Job)
	} else {
		//If JobsMap already exists then check that it does not already contain the id as a key
		if _, idPresent := js.JobsMap[id]; idPresent {
			return models.Job{}, errors.New("id must be unique")
		}
	}

	js.JobsMap[id] = newJob
	log.Event(ctx, "adding job to map", log.Data{"Job details: ": js.JobsMap[id], "Map length: ": len(js.JobsMap)})

	return newJob, nil
}

//GetJob gets a Job resource, from the JobsMap, that is associated with the id passed in.
func (js *JobStorer) GetJob(ctx context.Context, id string) (models.Job, error) {

	log.Event(ctx, "getting job", log.Data{"id": id})

	// If an empty id was passed in, return an error with a message.
	if id == "" {
		return models.Job{}, errors.New("id must not be an empty string")
	}

	//If no job store has been created yet, return an error with a message.
	if js.JobsMap == nil {
		return models.Job{}, errors.New("the job does not exist since there is no job store")
	} else {
		//If JobsMap already exists then check that it contains the id as a key
		if _, idPresent := js.JobsMap[id]; idPresent == false {
			return models.Job{}, errors.New("the job store does not contain the job id entered")
		}
	}

	job := js.JobsMap[id]
	log.Event(ctx, "getting job from map", log.Data{"Job details: ": js.JobsMap[id]})
	return job, nil
}
