package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpMongodb "github.com/ONSdigital/dp-mongodb"
	dpMongoLock "github.com/ONSdigital/dp-mongodb/dplock"
	dpMongoHealth "github.com/ONSdigital/dp-mongodb/health"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/pagination"
	"github.com/ONSdigital/log.go/log"
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
	Collection      string
	LocksCollection string
	TasksCollection string
	client          *dpMongoHealth.Client
	healthClient    *dpMongoHealth.CheckMongoClient
	lockClient      *dpMongoLock.Lock
	cfg             *config.Config
}

// CreateJob creates a new job, with the given id, in the collection, and assigns default values to its attributes
func (m *JobStore) CreateJob(ctx context.Context, id string) (job models.Job, err error) {
	log.Event(ctx, "creating job in mongo DB", log.Data{"id": id})

	// If an empty id was passed in, return an error with a message.
	if id == "" {
		return models.Job{}, ErrEmptyIDProvided
	}

	s := m.Session.Copy()
	defer s.Close()
	var jobToFind models.Job

	// Get all the jobs where state is "in-progress" and order them by "-reindex_started"
	// (in reverse order of reindex_started so that the one started most recently is first in the Iter).
	iter := s.DB(m.Database).C(m.Collection).Find(bson.M{"state": "in-progress"}).Sort("-reindex_started").Iter()
	result := models.Job{}

	// Check that there are no jobs in progress, which started between a calculated "check from" time and now.
	// The checkFromTime is calculated by subtracting the configured variable "MaxReindexJobRuntime" from now.
	checkFromTime := time.Now().Add(-1 * m.cfg.MaxReindexJobRuntime)
	var jobStartTime time.Time
	for iter.Next(&result) {
		// If the start time of the job in progress is later than the checkFromTime but earlier than now then a new job should not be created yet.
		jobStartTime = result.ReindexStarted
		if jobStartTime.After(checkFromTime) && jobStartTime.Before(time.Now()) {
			log.Event(ctx, "found job in progress", log.Data{"id": result.ID, "state": result.State, "start time": jobStartTime})
			return models.Job{}, ErrExistingJobInProgress
		}
	}
	defer func() {
		if err := iter.Close(); err != nil {
			log.Event(ctx, "error closing iterator", log.ERROR, log.Error(err))
		}
	}()

	// Create a Job that's populated with default values of all its attributes
	newJob, err := models.NewJob(id)
	if err != nil {
		log.Event(ctx, "error creating new job", log.ERROR, log.Error(err))
	}

	// Check that the jobs collection does not already contain the id as a key
	err = s.DB(m.Database).C(m.Collection).Find(bson.M{"id": id}).One(&jobToFind)
	if err != nil {
		if err == mgo.ErrNotFound {
			// This means we CAN insert the job as it does not already exist
			err = s.DB(m.Database).C(m.Collection).Insert(newJob)
			if err != nil {
				return models.Job{}, errors.New("error inserting job into mongo DB")
			}
			log.Event(ctx, "adding job to jobs collection", log.Data{"Job details: ": newJob})
		} else {
			return models.Job{}, err
		}
	} else {
		// As there is no error this means that it found a job with the id we're trying to insert
		return models.Job{}, ErrDuplicateIDProvided
	}

	return newJob, nil
}

// CreateTask creates a new task, for the given API and job ID, in the collection, and assigns default values to its attributes
func (m *JobStore) CreateTask(ctx context.Context, jobID string, nameOfApi string, numDocuments int) (task models.Task, err error) {
	log.Event(ctx, "creating task in mongo DB", log.Data{"jobID": jobID, "nameOfApi": nameOfApi})

	// Create a Task that's populated with the provided api name and number of documents
	newTask, err := models.NewTask(jobID, nameOfApi, numDocuments)
	if err != nil {
		log.Event(ctx, "error creating new task", log.ERROR, log.Error(err))
	}

	s := m.Session.Copy()
	defer s.Close()

	err = s.DB(m.Database).C(m.TasksCollection).Insert(newTask)
	if err != nil {
		return models.Task{}, errors.New("error inserting task into mongo DB")
	}
	log.Event(ctx, "adding task to tasks collection", log.Data{"Task details: ": newTask})

	return newTask, nil
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
	databaseCollectionBuilder[(dpMongoHealth.Database)(m.Database)] = []dpMongoHealth.Collection{(dpMongoHealth.Collection)(m.Collection),
		(dpMongoHealth.Collection)(m.LocksCollection)}
	// Create client and healthClient from session
	m.client = dpMongoHealth.NewClientWithCollections(m.Session, databaseCollectionBuilder)
	m.healthClient = &dpMongoHealth.CheckMongoClient{
		Client:      *m.client,
		Healthcheck: m.client.Healthcheck,
	}

	// Create MongoDB lock client, which also starts the purger loop
	m.lockClient = dpMongoLock.New(ctx, m.Session, m.Database, m.Collection)
	return nil
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
func (m *JobStore) GetJobs(ctx context.Context, offsetParam string, limitParam string) (models.Jobs, error) {
	s := m.Session.Copy()
	defer s.Close()
	log.Event(ctx, "getting list of jobs", log.INFO)

	results := models.Jobs{}
	numJobs, _ := s.DB(m.Database).C(m.Collection).Count()
	log.Event(ctx, "number of jobs found in jobs collection", log.Data{"numJobs": numJobs})

	if numJobs == 0 {
		log.Event(ctx, "there are no jobs in the job store - so the list is empty", log.INFO)
		return results, nil
	}

	// Get all the jobs from the jobs collection and order them by lastupdated
	iter := s.DB(m.Database).C(m.Collection).Find(bson.M{}).Sort("lastupdated").Iter()
	defer func() {
		err := iter.Close()
		if err != nil {
			log.Event(ctx, "error closing iterator", log.ERROR, log.Error(err))
		}
	}()

	jobs := make([]models.Job, numJobs)
	if err := iter.All(&jobs); err != nil {
		return results, err
	}

	paginator := pagination.NewPaginator(m.cfg.DefaultLimit, m.cfg.DefaultOffset, m.cfg.DefaultMaxLimit)
	offset, limit, err := paginator.ValidatePaginationParameters(offsetParam, limitParam, numJobs)
	if err != nil {
		return results, err
	}
	jobs = modifyJobs(jobs, offset, limit)

	results.JobList = jobs
	results.Count = len(jobs)
	results.Limit = limit
	results.Offset = offset
	results.TotalCount = numJobs
	log.Event(ctx, "list of jobs - sorted by last_updated", log.Data{"Sorted jobs: ": results.JobList}, log.INFO)

	return results, nil
}

// GetJob retrieves the details of a particular job, from the collection, specified by its id
func (m *JobStore) GetJob(ctx context.Context, id string) (models.Job, error) {
	s := m.Session.Copy()
	defer s.Close()
	log.Event(ctx, "getting job by ID", log.Data{"id": id})

	// If an empty id was passed in, return an error with a message.
	if id == "" {
		return models.Job{}, ErrEmptyIDProvided
	}

	var job models.Job
	err := s.DB(m.Database).C(m.Collection).Find(bson.M{"_id": id}).One(&job)
	if err != nil {
		if err == mgo.ErrNotFound {
			return models.Job{}, ErrJobNotFound
		}
		return models.Job{}, err
	}

	return job, nil
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *JobStore) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}

// PutNumberOfTasks updates the number_of_tasks in a particular job, from the collection, specified by its id
func (m *JobStore) PutNumberOfTasks(ctx context.Context, id string, numTasks int) (err error) {
	s := m.Session.Copy()
	defer s.Close()
	log.Event(ctx, "putting number of tasks", log.Data{"id": id, "numTasks": numTasks})

	updates := make(bson.M)
	updates["number_of_tasks"] = numTasks
	updates["last_updated"] = time.Now()
	err = m.UpdateJob(updates, s, id)

	return err
}

// UpdateJob updates a particular job with the values passed in through the 'updates' input parameter
func (m *JobStore) UpdateJob(updates bson.M, s *mgo.Session, id string) error {
	update := bson.M{"$set": updates}
	if err := s.DB(m.Database).C(m.Collection).UpdateId(id, update); err != nil {
		if err == mgo.ErrNotFound {
			return ErrJobNotFound
		}
		return err
	}
	return nil
}

// modifyJobs takes a slice, of all the jobs in the Job Store, determined by the offset and limit values
func modifyJobs(jobs []models.Job, offset int, limit int) []models.Job {
	var modifiedJobs []models.Job
	lastIndex := offset + limit
	if lastIndex >= len(jobs) {
		modifiedJobs = jobs[offset:]
	} else {
		modifiedJobs = jobs[offset:lastIndex]
	}
	return modifiedJobs
}
