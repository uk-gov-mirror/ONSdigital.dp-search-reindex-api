package mongo

import (
	"context"
	"errors"
	dpMongodb "github.com/ONSdigital/dp-mongodb"
	dpMongoLock "github.com/ONSdigital/dp-mongodb/dplock"
	dpMongoHealth "github.com/ONSdigital/dp-mongodb/health"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/log"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"sort"
)

// MgoJobStore defines the required methods from MongoDB
type MgoJobStore interface {
	Close(ctx context.Context) error
	CreateJob(ctx context.Context, id string) (job models.Job, err error)
	GetJob(ctx context.Context, id string) (job models.Job, err error)
	GetJobs(ctx context.Context) (job models.Jobs, err error)
}

//LastUpdatedSlice is a type that implements the sort interface so that the jobs in it can be sorted using the generic Sort function.
type LastUpdatedSlice []models.Job

//Len is a function that's required by the sort interface.
func (s LastUpdatedSlice) Len() int {
	return len(s)
}

//Less is a function that's required by the sort interface.
func (s LastUpdatedSlice) Less(i, j int) bool {
	return s[i].LastUpdated.Before(s[j].LastUpdated)
}

//Swap is a function that's required by the sort interface.
func (s LastUpdatedSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// jobs collection name
const jobsCol = "jobs"

// locked jobs collection name
const jobsLockCol = "jobs_locks"

//MgoDataStore is a type that contains an implementation of the MgoJobStore interface, which can be used for creating and getting Job resources.
//It also represents a simplistic MongoDB configuration, with session, health and lock clients
type MgoDataStore struct {
	Jobs         MgoJobStore
	Session      *mgo.Session
	URI          string
	Database     string
	Collection   string
	client       *dpMongoHealth.Client
	healthClient *dpMongoHealth.CheckMongoClient
	lockClient   *dpMongoLock.Lock
}

func (m *MgoDataStore) CreateJob(ctx context.Context, id string) (job models.Job, err error) {
	log.Event(ctx, "creating job in mongo DB", log.Data{"id": id})

	// If an empty id was passed in, return an error with a message.
	if id == "" {
		return models.Job{}, errors.New("id must not be an empty string")
	}

	//Create a Job that's populated with default values of all its attributes
	newJob := models.NewJob(id)

	s := m.Session.Copy()
	defer s.Close()
	var jobToFind models.Job

	//Check that the jobs collection does not already contain the id as a key
	err = s.DB(m.Database).C(jobsCol).Find(bson.M{"id": id}).One(&jobToFind)
	if err != nil {
		if err == mgo.ErrNotFound {
			//this means we CAN insert the job as it does not already exist
			err = s.DB(m.Database).C(m.Collection).Insert(newJob)
			if err != nil {
				return models.Job{}, errors.New("error inserting job into mongo DB")
			}
			log.Event(ctx, "adding job to jobs collection", log.Data{"Job details: ": newJob})
		} else {
			//an unexpected error has occurred
			return models.Job{}, err
		}
	} else {
		//no error means that it found a job already exists with the id we're trying to insert
		return models.Job{}, errors.New("id must be unique")
	}

	return newJob, nil
}

// Init creates a new mgo.Session with a strong consistency and a write mode of "majority".
func (m *MgoDataStore) Init(ctx context.Context) (err error) {
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
	databaseCollectionBuilder[(dpMongoHealth.Database)(m.Database)] = []dpMongoHealth.Collection{(dpMongoHealth.Collection)(m.Collection), (dpMongoHealth.Collection)(jobsLockCol)}
	// Create client and healthclient from session
	m.client = dpMongoHealth.NewClientWithCollections(m.Session, databaseCollectionBuilder)
	m.healthClient = &dpMongoHealth.CheckMongoClient{
		Client:      *m.client,
		Healthcheck: m.client.Healthcheck,
	}

	// Create MongoDB lock client, which also starts the purger loop
	m.lockClient = dpMongoLock.New(ctx, m.Session, m.Database, jobsCol)
	return nil
}

//AcquireJobLock tries to lock the provided jobID.
//If the job is already locked, this function will block until it's released,
//at which point we acquire the lock and return.
func (m *MgoDataStore) AcquireJobLock(ctx context.Context, jobID string) (lockID string, err error) {
	return m.lockClient.Acquire(ctx, jobID)
}

// UnlockJob releases an exclusive mongoDB lock for the provided lockId (if it exists)
func (m *MgoDataStore) UnlockJob(lockID string) error {
	return m.lockClient.Unlock(lockID)
}

// Close closes the mongo session and returns any error
func (m *MgoDataStore) Close(ctx context.Context) error {
	m.lockClient.Close(ctx)
	return dpMongodb.Close(ctx, m.Session)
}

func (m *MgoDataStore) GetJobs(ctx context.Context) (models.Jobs, error) {
	s := m.Session.Copy()
	defer s.Close()
	log.Event(ctx, "getting list of jobs", log.INFO)

	results := models.Jobs{}
	numJobs, _ := s.DB(m.Database).C(jobsCol).Count()
	log.Event(ctx, "number of jobs found in jobs collection", log.Data{"numJobs": numJobs})

	if numJobs == 0 {
		log.Event(ctx, "there are no jobs in the job store - so the list is empty", log.INFO)
		return results, nil
	}

	//need to get all the jobs from the jobs collection and order them by last_updated
	iter := s.DB(m.Database).C(jobsCol).Find(bson.M{}).Iter()
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

	//Use a LastUpdatedSlice to put the jobs in last_updated order (ascending).
	jobsToSort := make(LastUpdatedSlice, 0, numJobs)
	for k := range jobs {
		jobsToSort = append(jobsToSort, jobs[k])
	}
	sort.Sort(jobsToSort)
	results.JobList = jobsToSort
	log.Event(ctx, "list of jobs - sorted by last_updated", log.Data{"Sorted jobs: ": results.JobList}, log.INFO)

	return results, nil
}

func (m *MgoDataStore) GetJob(ctx context.Context, id string) (models.Job, error) {
	s := m.Session.Copy()
	defer s.Close()
	log.Event(ctx, "getting job by ID", log.Data{"id": id})

	// If an empty id was passed in, return an error with a message.
	if id == "" {
		return models.Job{}, errors.New("id must not be an empty string")
	}

	var job models.Job
	err := s.DB(m.Database).C(jobsCol).Find(bson.M{"id": id}).One(&job)
	if err != nil {
		if err == mgo.ErrNotFound {
			return models.Job{}, errors.New("the jobs collection does not contain the job id entered")
		}
		return models.Job{}, err
	}

	return job, nil
}
