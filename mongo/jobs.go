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

// CheckInProgressJob checks if a new reindex job can be created depending on if a reindex job is currently in progress between cfg.MaxReindexJobRuntime before now and now
// It returns an error if a reindex job already exists which is in_progress state
func (m *JobStore) CheckInProgressJob(ctx context.Context) error {
	log.Info(ctx, "checking if an existing reindex job is in progress")

	s := m.Session.Copy()
	defer s.Close()

	// checkFromTime is the time of configured variable "MaxReindexJobRuntime" from now
	checkFromTime := time.Now().Add(-1 * m.cfg.MaxReindexJobRuntime)

	var job models.Job

	// get job with state "in-progress" and its "reindex_started" time is between cfg.MaxReindexJobRuntime before now and now
	err := s.DB(m.Database).C(m.JobsCollection).
		Find(bson.M{
			"state":           "in-progress",
			"reindex_started": bson.M{"$gte": checkFromTime, "$lte": time.Now()},
		}).One(&job)

	if err != nil {
		if err == mgo.ErrNotFound {
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

	s := m.Session.Copy()
	defer s.Close()

	// insert job into mongoDB
	err := s.DB(m.Database).C(m.JobsCollection).Insert(job)
	if err != nil {
		log.Error(ctx, "failed to insert job into mongo DB", err, logData)
		return err
	}

	return nil
}

func (m *JobStore) findJob(ctx context.Context, jobID string) (*models.Job, error) {
	s := m.Session.Copy()
	defer s.Close()

	var job models.Job

	err := s.DB(m.Database).C(m.JobsCollection).Find(bson.M{"_id": jobID}).One(&job)
	if err != nil {
		logData := log.Data{
			"database":         m.Database,
			"jobs_collections": m.JobsCollection,
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

	s := m.Session.Copy()
	defer s.Close()

	// get all jobs from mongo
	jobsQuery := s.DB(m.Database).C(m.JobsCollection).Find(bson.M{}).Skip(option.Offset).Limit(option.Limit).Sort("last_updated")

	// populate jobsList from jobsQuery
	jobsList := make([]models.Job, numJobs)
	if err := jobsQuery.All(&jobsList); err != nil {
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
	s := m.Session.Copy()
	defer s.Close()

	logData := log.Data{}

	numJobs, err := s.DB(m.Database).C(m.JobsCollection).Count()
	if err != nil {
		logData["database"] = m.Database
		logData["jobs_collection"] = m.JobsCollection

		log.Error(ctx, "error counting jobs", err, logData)
		return 0, err
	}

	logData["no_of_jobs"] = numJobs
	log.Info(ctx, "number of jobs found in jobs collection", logData)

	if numJobs == 0 {
		log.Info(ctx, "there are no jobs in the data store", logData)
	}

	return numJobs, nil
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

		if err == mgo.ErrNotFound {
			log.Error(ctx, "job not found", err, logData)
			return ErrJobNotFound
		}

		log.Error(ctx, "failed to update job in mongo", err, logData)
		return err
	}

	return nil
}
