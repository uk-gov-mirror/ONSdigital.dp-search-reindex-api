package mongo

import (
	"context"
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
func (m *JobStore) UnlockJob(ctx context.Context, lockID string) {
	m.lockClient.Unlock(lockID)
	log.Info(ctx, "job lockID has unlocked successfully")
}

// CheckNewReindexCanBeCreated checks if a new reindex job can be created depending on if a reindex job is currently in progress between cfg.MaxReindexJobRuntime before now and now
// It returns an error if a reindex job already exists which is in_progress state
func (m *JobStore) CheckNewReindexCanBeCreated(ctx context.Context) error {
	log.Info(ctx, "checking if a new reindex job can be created")

	s := m.Session.Copy()
	defer s.Close()

	// get all the jobs with state "in-progress" and order them by "-reindex_started" starting with most recent
	iter := s.DB(m.Database).C(m.JobsCollection).Find(bson.M{"state": "in-progress"}).Sort("-reindex_started").Iter()

	// checkFromTime is the time of configured variable "MaxReindexJobRuntime" from now
	checkFromTime := time.Now().Add(-1 * m.cfg.MaxReindexJobRuntime)
	result := models.Job{}

	for iter.Next(&result) {
		jobStartTime := result.ReindexStarted

		// check if start time of the job in progress is later than the checkFromTime but earlier than now
		if jobStartTime.After(checkFromTime) && jobStartTime.Before(time.Now()) {
			logData := log.Data{
				"id":                   result.ID,
				"state":                result.State,
				"reindex_started_time": jobStartTime,
			}
			log.Info(ctx, "found an existing reindex job in progress", logData)
			return ErrExistingJobInProgress
		}
	}

	defer func() {
		if err := iter.Close(); err != nil {
			log.Error(ctx, "error closing iterator", err)
		}
	}()

	return nil
}

// CreateJob creates a new job, with the given search index name, in the collection, and assigns default values to its attributes
func (m *JobStore) CreateJob(ctx context.Context, searchIndexName string) (*models.Job, error) {
	logData := log.Data{"search_index_name": searchIndexName}
	log.Info(ctx, "creating reindex job in mongo DB", logData)

	s := m.Session.Copy()
	defer s.Close()

	// create a new job
	newJob, err := models.NewJob(searchIndexName)
	if err != nil {
		log.Error(ctx, "failed to create new job", err, logData)
		return nil, err
	}

	// checks if there exists a reindex job with the same id as the new job
	err = m.validateJobIDUnique(newJob.ID)
	if err != nil {
		logData["jobID"] = newJob.ID
		log.Error(ctx, "failed to validate job id is unique", err, logData)
		return nil, err
	}

	// insert new job into mongoDB
	err = s.DB(m.Database).C(m.JobsCollection).Insert(*newJob)
	if err != nil {
		logData["new_job"] = newJob
		log.Error(ctx, "failed to insert new job into mongo DB", err, logData)
		return nil, err
	}

	return newJob, err
}

// validateJobIDUnique checks if there exists a reindex job in mongo with the given id
func (m *JobStore) validateJobIDUnique(id string) error {
	s := m.Session.Copy()
	defer s.Close()

	var jobToFind models.Job
	err := s.DB(m.Database).C(m.JobsCollection).Find(bson.M{"_id": id}).One(&jobToFind)
	if err != nil {
		if err == mgo.ErrNotFound {
			// success as none of the existing jobs have the given id
			return nil
		}
		return err
	}

	// if found then there exists an existing job with the given id
	return ErrDuplicateIDProvided
}

// GetJob retrieves the details of a particular job, from the collection, specified by its id
func (m *JobStore) GetJob(ctx context.Context, id string) (models.Job, error) {
	log.Info(ctx, "getting job by ID", log.Data{"id": id})

	s := m.Session.Copy()
	defer s.Close()

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
	log.Info(ctx, "getting list of jobs")

	s := m.Session.Copy()
	defer s.Close()

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

// UpdateJob updates a particular job with the values passed in through the 'updates' input parameter
func (m *JobStore) UpdateJob(ctx context.Context, id string, updates bson.M) error {
	s := m.Session.Copy()
	defer s.Close()

	update := bson.M{"$set": updates}

	if err := s.DB(m.Database).C(m.JobsCollection).UpdateId(id, update); err != nil {
		logData := log.Data{
			"job_id":  id,
			"updates": updates,
		}
		log.Error(ctx, "failed to update job in mongo", err, logData)

		if err == mgo.ErrNotFound {
			return ErrJobNotFound
		}
		return err
	}

	return nil
}
