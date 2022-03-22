package models

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewJob(t *testing.T) {
	Convey("Given an id for a new job", t, func() {
		id := "test1234"

		currentTime := time.Now()
		zeroTime := time.Time{}.UTC()

		Convey("When NewJob is called", func() {
			job, err := NewJob(id)

			Convey("Then a new job resource should be created and returned", func() {
				So(job, ShouldNotBeEmpty)
				So(job.ID, ShouldEqual, id)

				// check LastUpdated to a degree of seconds
				So(job.LastUpdated.Day(), ShouldEqual, currentTime.Day())
				So(job.LastUpdated.Month().String(), ShouldEqual, currentTime.Month().String())
				So(job.LastUpdated.Year(), ShouldEqual, currentTime.Year())
				So(job.LastUpdated.Hour(), ShouldEqual, currentTime.Hour())
				So(job.LastUpdated.Minute(), ShouldEqual, currentTime.Minute())
				So(job.LastUpdated.Second(), ShouldEqual, currentTime.Second())

				So(job.Links.Tasks, ShouldEqual, "http://localhost:25700/jobs/test1234/tasks")
				So(job.Links.Self, ShouldEqual, "http://localhost:25700/jobs/test1234")

				So(job.NumberOfTasks, ShouldBeZeroValue)
				So(job.ReindexCompleted, ShouldEqual, zeroTime)
				So(job.ReindexFailed, ShouldEqual, zeroTime)
				So(job.ReindexStarted, ShouldEqual, zeroTime)
				So(job.SearchIndexName, ShouldEqual, "Default Search Index Name")
				So(job.State, ShouldEqual, JobStateCreated)
				So(job.TotalSearchDocuments, ShouldBeZeroValue)
				So(job.TotalInsertedSearchDocuments, ShouldBeZeroValue)

				Convey("And no errors returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})
}
