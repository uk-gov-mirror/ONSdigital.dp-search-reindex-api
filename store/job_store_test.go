package store

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

// Constants for testing
const (
	testJobID1 = "UUID1"
	testJobID2 = "UUID2"
	testJobID3 = "UUID3"
	testJobID4 = "UUID4"
	testJobID5 = "UUID5"
)

var ctx = context.Background()
var jobStore = DataStore{}

//TestCreateJob tests the CreateJob function in job_store.
func TestCreateJob(t *testing.T) {

	t.Parallel()

	Convey("Successfully return without any errors", t, func() {
		Convey("when the job id is unique and is not an empty string", func() {
			inputID := testJobID1

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

			Convey("Return with error when the job id already exists", func() {
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
			inputID1 := testJobID4
			inputID2 := testJobID5

			//first create some jobs so that they exist in the job store
			job1, err := jobStore.CreateJob(ctx, inputID1)
			So(err, ShouldBeNil)
			job2, err := jobStore.CreateJob(ctx, inputID2)
			So(err, ShouldBeNil)

			//then get all the jobs from the jobStore and check that the newly created ones are amongst them
			jobs_returned, err := jobStore.GetJobs(ctx)
			So(err, ShouldBeNil)
			job_list_returned := jobs_returned.Job_List
			isJob1InList := contains(job_list_returned, job1)
			So(isJob1InList, ShouldEqual, true)
			isJob2InList := contains(job_list_returned, job2)
			So(isJob2InList, ShouldEqual, true)

			//then check that the first job in the list was last updated earlier than the second
			So(job_list_returned[0].LastUpdated, ShouldHappenBefore, job_list_returned[1].LastUpdated)
		})
		Convey("when the job store contains no jobs", func() {
			//first delete any jobs that exist in the job store
			err := jobStore.DeleteAllJobs(ctx)
			So(err, ShouldBeNil)

			//then get all the jobs from the jobStore and check that the returned list of jobs is empty
			jobs_returned, err := jobStore.GetJobs(ctx)
			So(err, ShouldBeNil)
			job_list_returned := jobs_returned.Job_List
			So(len(job_list_returned), ShouldEqual, 0)
		})
	})
}

//contains checks if a Job is present in a slice of Jobs
func contains(jobs_list []models.Job, job models.Job) bool {
	for _, v := range jobs_list {
		if v == job {
			return true
		}
	}

	return false
}
