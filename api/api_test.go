package api_test

import (
	"context"
	"net/http/httptest"
	"testing"

	dpHTTP "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-reindex-api/api"
	"github.com/ONSdigital/dp-search-reindex-api/api/mock"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var taskName1, taskName2 = "dataset-api", "zebedee"

var taskNames = map[string]bool{
	taskName1: true,
	taskName2: true,
}

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		httpClient := dpHTTP.NewClient()

		apiClient := api.Setup(mux.NewRouter(), &mock.DataStorerMock{}, &mock.AuthHandlerMock{}, taskNames, cfg, httpClient, &mock.IndexerMock{}, &mock.ReindexRequestedProducerMock{})

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(apiClient.Router, "/jobs", "POST"), ShouldBeTrue)
			So(hasRoute(apiClient.Router, "/jobs/{id}", "GET"), ShouldBeTrue)
			So(hasRoute(apiClient.Router, "/jobs", "GET"), ShouldBeTrue)
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Given an API instance", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		httpClient := dpHTTP.NewClient()
		ctx := context.Background()
		apiClient := api.Setup(mux.NewRouter(), &mock.DataStorerMock{}, &mock.AuthHandlerMock{}, taskNames, cfg, httpClient, &mock.IndexerMock{}, &mock.ReindexRequestedProducerMock{})

		Convey("When the apiClient is closed then there is no error returned", func() {
			err := apiClient.Close(ctx)
			So(err, ShouldBeNil)
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, nil)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}
