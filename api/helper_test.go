package api_test

import (
	"testing"

	"github.com/ONSdigital/dp-search-reindex-api/api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetValueType(t *testing.T) {
	Convey("Given value is a string type", t, func() {
		value := "test"

		Convey("When GetValueType is called", func() {
			valType := api.GetValueType(value)

			Convey("Then type `string` is returned", func() {
				So(valType, ShouldEqual, "string")
			})
		})
	})

	Convey("Given value is a boolean type", t, func() {
		value := true

		Convey("When GetValueType is called", func() {
			valType := api.GetValueType(value)

			Convey("Then type `boolean` is returned", func() {
				So(valType, ShouldEqual, "boolean")
			})
		})
	})

	Convey("Given value is a float32 type", t, func() {
		value := float32(1)

		Convey("When GetValueType is called", func() {
			valType := api.GetValueType(value)

			Convey("Then type `integer` is returned", func() {
				So(valType, ShouldEqual, "integer")
			})
		})
	})

	Convey("Given value is a float64 type", t, func() {
		value := float64(1)

		Convey("When GetValueType is called", func() {
			valType := api.GetValueType(value)

			Convey("Then type `integer` is returned", func() {
				So(valType, ShouldEqual, "integer")
			})
		})
	})

	Convey("Given value is an int type", t, func() {
		value := 1

		Convey("When GetValueType is called", func() {
			valType := api.GetValueType(value)

			Convey("Then type `integer` is returned", func() {
				So(valType, ShouldEqual, "integer")
			})
		})
	})

	Convey("Given value is an int8 type", t, func() {
		value := int8(1)

		Convey("When GetValueType is called", func() {
			valType := api.GetValueType(value)

			Convey("Then type `integer` is returned", func() {
				So(valType, ShouldEqual, "integer")
			})
		})
	})

	Convey("Given value is an int16 type", t, func() {
		value := int16(1)

		Convey("When GetValueType is called", func() {
			valType := api.GetValueType(value)

			Convey("Then type `integer` is returned", func() {
				So(valType, ShouldEqual, "integer")
			})
		})
	})

	Convey("Given value is an int32 type", t, func() {
		value := int32(1)

		Convey("When GetValueType is called", func() {
			valType := api.GetValueType(value)

			Convey("Then type `integer` is returned", func() {
				So(valType, ShouldEqual, "integer")
			})
		})
	})

	Convey("Given value is an int64 type", t, func() {
		value := int64(1)

		Convey("When GetValueType is called", func() {
			valType := api.GetValueType(value)

			Convey("Then type `integer` is returned", func() {
				So(valType, ShouldEqual, "integer")
			})
		})
	})
}
