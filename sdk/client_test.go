package sdk

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	authHeader    = "Authorization"
	authPrefix    = "Bearer "
	eTagHeader    = "ETag"
	ifMatchHeader = "If-Match"
)

func TestHeaders_Add(t *testing.T) {
	t.Parallel()

	req := &http.Request{
		Header: http.Header{},
	}

	Convey("Given the sdk Headers struct contains a value for ETag", t, func() {
		headers := &Headers{
			ETag: "dsalfhjsadf",
		}

		Convey("When calling the Add method on Headers", func() {
			headers.Add(req)

			Convey("Then an empty ETag header is set on the request", func() {
				So(req.Header, ShouldNotBeEmpty)
				So(req.Header.Get(eTagHeader), ShouldEqual, headers.ETag)
			})
		})
	})

	Convey("Given the sdk Headers struct contains a value for If-Match", t, func() {
		headers := &Headers{
			IfMatch: "*",
		}

		Convey("When calling the Add method on Headers", func() {
			headers.Add(req)

			Convey("Then an If-Match header is set on the request", func() {
				So(req.Header, ShouldNotBeEmpty)
				So(req.Header.Get(ifMatchHeader), ShouldEqual, headers.IfMatch)
			})
		})
	})

	Convey("Given the sdk Headers struct contains a value for Service Token ", t, func() {
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
