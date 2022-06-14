package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewJob(t *testing.T) {
	Convey("Given a search index name for a new job", t, func() {
		ctx := context.Background()
		searchIndex := "testSearchIndexName"

		currentTime := time.Now().UTC()
		zeroTime := time.Time{}.UTC()

		Convey("When NewJob is called", func() {
			job, err := NewJob(ctx, searchIndex)

			Convey("Then a new job resource should be created and returned", func() {
				So(job, ShouldNotBeEmpty)
				So(job.ID, ShouldNotBeEmpty)
				So(job.ETag, ShouldNotBeEmpty)

				// check LastUpdated to a degree of seconds
				So(job.LastUpdated.Day(), ShouldEqual, currentTime.Day())
				So(job.LastUpdated.Month().String(), ShouldEqual, currentTime.Month().String())
				So(job.LastUpdated.Year(), ShouldEqual, currentTime.Year())
				So(job.LastUpdated.Hour(), ShouldEqual, currentTime.Hour())
				So(job.LastUpdated.Minute(), ShouldEqual, currentTime.Minute())
				So(job.LastUpdated.Second(), ShouldEqual, currentTime.Second())

				selfLink := fmt.Sprintf("/jobs/%s", job.ID)
				So(job.Links.Self, ShouldEqual, selfLink)
				So(job.Links.Tasks, ShouldEqual, selfLink+"/tasks")

				So(job.NumberOfTasks, ShouldBeZeroValue)
				So(job.ReindexCompleted, ShouldEqual, zeroTime)
				So(job.ReindexFailed, ShouldEqual, zeroTime)
				So(job.ReindexStarted, ShouldEqual, zeroTime)
				So(job.SearchIndexName, ShouldEqual, searchIndex)
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
