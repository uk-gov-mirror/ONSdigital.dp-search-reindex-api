package mongotest

import (
	"context"
	apiMock "github.com/ONSdigital/dp-search-reindex-api/api/mock"
	"testing"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

// Constants for testing
const (
	testJobID1  = "UUID1"
	testJobID2  = "UUID2"
	notFoundID  = "NOT_FOUND_UUID"
	duplicateID = "DUPLICATE_UUID"
)

var ctx = context.Background()

//var jobStore = MgoDataStore{}
var jobStore = &apiMock.MgoJobStoreMock{}

//TestCreateJob tests the CreateJob function in mgo_job_store.
func TestCreateJob(t *testing.T) {
	t.Parallel()
	Convey("Successfully return without any errors", t, func() {
		Convey("when the job id is unique and is not an empty string", func() {
			job, err := jobStore.CreateJob(ctx, testJobID1)
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

			Convey("Return with error when the job id already exists", func() {
				expectedErrorMsg := "id must be unique"
				job, err := jobStore.CreateJob(ctx, duplicateID)

				So(err, ShouldNotBeNil)
				So(job, ShouldResemble, models.Job{})
				So(err.Error(), ShouldEqual, expectedErrorMsg)
			})
		})
	})
	Convey("Return with error when the job id is an empty string", t, func() {
		inputID := ""
		expectedErrorMsg := "id must not be an empty string"
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

			//first create a job so that it exists in the job store
			job, err := jobStore.CreateJob(ctx, inputID)
			So(err, ShouldBeNil)

			//check that the job created is not an empty job
			So(job, ShouldNotResemble, models.Job{})

			//then get the job and check that it contains the same values as the one that's just been created
			jobReturned, err := jobStore.GetJob(ctx, inputID)
			So(err, ShouldBeNil)
			So(job.ID, ShouldEqual, jobReturned.ID)
			So(job.Links, ShouldResemble, jobReturned.Links)
			So(job.NumberOfTasks, ShouldEqual, jobReturned.NumberOfTasks)
			So(job.ReindexCompleted, ShouldEqual, jobReturned.ReindexCompleted)
			So(job.ReindexFailed, ShouldEqual, jobReturned.ReindexFailed)
			So(job.ReindexStarted, ShouldEqual, jobReturned.ReindexStarted)
			So(job.SearchIndexName, ShouldEqual, jobReturned.SearchIndexName)
			So(job.State, ShouldEqual, jobReturned.State)
			So(job.TotalSearchDocuments, ShouldEqual, jobReturned.TotalSearchDocuments)
			So(job.TotalInsertedSearchDocuments, ShouldEqual, jobReturned.TotalInsertedSearchDocuments)
		})
	})
	Convey("Return with error when the job id does not exist in the jobStore", t, func() {
		inputID := notFoundID
		expectedErrorMsg := "the jobs collection does not contain the job id entered"
		job, err := jobStore.GetJob(ctx, inputID)

		So(err, ShouldNotBeNil)
		So(job, ShouldResemble, models.Job{})
		So(err.Error(), ShouldEqual, expectedErrorMsg)
	})
	Convey("Return with error when the job id is an empty String", t, func() {
		inputID := ""
		expectedErrorMsg := "id must not be an empty string"
		job, err := jobStore.GetJob(ctx, inputID)

		So(err, ShouldNotBeNil)
		So(job, ShouldResemble, models.Job{})
		So(err.Error(), ShouldEqual, expectedErrorMsg)
	})
}

//TestGetJobs tests the GetJob function in job_store.
func TestGetJobs(t *testing.T) {
	Convey("Successfully return without any errors", t, func() {
		Convey("when the job store contains some jobs", func() {
			//get all the jobs from the jobStore
			jobsReturned, err := jobStore.GetJobs(ctx)
			So(err, ShouldBeNil)
			jobListReturned := jobsReturned.JobList

			//and check that the first job in the list was last updated earlier than the second
			So(jobListReturned[0].LastUpdated, ShouldHappenBefore, jobListReturned[1].LastUpdated)
		})
	})
}
