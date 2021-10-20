package api_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-search-reindex-api/api"
	"github.com/ONSdigital/dp-search-reindex-api/api/mock"
	"github.com/ONSdigital/dp-search-reindex-api/config"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"net/http/httptest"
)

func TestSetup(t *testing.T) {
	cfg, err := config.Get()
	if err != nil {
		t.Fatalf("failed to get config: %s", err)
	}
	validTaskNames := strings.Split(cfg.TaskNameValues, ",")

	//create map of valid task name values
	taskNameValues := make(map[string]int)
	for t, taskName := range validTaskNames {
		taskNameValues[taskName] = t
	}

	Convey("Given an API instance", t, func() {
		api := api.Setup(context.Background(), mux.NewRouter(), &mock.DataStorerMock{}, &mock.AuthHandlerMock{}, taskNameValues)

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(api.Router, "/jobs", "POST"), ShouldBeTrue)
			So(hasRoute(api.Router, "/jobs/{id}", "GET"), ShouldBeTrue)
			So(hasRoute(api.Router, "/jobs", "GET"), ShouldBeTrue)
		})
	})
}

func TestClose(t *testing.T) {
	cfg, err := config.Get()
	if err != nil {
		t.Fatalf("failed to get config: %s", err)
	}
	validTaskNames := strings.Split(cfg.TaskNameValues, ",")

	//create map of valid task name values
	taskNameValues := make(map[string]int)
	for t, taskName := range validTaskNames {
		taskNameValues[taskName] = t
	}

	Convey("Given an API instance", t, func() {
		ctx := context.Background()
		api := api.Setup(context.Background(), mux.NewRouter(), &mock.DataStorerMock{}, &mock.AuthHandlerMock{}, taskNameValues)

		Convey("When the api is closed then there is no error returned", func() {
			err := api.Close(ctx)
			So(err, ShouldBeNil)
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, nil)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}
