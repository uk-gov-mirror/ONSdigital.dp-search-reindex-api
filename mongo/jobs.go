package mongo

import (
	"context"
	"errors"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// AcquireJobLock tries to lock the provided jobID.
// If the job is already locked, this function will block until it's released,
// at which point we acquire the lock and return.
func (m *JobStore) AcquireJobLock(ctx context.Context, jobID string) (lockID string, err error) {
	return m.LockClient.Acquire(ctx, jobID)
}

// UnlockJob releases an exclusive mongoDB lock for the provided lockId (if it exists)
func (m *JobStore) UnlockJob(ctx context.Context, lockID string) {
	m.LockClient.Unlock(ctx, lockID)
	log.Info(ctx, "job lockID has unlocked successfully")
}

// CheckInProgressJob checks if a new reindex job can be created depending on if a reindex job is currently in progress between cfg.MaxReindexJobRuntime before now and now
// It returns an error if a reindex job already exists which is in_progress state
func (m *JobStore) CheckInProgressJob(ctx context.Context, cfg *config.Config) error {
	log.Info(ctx, "checking if an existing reindex job is in progress")

	// checkFromTime is the time of configured variable "MaxReindexJobRuntime" from now
	checkFromTime := time.Now().Add(-1 * cfg.MaxReindexJobRuntime)

	var job models.Job

	// get job with state "in-progress" and its "reindex_started" time is between cfg.MaxReindexJobRuntime before now and now
	err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollection)).
		FindOne(ctx, bson.M{
			"state":           "in-progress",
			"reindex_started": bson.M{"$gte": checkFromTime, "$lte": time.Now()},
		}, &job)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Info(ctx, "no reindex jobs with state in progress")
			return nil
		}
		log.Error(ctx, "failed to get job with state in-progress and the most recent reindex_started time", err)
		return err
	}

	// found an existing reindex job in progress
	if (job != models.Job{}) {
		logData := log.Data{
			"id":              job.ID,
			"state":           job.State,
			"reindex_started": job.ReindexStarted,
		}
		log.Info(ctx, "found an existing reindex job in progress", logData)
		return ErrExistingJobInProgress
	}

	return nil
}

// CreateJob creates a new job, with the given search index name, in the collection, and assigns default values to its attributes
func (m *JobStore) CreateJob(ctx context.Context, job models.Job) error {
	logData := log.Data{"job": job}
	log.Info(ctx, "creating reindex job in mongo DB", logData)

	// insert job into mongoDB
	_, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollection)).Insert(ctx, job)
	if err != nil {
		log.Error(ctx, "failed to insert job into mongo DB", err, logData)
		return err
	}

	return nil
}

func (m *JobStore) findJob(ctx context.Context, jobID string) (*models.Job, error) {
	var job models.Job

	err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollection)).FindOne(ctx, bson.M{"_id": jobID}, &job)
	if err != nil {
		logData := log.Data{
			"database":         m.Database,
			"jobs_collections": config.JobsCollection,
			"job_id":           jobID,
		}

		log.Error(ctx, "failed to find job in mongo", err, logData)
		return nil, err
	}

	return &job, nil
}

// GetJob retrieves the details of a particular job, from the collection, specified by its id
func (m *JobStore) GetJob(ctx context.Context, id string) (*models.Job, error) {
	logData := log.Data{"id": id}

	log.Info(ctx, "getting job by ID", logData)

	job, err := m.findJob(ctx, id)
	if err != nil {
		if err == mgo.ErrNotFound {
			log.Error(ctx, "job not found in mongo", err, logData)
			return nil, ErrJobNotFound
		}
		log.Error(ctx, "failed to find job in mongo", err, logData)
		return nil, err
	}

	return job, nil
}

// GetJobs retrieves all the jobs, from the collection, and lists them in order of last_updated
func (m *JobStore) GetJobs(ctx context.Context, option Options) (*models.Jobs, error) {
	logData := log.Data{"option": option}
	log.Info(ctx, "getting list of jobs", logData)

	// get jobs count
	numJobs, err := m.getJobsCount(ctx)
	if err != nil {
		log.Error(ctx, "failed to get jobs count", err, logData)
		return nil, err
	}

	// create and populate jobsList
	jobsList := make([]models.Job, numJobs)
	_, err = m.Connection.Collection(m.ActualCollectionName(config.JobsCollection)).Find(ctx, bson.M{}, &jobsList,
		mongodriver.Sort(bson.M{"last_updated": 1}), mongodriver.Offset(option.Offset), mongodriver.Limit(option.Limit))
	if err != nil {
		log.Error(ctx, "failed to populate jobs list", err, logData)
		return nil, err
	}

	jobs := &models.Jobs{
		Count:      len(jobsList),
		JobList:    jobsList,
		Limit:      option.Limit,
		Offset:     option.Offset,
		TotalCount: numJobs,
	}

	return jobs, nil
}

// getJobsCount returns the total number of jobs stored in the jobs collection in mongo
func (m *JobStore) getJobsCount(ctx context.Context) (int, error) {
	logData := log.Data{}

	numJobs, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollection)).Count(ctx, bson.M{})
	if err != nil {
		logData["database"] = m.Database
		logData["jobs_collection"] = config.JobsCollection

		log.Error(ctx, "error counting jobs", err, logData)
		return 0, err
	}

	logData["no_of_jobs"] = numJobs
	log.Info(ctx, "number of jobs found in jobs collection", logData)

	return numJobs, nil
}

// UpdateJob updates a particular job with the values passed in through the 'updates' input parameter
func (m *JobStore) UpdateJob(ctx context.Context, id string, updates bson.M) error {
	update := bson.M{"$set": updates}

	if _, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollection)).Must().Update(ctx, bson.M{"_id": id}, update); err != nil {
		logData := log.Data{
			"job_id":  id,
			"updates": updates,
		}

		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "job not found", err, logData)
			return ErrJobNotFound
		}

		log.Error(ctx, "failed to update job in mongo", err, logData)
		return err
	}

	return nil
}
