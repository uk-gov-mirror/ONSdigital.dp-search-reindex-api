package url_test

import (
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-search-reindex-api/url"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	websiteURL = "localhost:20000"
	jobID      = "123"
	nameOfApi  = "zebedee"
)

func TestBuilder_BuildWebsiteDatasetVersionURL(t *testing.T) {

	Convey("Given a URL builder", t, func() {

		urlBuilder := url.NewBuilder(websiteURL)

		Convey("When BuildJobURL is called", func() {

			url := urlBuilder.BuildJobURL(jobID)

			expectedURL := fmt.Sprintf("%s/jobs/%s",
				websiteURL, jobID)

			Convey("Then the expected URL is returned", func() {
				So(url, ShouldEqual, expectedURL)
			})
		})

		Convey("When BuildJobTasksURL is called", func() {

			url := urlBuilder.BuildJobTasksURL(jobID)

			expectedURL := fmt.Sprintf("%s/jobs/%s/tasks",
				websiteURL, jobID)

			Convey("Then the expected URL is returned", func() {
				So(url, ShouldEqual, expectedURL)
			})
		})

		Convey("When BuildJobTaskURL is called", func() {

			url := urlBuilder.BuildJobTaskURL(jobID, nameOfApi)

			expectedURL := fmt.Sprintf("%s/jobs/%s/tasks/%s",
				websiteURL, jobID, nameOfApi)

			Convey("Then the expected URL is returned", func() {
				So(url, ShouldEqual, expectedURL)
			})
		})
	})
}
