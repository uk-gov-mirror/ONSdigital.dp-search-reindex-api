package api

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-search-reindex-api/mongo"
	"github.com/ONSdigital/dp-search-reindex-api/store"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"net/http/httptest"
)

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {

		api := Setup(context.Background(), mux.NewRouter(), &store.DataStore{}, mongo.MgoDataStore{})

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(api.Router, "/jobs", "POST"), ShouldBeTrue)
			So(hasRoute(api.Router, "/jobs/{id}", "GET"), ShouldBeTrue)
			So(hasRoute(api.Router, "/jobs", "GET"), ShouldBeTrue)
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, nil)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}
