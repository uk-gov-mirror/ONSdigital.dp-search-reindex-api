package store

import (
	"context"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"testing"
	"time"
	. "github.com/smartystreets/goconvey/convey"
)

// Constants for testing
const (
	testJobID1 = "UUID1"
)

var ctx = context.Background()

func TestCreateJob(t *testing.T) {
	Convey("Successfully return without any errors", t, func() {
		Convey("when the job has the id field for POST request", func() {
			inputID := testJobID1

			jobStorer := JobStorer{}
			job, err := jobStorer.CreateJob(ctx, inputID)
			So(err, ShouldBeNil)
			expectedJob := expectedJob()
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
		})
	})
}

func expectedJob() models.Job {
	return models.Job{
		ID:          testJobID1,
		LastUpdated: time.Now().UTC(),
		Links: &models.JobLinks{
			Tasks: "http://localhost:12150/jobs/" + testJobID1 + "/tasks",
			Self:  "http://localhost:12150/jobs/" + testJobID1,
		},
		NumberOfTasks:                0,
		ReindexCompleted:             time.Time{}.UTC(),
		ReindexFailed:                time.Time{}.UTC(),
		ReindexStarted:               time.Time{}.UTC(),
		SearchIndexName:              "Default Search Index Name",
		State:                        "created",
		TotalSearchDocuments:         0,
		TotalInsertedSearchDocuments: 0,
	}
}
