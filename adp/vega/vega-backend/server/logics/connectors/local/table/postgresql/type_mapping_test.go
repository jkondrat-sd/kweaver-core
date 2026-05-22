// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package postgresql

import (
	"testing"

	"vega-backend/interfaces"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPostgresqlMapType(t *testing.T) {
	Convey("Test Postgresql MapType", t, func() {
		c := &PostgresqlConnector{}

		Convey("Known int4 scalar maps to integer", func() {
			So(c.MapType("int4"), ShouldEqual, interfaces.DataType_Integer)
		})

		Convey("Known INT4 scalar maps to integer", func() {
			So(c.MapType("INT4"), ShouldEqual, interfaces.DataType_Integer)
		})

		Convey("Trimmed text scalar maps to text", func() {
			So(c.MapType("  text  "), ShouldEqual, interfaces.DataType_Text)
		})

		Convey("Known jsonb scalar maps to json", func() {
			So(c.MapType("jsonb"), ShouldEqual, interfaces.DataType_Json)
		})

		Convey("UDT int array is not recognized", func() {
			So(c.MapType("_int4"), ShouldEqual, interfaces.DataType_Other)
		})

		Convey("UDT text array is not recognized", func() {
			So(c.MapType("_text"), ShouldEqual, interfaces.DataType_Other)
		})

		Convey("Data_type array is not recognized", func() {
			So(c.MapType("ARRAY"), ShouldEqual, interfaces.DataType_Other)
		})

		Convey("Unknown type maps to other", func() {
			So(c.MapType("unknown_type"), ShouldEqual, interfaces.DataType_Other)
		})

		Convey("Empty type maps to other", func() {
			So(c.MapType(""), ShouldEqual, interfaces.DataType_Other)
		})
	})
}
