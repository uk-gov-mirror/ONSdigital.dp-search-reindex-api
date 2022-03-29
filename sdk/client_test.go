package sdk

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	authHeader = "Authorization"
	authPrefix = "Bearer "
	eTagHeader = "ETag"
)

func TestHeaders_Add(t *testing.T) {
	t.Parallel()

	Convey("Given the sdk Headers struct contains a value for ETag", t, func() {
		req := &http.Request{
			Header: http.Header{},
		}
		headers := &Headers{
			ETag:             "dsalfhjsadf",
		}

		Convey("When calling the Add method on Headers", func() {
			headers.Add(req)

			Convey("Then an ETag header is set on the request", func() {
				So(req.Header.Get(eTagHeader), ShouldEqual, headers.ETag)
			})
		})
	})

	Convey("Given the sdk Headers struct contains avalue for If-Match", t, func() {
		req := &http.Request{}
		headers := &Headers{
			IfMatch: "*",
		}

		Convey("When calling the Add method on Headers", func() {
			headers.Add(req)

			Convey("Then an If-Match header is set on the request", func() {
				So(req.Header, ShouldBeEmpty)
			})
		})
	})

	Convey("Given the sdk Headers struct contains a value for Service Token ", t, func() {
		req := &http.Request{
			Header: http.Header{},
		}
		headers := &Headers{
			ServiceAuthToken: "4egqsf4378gfwqf44356h",
		}

		Convey("When calling the Add method on Headers", func() {
			headers.Add(req)

			expectedHeader := authPrefix + headers.ServiceAuthToken
			Convey("Then an Authorization header is set on the request", func() {
				So(req.Header, ShouldContainKey, authHeader)
				So(req.Header[authHeader], ShouldHaveLength, 1)
				So(req.Header[authHeader][0], ShouldEqual, expectedHeader)
			})
		})
	})
}
