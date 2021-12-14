package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpMongodb "github.com/ONSdigital/dp-mongodb"
	dpMongoLock "github.com/ONSdigital/dp-mongodb/dplock"
	dpMongoHealth "github.com/ONSdigital/dp-mongodb/health"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// JobStore is a type that contains an implementation of the MongoJobStorer interface, which can be used for creating
// and getting Job resources. It also represents a simplistic MongoDB configuration, with session,
// health and lock clients
type JobStore struct {
	Session         *mgo.Session
	URI             string
	Database        string
	JobsCollection  string
	LocksCollection string
	TasksCollection string
	client          *dpMongoHealth.Client
	healthClient    *dpMongoHealth.CheckMongoClient
	lockClient      *dpMongoLock.Lock
	cfg             *config.Config
}

// Init creates a new mgo.Session with a strong consistency and a write mode of "majority".
func (m *JobStore) Init(ctx context.Context, cfg *config.Config) (err error) {
	m.cfg = cfg
	if m.Session != nil {
		return errors.New("session already exists")
	}

	// Create session
	if m.Session, err = mgo.Dial(m.URI); err != nil {
		return err
	}
	m.Session.EnsureSafe(&mgo.Safe{WMode: "majority"})
	m.Session.SetMode(mgo.Strong, true)

	databaseCollectionBuilder := make(map[dpMongoHealth.Database][]dpMongoHealth.Collection)
	databaseCollectionBuilder[(dpMongoHealth.Database)(m.Database)] = []dpMongoHealth.Collection{(dpMongoHealth.Collection)(m.JobsCollection),
		(dpMongoHealth.Collection)(m.LocksCollection), (dpMongoHealth.Collection)(m.TasksCollection)}
	// Create client and healthClient from session
	m.client = dpMongoHealth.NewClientWithCollections(m.Session, databaseCollectionBuilder)
	m.healthClient = &dpMongoHealth.CheckMongoClient{
		Client:      *m.client,
		Healthcheck: m.client.Healthcheck,
	}

	// Create MongoDB lock client, which also starts the purger loop
	m.lockClient = dpMongoLock.New(ctx, m.Session, m.Database, m.JobsCollection)

	return nil
}

// CreateJob creates a new job, with the given id, in the collection, and assigns default values to its attributes
func (m *JobStore) CreateJob(ctx context.Context, id string) (models.Job, error) {
	log.Info(ctx, "creating job in mongo DB", log.Data{"id": id})

	// If an empty id was passed in, return an error with a message.
	if id == "" {
		return models.Job{}, ErrEmptyIDProvided
	}

	s := m.Session.Copy()
	defer s.Close()
	var jobToFind models.Job

	// Get all the jobs where state is "in-progress" and order them by "-reindex_started"
	// (in reverse order of reindex_started so that the one started most recently is first in the Iter).
	iter := s.DB(m.Database).C(m.JobsCollection).Find(bson.M{"state": "in-progress"}).Sort("-reindex_started").Iter()
	result := models.Job{}

	// Check that there are no jobs in progress, which started between a calculated "check from" time and now.
	// The checkFromTime is calculated by subtracting the configured variable "MaxReindexJobRuntime" from now.
	checkFromTime := time.Now().Add(-1 * m.cfg.MaxReindexJobRuntime)
	var jobStartTime time.Time
	for iter.Next(&result) {
		// If the start time of the job in progress is later than the checkFromTime but earlier than now then a new job should not be created yet.
		jobStartTime = result.ReindexStarted
		if jobStartTime.After(checkFromTime) && jobStartTime.Before(time.Now()) {
			log.Info(ctx, "found job in progress", log.Data{"id": result.ID, "state": result.State, "start time": jobStartTime})
			return models.Job{}, ErrExistingJobInProgress
		}
	}
	defer func() {
		if err := iter.Close(); err != nil {
			log.Error(ctx, "error closing iterator", err)
		}
	}()

	// Create a Job that's populated with default values of all its attributes
	newJob, err := models.NewJob(id)
	if err != nil {
		log.Error(ctx, "error creating new job", err)
	}

	// Check that the jobs collection does not already contain the id as a key
	err = s.DB(m.Database).C(m.JobsCollection).Find(bson.M{"_id": id}).One(&jobToFind)
	if err != nil {
		if err == mgo.ErrNotFound {
			// This means we CAN insert the job as it does not already exist
			err = s.DB(m.Database).C(m.JobsCollection).Insert(newJob)
			if err != nil {
				return models.Job{}, errors.New("error inserting job into mongo DB")
			}
			log.Info(ctx, "adding job to jobs collection", log.Data{"Job details: ": newJob})
		} else {
			return models.Job{}, err
		}
	} else {
		// As there is no error this means that it found a job with the id we're trying to insert
		return models.Job{}, ErrDuplicateIDProvided
	}

	return newJob, err
}

// UpdateIndexName updates the search_index_name, of the relevant Job Resource, with the indexName provided
func (m *JobStore) UpdateIndexName(indexName string, jobID string) error {
	s := m.Session.Copy()
	defer s.Close()
	updates := make(bson.M)
	updates["search_index_name"] = indexName
	updates["last_updated"] = time.Now()
	err := m.UpdateJob(updates, s, jobID)
	return err
}

// CreateTask creates a new task, for the given API and job ID, in the collection, and assigns default values to its attributes
func (m *JobStore) CreateTask(ctx context.Context, jobID string, taskName string, numDocuments int) (models.Task, error) {
	log.Info(ctx, "creating task in mongo DB", log.Data{"jobID": jobID, "taskName": taskName, "numDocuments": numDocuments})

	// If an empty job id was passed in, return an error with a message.
	if jobID == "" {
		return models.Task{}, ErrEmptyIDProvided
	}

	s := m.Session.Copy()
	defer s.Close()

	// Check that the jobs collection contains the job that the task will be part of
	var jobToFind models.Job
	jobToFind.ID = jobID
	_, err := m.findJob(s, jobID, jobToFind)
	if err != nil {
		log.Error(ctx, "error finding job for task", err)
		if err == mgo.ErrNotFound {
			return models.Task{}, ErrJobNotFound
		} else {
			return models.Task{}, fmt.Errorf("an unexpected error has occurred: %w", err)
		}
	}

	newTask := models.NewTask(jobID, taskName, numDocuments, m.cfg.BindAddr)

	err = m.UpsertTask(jobID, taskName, newTask)
	if err != nil {
		return models.Task{}, fmt.Errorf("error creating or overwriting task in mongo DB: %w", err)
	}
	log.Info(ctx, "creating or overwriting task in tasks collection", log.Data{"Task details: ": newTask})

	return newTask, err
}

// GetTask retrieves the details of a particular task, from the collection, specified by its task name and associated job id
func (m *JobStore) GetTask(ctx context.Context, jobID string, taskName string) (models.Task, error) {
	s := m.Session.Copy()
	defer s.Close()
	log.Info(ctx, "getting task from the data store", log.Data{"jobID": jobID, "taskName": taskName})
	var task models.Task

	// If an empty jobID or taskName was passed in, return an error with a message.
	if jobID == "" {
		return task, ErrEmptyIDProvided
	} else if taskName == "" {
		return task, ErrEmptyTaskNameProvided
	}

	_, err := m.findJob(s, jobID, models.Job{})
	if err != nil {
		if err == mgo.ErrNotFound {
			return task, ErrJobNotFound
		}
		return task, err
	}

	err = s.DB(m.Database).C(m.TasksCollection).Find(bson.M{"job_id": jobID, "task_name": taskName}).One(&task)
	if err != nil {
		if err == mgo.ErrNotFound {
			return task, ErrTaskNotFound
		}
		return task, err
	}

	return task, err
}

// GetTasks retrieves all the tasks, from the collection, and lists them in order of last_updated
func (m *JobStore) GetTasks(ctx context.Context, offset int, limit int, jobID string) (models.Tasks, error) {
	s := m.Session.Copy()
	defer s.Close()
	log.Info(ctx, "getting list of tasks")
	results := models.Tasks{}

	// If an empty jobID or taskName was passed in, return an error with a message.
	if jobID == "" {
		return results, ErrEmptyIDProvided
	}

	_, err := m.findJob(s, jobID, models.Job{})
	if err != nil {
		if err == mgo.ErrNotFound {
			return results, ErrJobNotFound
		}
		return results, err
	}

	numTasks := 0
	numTasks, err = s.DB(m.Database).C(m.TasksCollection).Find(bson.M{"job_id": jobID}).Count()
	if err != nil {
		log.Error(ctx, "error counting tasks for given job id", err, log.Data{"job_id": jobID})
		return results, err
	}
	log.Info(ctx, "number of tasks found in tasks collection", log.Data{"numTasks": numTasks})

	if numTasks == 0 {
		log.Info(ctx, "there are no tasks in the data store - so the list is empty")
		results.TaskList = make([]models.Task, 0)
		return results, nil
	}

	//Get the requested tasks from the tasks collection, using the given job_id, offset, and limit, and order them by last_updated
	tasksQuery := s.DB(m.Database).C(m.TasksCollection).Find(bson.M{"job_id": jobID}).Skip(offset).Limit(limit).Sort("last_updated")
	tasks := make([]models.Task, numTasks)
	if err := tasksQuery.All(&tasks); err != nil {
		return results, err
	}

	results.TaskList = tasks
	results.Count = len(tasks)
	results.Limit = limit
	results.Offset = offset
	results.TotalCount = numTasks
	log.Info(ctx, "list of tasks - sorted by last_updated", log.Data{"Sorted tasks: ": results.TaskList})

	return results, nil
}

// AcquireJobLock tries to lock the provided jobID.
// If the job is already locked, this function will block until it's released,
// at which point we acquire the lock and return.
func (m *JobStore) AcquireJobLock(ctx context.Context, jobID string) (lockID string, err error) {
	return m.lockClient.Acquire(ctx, jobID)
}

// UnlockJob releases an exclusive mongoDB lock for the provided lockId (if it exists)
func (m *JobStore) UnlockJob(lockID string) error {
	return m.lockClient.Unlock(lockID)
}

// Close closes the mongo session and returns any error
func (m *JobStore) Close(ctx context.Context) error {
	m.lockClient.Close(ctx)
	return dpMongodb.Close(ctx, m.Session)
}

// GetJobs retrieves all the jobs, from the collection, and lists them in order of last_updated
func (m *JobStore) GetJobs(ctx context.Context, offset int, limit int) (models.Jobs, error) {
	s := m.Session.Copy()
	defer s.Close()
	log.Info(ctx, "getting list of jobs")

	results := models.Jobs{}

	numJobs, err := s.DB(m.Database).C(m.JobsCollection).Count()
	if err != nil {
		log.Error(ctx, "error counting jobs", err)
		return results, err
	}
	log.Info(ctx, "number of jobs found in jobs collection", log.Data{"numJobs": numJobs})

	if numJobs == 0 {
		log.Info(ctx, "there are no jobs in the data store - so the list is empty")
		results.JobList = make([]models.Job, 0)
		return results, nil
	}

	jobsQuery := s.DB(m.Database).C(m.JobsCollection).Find(bson.M{}).Skip(offset).Limit(limit).Sort("last_updated")
	jobs := make([]models.Job, numJobs)
	if err := jobsQuery.All(&jobs); err != nil {
		return results, err
	}

	results.JobList = jobs
	results.Count = len(jobs)
	results.Limit = limit
	results.Offset = offset
	results.TotalCount = numJobs
	log.Info(ctx, "list of jobs - sorted by last_updated", log.Data{"Sorted jobs: ": results.JobList})

	return results, nil
}

// GetJob retrieves the details of a particular job, from the collection, specified by its id
func (m *JobStore) GetJob(ctx context.Context, id string) (models.Job, error) {
	s := m.Session.Copy()
	defer s.Close()
	log.Info(ctx, "getting job by ID", log.Data{"id": id})

	// If an empty id was passed in, return an error with a message.
	if id == "" {
		return models.Job{}, ErrEmptyIDProvided
	}

	var job models.Job
	job, err := m.findJob(s, id, job)
	if err != nil {
		if err == mgo.ErrNotFound {
			return models.Job{}, ErrJobNotFound
		}
		return models.Job{}, err
	}

	return job, nil
}

func (m *JobStore) findJob(s *mgo.Session, jobID string, job models.Job) (models.Job, error) {
	err := s.DB(m.Database).C(m.JobsCollection).Find(bson.M{"_id": jobID}).One(&job)
	return job, err
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *JobStore) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}

// PutNumberOfTasks updates the number_of_tasks in a particular job, from the collection, specified by its id
func (m *JobStore) PutNumberOfTasks(ctx context.Context, id string, numTasks int) error {
	s := m.Session.Copy()
	defer s.Close()
	log.Info(ctx, "putting number of tasks", log.Data{"id": id, "numTasks": numTasks})

	updates := make(bson.M)
	updates["number_of_tasks"] = numTasks
	updates["last_updated"] = time.Now()
	err := m.UpdateJob(updates, s, id)

	return err
}

// UpdateJob updates a particular job with the values passed in through the 'updates' input parameter
func (m *JobStore) UpdateJob(updates bson.M, s *mgo.Session, id string) error {
	update := bson.M{"$set": updates}
	if err := s.DB(m.Database).C(m.JobsCollection).UpdateId(id, update); err != nil {
		if err == mgo.ErrNotFound {
			return ErrJobNotFound
		}
		return err
	}
	return nil
}

// UpsertTask creates a new task document or overwrites an existing one
func (m *JobStore) UpsertTask(jobID, taskName string, task models.Task) error {
	s := m.Session.Copy()
	defer s.Close()

	selector := bson.M{
		"task_name": taskName,
		"job_id":    jobID,
	}

	task.LastUpdated = time.Now()

	update := bson.M{
		"$set": task,
	}

	_, err := s.DB(m.Database).C(m.TasksCollection).Upsert(selector, update)
	return err
}
