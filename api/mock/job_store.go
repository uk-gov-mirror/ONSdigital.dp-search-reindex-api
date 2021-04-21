package mock

import (
	"context"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/dp-search-reindex-api/store"
)

// Ensure, that JobStoreMock does implement api.JobStore.
// If this is not the case, regenerate this file with moq.
var _ store.JobStore = &JobStoreMock{}

type JobStoreMock struct {

	// CreateJobFunc mocks the CreateJob method.
	CreateJobFunc func(ctx context.Context, id string) (models.Job, error)

	// GetJobFunc mocks the GetJob method.
	GetJobFunc func(ctx context.Context, id string) (models.Job, error)

	// GetJobsFunc mocks the GetJobs method.
	GetJobsFunc func(ctx context.Context) (models.Jobs, error)

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
	}
}

// CreateJob calls CreateJobFunc.
func (mock *JobStoreMock) CreateJob(ctx context.Context, id string) (job models.Job, err error) {
	if mock.CreateJobFunc == nil {
		panic("JobStoreMock.CreateJobFunc: method is nil but CreateJob was just called")
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
func (mock *JobStoreMock) GetJob(ctx context.Context, id string) (job models.Job, err error) {
	if mock.GetJobFunc == nil {
		panic("JobStoreMock.GetJobFunc: method is nil but GetJob was just called")
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
func (mock *JobStoreMock) GetJobs(ctx context.Context) (job models.Jobs, err error) {
	if mock.GetJobsFunc == nil {
		panic("JobStoreMock.GetJobFunc: method is nil but GetJob was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.calls.GetJobs = append(mock.calls.GetJobs, callInfo)
	return mock.GetJobsFunc(ctx)
}
