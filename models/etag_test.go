package models_test

import (
	"testing"
	"time"

	"github.com/ONSdigital/dp-search-reindex-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateETagForJobPatch(t *testing.T) {
	currentJob := getTestJob()

	Convey("Given an updated job", t, func() {
		updatedJob := currentJob
		updatedJob.State = "completed"

		Convey("When GenerateETagForJob is called", func() {
			newETag, err := models.GenerateETagForJob(updatedJob)

			Convey("Then a new eTag is created", func() {
				So(newETag, ShouldNotBeEmpty)

				Convey("And new eTag should not be the same as the existing eTag", func() {
					So(newETag, ShouldNotEqual, currentJob.ETag)

					Convey("And no errors should be returned", func() {
						So(err, ShouldBeNil)
					})
				})
			})
		})
	})

	Convey("Given an existing job with no new job updates", t, func() {
		Convey("When GenerateETagForJob is called", func() {
			newETag, err := models.GenerateETagForJob(currentJob)

			Convey("Then an eTag is returned", func() {
				So(newETag, ShouldNotBeEmpty)

				Convey("And the eTag should be the same as the existing eTag", func() {
					So(newETag, ShouldEqual, currentJob.ETag)

					Convey("And no errors should be returned", func() {
						So(err, ShouldBeNil)
					})
				})
			})
		})
	})
}

func getTestJob() models.Job {
	zeroTime := time.Time{}.UTC()

	return models.Job{
		ID:          "test1",
		ETag:        `"56b6890f1321590998d5fd8d293b620581ff3531"`,
		LastUpdated: time.Now().UTC(),
		Links: &models.JobLinks{
			Tasks: "http://localhost:25700/jobs/test1234/tasks",
			Self:  "http://localhost:25700/jobs/test1234",
		},
		NumberOfTasks:                0,
		ReindexCompleted:             zeroTime,
		ReindexFailed:                zeroTime,
		ReindexStarted:               zeroTime,
		SearchIndexName:              "Test Search Index Name",
		State:                        models.JobStateCreated,
		TotalSearchDocuments:         0,
		TotalInsertedSearchDocuments: 0,
	}
}

func TestGenerateETagForTask(t *testing.T) {
	currentTask  := getTestTask()

	Convey("Given an updated task", t, func() {
		updatedTask := currentTask
		updatedTask.TaskName = "zebedee"

		Convey("When GenerateETagForTask is called", func() {
			newETag, err := models.GenerateETagForTask(updatedTask)

			Convey("Then a new eTag is created", func() {
				So(newETag, ShouldNotBeEmpty)

				Convey("And new eTag should not be the same as the existing eTag", func() {
					So(newETag, ShouldNotEqual, currentTask.ETag)

					Convey("And no errors should be returned", func() {
						So(err, ShouldBeNil)
					})
				})
			})
		})
	})

	Convey("Given an existing task with no new updates", t, func() {
		Convey("When GenerateETagForTask is called", func() {
			newETag, err := models.GenerateETagForTask(currentTask)

			Convey("Then an eTag is returned", func() {
				So(newETag, ShouldNotBeEmpty)

				Convey("And the eTag should be the same as the existing eTag", func() {
					So(newETag, ShouldEqual, currentTask.ETag)

					Convey("And no errors should be returned", func() {
						So(err, ShouldBeNil)
					})
				})
			})
		})
	})
}

func getTestTask() models.Task {

	return models.Task{
		ETag:  `"c644f142e428485848c5272759bdd216b5d7560e"`,
		JobID: "task1234",
		TaskName: "task",
		NumberOfDocuments: 3,
	}
}

