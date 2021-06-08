package mock

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/mongo"
)

// Ensure, that MgoJobStoreMock does implement api.MgoJobStore.
// If this is not the case, regenerate this file with moq.
var _ mongo.MgoJobStore = &MgoJobStoreMock{}

type MgoJobStoreMock struct {
	// CreateJobFunc mocks the CreateJob method.
	CreateJobFunc func(ctx context.Context, id string) (models.Job, error)

	// GetJobFunc mocks the GetJob method.
	GetJobFunc func(ctx context.Context, id string) (models.Job, error)

	// GetJobsFunc mocks the GetJobs method.
	GetJobsFunc func(ctx context.Context) (models.Jobs, error)

	// CloseFunc mocks the Close method.
	CloseFunc func(ctx context.Context) error

	// calls tracks calls to the methods.
	calls struct {
		// CreateJob holds details about calls to the CreateJob method.
		CreateJob []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ID is the id argument value.
			ID string
		}
		// GetJob holds details about calls to the GetJob method.
		GetJob []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ID is the id argument value.
			ID string
		}
		// GetJobs holds details about calls to the GetJobs method.
		GetJobs []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		Close []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
}

// Constants for testing
const (
	notFoundID        = "NOT_FOUND_UUID"
	duplicateID       = "DUPLICATE_UUID"
	jobUpdatedFirstID = "JOB_UPDATED_FIRST_ID"
	jobUpdatedLastID  = "JOB_UPDATED_LAST_ID"
)

func (mock *MgoJobStoreMock) Close(ctx context.Context) error {
	if mock.CloseFunc == nil {
		panic("MgoJobStoreMock.CloseFunc: method is nil but Close was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.calls.Close = append(mock.calls.Close, callInfo)
	return mock.CloseFunc(ctx)
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//     len(mockedMongoServer.CloseCalls())
func (mock *MgoJobStoreMock) CloseCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	//lockMongoServerMockClose.RLock()
	calls = mock.calls.Close
	//lockMongoServerMockClose.RUnlock()
	return calls
}

// CreateJob calls CreateJobFunc.
func (mock *MgoJobStoreMock) CreateJob(ctx context.Context, id string) (job models.Job, err error) {
	if mock.CreateJobFunc == nil {
		if id == "" {
			return models.Job{}, errors.New("id must not be an empty string")
		} else if id == duplicateID {
			return models.Job{}, errors.New("id must be unique")
		} else {
			return models.NewJob(id), nil
		}
	}
	callInfo := struct {
		Ctx context.Context
		ID  string
	}{
		Ctx: ctx,
		ID:  id,
	}
	mock.calls.CreateJob = append(mock.calls.CreateJob, callInfo)
	return mock.CreateJobFunc(ctx, id)
}

// GetJob calls GetJobFunc.
func (mock *MgoJobStoreMock) GetJob(ctx context.Context, id string) (job models.Job, err error) {
	if mock.GetJobFunc == nil {
		if id == "" {
			return models.Job{}, errors.New("id must not be an empty string")
		} else if id == notFoundID {
			return models.Job{}, errors.New("the jobs collection does not contain the job id entered")
		} else {
			return models.NewJob(id), nil
		}
	}
	callInfo := struct {
		Ctx context.Context
		ID  string
	}{
		Ctx: ctx,
		ID:  id,
	}
	mock.calls.GetJob = append(mock.calls.GetJob, callInfo)
	return mock.GetJobFunc(ctx, id)
}

// GetJobs calls GetJobsFunc.
func (mock *MgoJobStoreMock) GetJobs(ctx context.Context) (job models.Jobs, err error) {
	if mock.GetJobsFunc == nil {
		results := models.Jobs{}
		jobs := make([]models.Job, 2)
		jobs[0] = models.NewJob(jobUpdatedFirstID)
		jobs[1] = models.NewJob(jobUpdatedLastID)
		results.JobList = jobs
		return results, nil
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.calls.GetJobs = append(mock.calls.GetJobs, callInfo)
	return mock.GetJobsFunc(ctx)
}
