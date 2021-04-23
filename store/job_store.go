package store

import (
	"context"
	"errors"
	"fmt"
	models "github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/log"
	"sort"
	"time"
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

type timeSlice []models.Job

func (p timeSlice) Len() int {
	return len(p)
}

func (p timeSlice) Less(i, j int) bool {
	return p[i].LastUpdated.Before(p[j].LastUpdated)
}

func (p timeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

//DeleteAllJobs empties the JobStore by deleting everything from the JobsMap
func (ds *DataStore) DeleteAllJobs(ctx context.Context) error {
	log.Event(ctx, "deleting all jobs from the job store")
	for k := range JobsMap {
		delete(JobsMap, k)
	}
	return nil
}

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

	log.Event(ctx, "getting list of jobs")

	jobs := models.Jobs{}
	numJobs := len(JobsMap)

	if numJobs == 0 {
		log.Event(ctx, "there are no jobs in the job store - so the list is empty")
		return jobs, nil
	}

	//Use a temporary map to order the jobs by last_updated in ascending order (i.e. oldest first)
	//Start by adding all the jobs from the JobsMap into the tempMap using last_updated as the key
	//tempMap := make(map[time.Time]models.Job)
	//ds.MoveJobs(ctx, tempMap)

	//Then delete everything from JobsMap
	//ds.DeleteAllJobs()

	//Use a sorted slice of last_updated times to put the jobs back into JobsMap but in last_updated order
	jobsToSort := make(timeSlice, 0, len(JobsMap))
	for k := range JobsMap {
		jobsToSort = append(jobsToSort, JobsMap[k])
	}
	fmt.Println("before...")
	fmt.Println(jobsToSort)
	sort.Sort(jobsToSort)
	fmt.Println("after...")
	fmt.Println(jobsToSort)

	//Create an empty list, the same length as the JobsMap, to put the requested Job resources into
	log.Event(ctx, "generating empty list")
	jobsList := make([]models.Job, numJobs)

	//Loop through the map and add each Job to the list
	i := 0
	for k := range JobsMap {
		log.Event(ctx, "adding job "+k+" to the list", log.Data{"JobsMap[" + k + "]": JobsMap[k]})
		jobsList[i] = JobsMap[k]
		i++
	}

	jobs.Job_List = jobsList

	return jobs, nil
}

//moveJobs moves all the jobs from the JobsMap into the tempMap, which uses last_updated as its key
func (ds *DataStore) MoveJobs(ctx context.Context, tempMap map[time.Time]models.Job) {
	for k := range JobsMap {
		keyVal := JobsMap[k].LastUpdated
		tempMap[keyVal] = JobsMap[k]
		log.Event(ctx, "adding job "+k+" to the temporary map", log.Data{"tempMap[" + keyVal.String() + "]": tempMap[keyVal]})
	}
}
