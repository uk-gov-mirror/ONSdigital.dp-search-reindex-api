package store

import (
	"context"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// Constants for testing
const (
	testJobID1 = "UUID1"
	testJobID2 = "UUID2"
	testJobID3 = "UUID3"
)

var ctx = context.Background()

//TestCreateJob tests the CreateJob function in job_store.
func TestCreateJob(t *testing.T) {
	Convey("Successfully return without any errors", t, func() {
		Convey("when the job id is unique and is not an empty string", func() {
			inputID := testJobID1

			jobStore := DataStore{}
			job, err := jobStore.CreateJob(ctx, inputID)
			So(err, ShouldBeNil)
			expectedJob := models.NewJob(testJobID1)
			So(job.ID, ShouldEqual, expectedJob.ID)
			So(job.Links, ShouldResemble, expectedJob.Links)
			So(job.NumberOfTasks, ShouldEqual, expectedJob.NumberOfTasks)
			So(job.ReindexCompleted, ShouldEqual, expectedJob.ReindexCompleted)
			So(job.ReindexFailed, ShouldEqual, expectedJob.ReindexFailed)
			So(job.ReindexStarted, ShouldEqual, expectedJob.ReindexStarted)
			So(job.SearchIndexName, ShouldEqual, expectedJob.SearchIndexName)
			So(job.State, ShouldEqual, expectedJob.State)
			So(job.TotalSearchDocuments, ShouldEqual, expectedJob.TotalSearchDocuments)
			So(job.TotalInsertedSearchDocuments, ShouldEqual, expectedJob.TotalInsertedSearchDocuments)

			Convey("Return with error when the same job id is used a second time", func() {
				expectedErrorMsg := "id must be unique"
				job, err := jobStore.CreateJob(ctx, inputID)

				So(err, ShouldNotBeNil)
				So(job, ShouldResemble, models.Job{})
				So(err.Error(), ShouldEqual, expectedErrorMsg)
			})
		})
	})
	Convey("Return with error when the job id is an empty string", t, func() {
		inputID := ""
		expectedErrorMsg := "id must not be an empty string"
		jobStore := DataStore{}
		job, err := jobStore.CreateJob(ctx, inputID)

		So(job, ShouldResemble, models.Job{})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, expectedErrorMsg)
	})
}

//TestGetJob tests the GetJob function in job_store.
func TestGetJob(t *testing.T) {
	Convey("Successfully return without any errors", t, func() {
		Convey("when the job id exists in the jobStore", func() {
			inputID := testJobID2
			jobStore := DataStore{}

			//first create a job so that it exists in the job store
			job, err := jobStore.CreateJob(ctx, inputID)
			So(err, ShouldBeNil)

			//check that the job created is not an empty job
			So(job, ShouldNotResemble, models.Job{})

			//then get the job and check that it contains the same values as the one that's just been created
			job_returned, err := jobStore.GetJob(ctx, inputID)
			So(err, ShouldBeNil)
			So(job.ID, ShouldEqual, job_returned.ID)
			So(job.Links, ShouldResemble, job_returned.Links)
			So(job.NumberOfTasks, ShouldEqual, job_returned.NumberOfTasks)
			So(job.ReindexCompleted, ShouldEqual, job_returned.ReindexCompleted)
			So(job.ReindexFailed, ShouldEqual, job_returned.ReindexFailed)
			So(job.ReindexStarted, ShouldEqual, job_returned.ReindexStarted)
			So(job.SearchIndexName, ShouldEqual, job_returned.SearchIndexName)
			So(job.State, ShouldEqual, job_returned.State)
			So(job.TotalSearchDocuments, ShouldEqual, job_returned.TotalSearchDocuments)
			So(job.TotalInsertedSearchDocuments, ShouldEqual, job_returned.TotalInsertedSearchDocuments)
		})
	})
	Convey("Return with error when the job id does not exist in the jobStore", t, func() {
		inputID := testJobID3
		expectedErrorMsg := "the job store does not contain the job id entered"
		jobStore := DataStore{}
		job, err := jobStore.GetJob(ctx, inputID)

		So(err, ShouldNotBeNil)
		So(job, ShouldResemble, models.Job{})
		So(err.Error(), ShouldEqual, expectedErrorMsg)
	})
}
