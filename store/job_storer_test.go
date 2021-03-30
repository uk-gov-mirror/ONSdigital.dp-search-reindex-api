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
)

var ctx = context.Background()

//TestCreateJob tests the CreateJob function in job_storer.
func TestCreateJob(t *testing.T) {
	Convey("Successfully return without any errors", t, func() {
		Convey("when the job id is unique and is not an empty string", func() {
			inputID := testJobID1

			jobStorer := JobStorer{}
			job, err := jobStorer.CreateJob(ctx, inputID)
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

				job, err := jobStorer.CreateJob(ctx, inputID)
				So(job, ShouldResemble, models.Job{})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, expectedErrorMsg)
			})
		})
	})
	Convey("Return with error when the job id is an empty string", t, func() {
		inputID := ""
		expectedErrorMsg := "id must not be an empty string"

		jobStorer := JobStorer{}
		job, err := jobStorer.CreateJob(ctx, inputID)
		So(job, ShouldResemble, models.Job{})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, expectedErrorMsg)
	})
}
