package mongo

import (
	"context"
	"errors"
	"time"

	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// AcquireJobLock tries to lock the provided jobID.
// If the job is already locked, this function will block until it's released,
// at which point we acquire the lock and return.
func (m *JobStore) AcquireJobLock(ctx context.Context, jobID string) (lockID string, err error) {
	traceID := dprequest.GetRequestId(ctx)
	return m.lockClient.Acquire(ctx, jobID, traceID)
}

// UnlockJob releases an exclusive mongoDB lock for the provided lockId (if it exists)
func (m *JobStore) UnlockJob(lockID string) {
	m.lockClient.Unlock(lockID)
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

// GetJobs retrieves all the jobs, from the collection, and lists them in order of last_updated
func (m *JobStore) GetJobs(ctx context.Context, option Options) (models.Jobs, error) {
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

	jobsQuery := s.DB(m.Database).C(m.JobsCollection).Find(bson.M{}).Skip(option.Offset).Limit(option.Limit).Sort("last_updated")
	jobs := make([]models.Job, numJobs)
	if err := jobsQuery.All(&jobs); err != nil {
		return results, err
	}

	results.JobList = jobs
	results.Count = len(jobs)
	results.Limit = option.Limit
	results.Offset = option.Offset
	results.TotalCount = numJobs
	log.Info(ctx, "list of jobs - sorted by last_updated", log.Data{"Sorted jobs: ": results.JobList})

	return results, nil
}

// UpdateIndexName updates the search_index_name, of the relevant Job Resource, with the indexName provided
func (m *JobStore) UpdateIndexName(indexName, jobID string) error {
	s := m.Session.Copy()
	defer s.Close()
	updates := make(bson.M)
	updates["search_index_name"] = indexName
	updates["last_updated"] = time.Now()
	err := m.UpdateJob(updates, s, jobID)
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

// UpdateJobState updates the state of the relevant Job Resource with the one provided
func (m *JobStore) UpdateJobState(state, jobID string) error {
	s := m.Session.Copy()
	defer s.Close()
	updates := make(bson.M)
	updates["state"] = state
	updates["last_updated"] = time.Now()
	err := m.UpdateJob(updates, s, jobID)
	return err
}
