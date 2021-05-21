package mock

import (
	"context"
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
	GetJobsFunc func(ctx context.Context, collectionID string) (models.Jobs, error)

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
			// CollectionID is the collectionID argument value.
			CollectionID string
		}
		Close []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
}

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
func (mock *MgoJobStoreMock) GetJob(ctx context.Context, id string) (job models.Job, err error) {
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
func (mock *MgoJobStoreMock) GetJobs(ctx context.Context, collectionID string) (job models.Jobs, err error) {
	if mock.GetJobsFunc == nil {
		panic("JobStoreMock.GetJobFunc: method is nil but GetJob was just called")
	}
	callInfo := struct {
		Ctx context.Context
		CollectionID string
	}{
		Ctx: ctx,
		CollectionID: collectionID,
	}
	mock.calls.GetJobs = append(mock.calls.GetJobs, callInfo)
	return mock.GetJobsFunc(ctx, collectionID)
}
