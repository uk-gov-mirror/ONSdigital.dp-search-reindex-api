package pagination_test

import (
	"testing"

	"github.com/ONSdigital/dp-search-reindex-api/pagination"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidatePaginationParametersReturnsErrorWhenOffsetIsNegative(t *testing.T) {

	Convey("Given a minus offset value passed in and a JobStore containing 20 jobs", t, func() {

		offset := "-1"
		limit := ""
		totalCount := 20

		Convey("When ReadPaginationValues is called", func() {

			paginator := pagination.NewPaginator(20, 0, 1000)
			offset, limit, err := paginator.ValidatePaginationParameters(offset, limit, totalCount)

			Convey("Then the expected error is returned", func() {

				So(err, ShouldEqual, pagination.ErrInvalidOffsetParameter)
				So(offset, ShouldBeZeroValue)
				So(limit, ShouldBeZeroValue)
			})
		})
	})
}

//func TestReadPaginationParametersReturnsErrorWhenLimitIsNegative(t *testing.T) {
//
//	Convey("Given a request with a minus limit value", t, func() {
//
//		r := httptest.NewRequest("GET", "/test?limit=-1", nil)
//
//		Convey("When ReadPaginationValues is called", func() {
//
//			paginator := pagination.Paginator{}
//			offset, limit, err := paginator.ReadPaginationParameters(r)
//
//			Convey("Then the expected error is returned", func() {
//				So(err, ShouldEqual, pagination.ErrInvalidLimitParameter)
//				So(offset, ShouldBeZeroValue)
//				So(limit, ShouldBeZeroValue)
//			})
//		})
//	})
//}

//func TestReadPaginationParametersReturnsErrorWhenLimitIsGreaterThanMaxLimit(t *testing.T) {
//
//	Convey("Given a request with a limit value over the maximum", t, func() {
//
//		r := httptest.NewRequest("GET", "/test?limit=1001", nil)
//
//		Convey("When ReadPaginationValues is called", func() {
//
//			paginator := pagination.Paginator{DefaultMaxLimit: 1000}
//			offset, limit, err := paginator.ReadPaginationParameters(r)
//
//			Convey("Then the expected error is returned", func() {
//				So(err, ShouldEqual, pagination.ErrLimitOverMax)
//				So(offset, ShouldBeZeroValue)
//				So(limit, ShouldBeZeroValue)
//			})
//		})
//	})
//}

//func TestReadPaginationParametersReturnsLimitAndOffsetProvidedFromQuery(t *testing.T) {
//
//	Convey("Given a request with a valid limit and offset", t, func() {
//
//		r := httptest.NewRequest("GET", "/test?limit=10&offset=5", nil)
//
//		Convey("When ReadPaginationValues is called", func() {
//
//			paginator := pagination.Paginator{DefaultMaxLimit: 1000}
//			offset, limit, err := paginator.ReadPaginationParameters(r)
//
//			Convey("Then the expected values are returned", func() {
//				So(err, ShouldBeNil)
//				So(offset, ShouldEqual, 5)
//				So(limit, ShouldEqual, 10)
//			})
//		})
//	})
//}

//func TestReadPaginationParametersReturnsDefaultValuesWhenNotProvided(t *testing.T) {
//
//	Convey("Given a request without pagination parameters", t, func() {
//
//		r := httptest.NewRequest("GET", "/test", nil)
//
//		Convey("When ReadPaginationValues is called", func() {
//			expectedLimit := 20
//			expectedOffset := 1
//			paginator := pagination.Paginator{DefaultLimit: expectedLimit, DefaultOffset: expectedOffset, DefaultMaxLimit: 1000}
//			offset, limit, err := paginator.ReadPaginationParameters(r)
//
//			Convey("Then the configured default values are returned", func() {
//				So(err, ShouldBeNil)
//				So(offset, ShouldEqual, expectedOffset)
//				So(limit, ShouldEqual, expectedLimit)
//			})
//		})
//	})
//}

func TestNewPaginatorReturnsPaginatorStructWithFilledValues(t *testing.T) {

	Convey("Given a set of expected paginator values", t, func() {

		expectedPaginator := &pagination.Paginator{
			DefaultLimit:    10,
			DefaultOffset:   5,
			DefaultMaxLimit: 100,
		}

		Convey("When NewPaginator is called", func() {

			actualPaginator := pagination.NewPaginator(10, 5, 100)

			Convey("Then the paginator is configured as expected", func() {
				So(actualPaginator, ShouldResemble, expectedPaginator)
			})
		})
	})
}
