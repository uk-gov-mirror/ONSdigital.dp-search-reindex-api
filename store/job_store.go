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
	GetJobs(ctx context.Context) (job models.Jobs, err error)
}

//DataStore is a type that contains an implementation of the JobStore interface, which can be used for creating and getting Job resources.
type DataStore struct {
	Jobs JobStore
}

//JobsMap is a map used for storing Job resources with the keys being string values.
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

	//Check that the JobsMap contains the id as a key
	if _, idPresent := JobsMap[id]; idPresent == false {
		return models.Job{}, errors.New("the job store does not contain the job id entered")
	}

	job := JobsMap[id]
	log.Event(ctx, "getting job from job store", log.Data{"Job details: ": JobsMap[id]})
	return job, nil
}

//GetJobs gets a list of Job resources, from the JobsMap, as defined by the four parameters passed in.
//As a default it will return all Job resources sorted by their last_updated time value.
func (ds *DataStore) GetJobs(ctx context.Context) (models.Jobs, error) {

	//log.Event(ctx, "getting jobs", log.Data{"id": id})
	log.Event(ctx, "getting jobs")

	//Create an empty list to put the Job resources into
	jobsList := make([]models.Job, len(JobsMap))

	//Loop through the map and add each Job to the list
	i := 0
	for k := range JobsMap {
		log.Event(ctx, "The value of JobsMap[" + k + "]", log.Data{"Job details: ": JobsMap[k]})
		jobsList[i] = JobsMap[k]
		i++
	}

	jobs := models.Jobs{}

	jobs.Job_List = jobsList

	return jobs, nil
}
