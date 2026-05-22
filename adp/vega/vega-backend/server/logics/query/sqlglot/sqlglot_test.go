// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package sqlglot

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestMapDataSourceTypeToDialect_MySQL(t *testing.T) {
	Convey("Test MapDataSourceTypeToDialect maps mysql", t, func() {
		dialect, err := MapDataSourceTypeToDialect(interfaces.ConnectorTypeMySQL)
		So(err, ShouldBeNil)
		So(dialect, ShouldEqual, "mysql")
	})
}

func TestMapDataSourceTypeToDialect_PostgresAlias(t *testing.T) {
	Convey("Test MapDataSourceTypeToDialect maps postgres alias", t, func() {
		dialect, err := MapDataSourceTypeToDialect("postgres")
		So(err, ShouldBeNil)
		So(dialect, ShouldEqual, "postgres")
	})
}

func TestMapDataSourceTypeToDialect_MariaDB(t *testing.T) {
	Convey("Test MapDataSourceTypeToDialect maps mariadb to mysql dialect", t, func() {
		dialect, err := MapDataSourceTypeToDialect(interfaces.ConnectorTypeMariaDB)
		So(err, ShouldBeNil)
		So(dialect, ShouldEqual, "mysql")
	})
}

func TestMapDataSourceTypeToDialect_Unsupported(t *testing.T) {
	Convey("Test MapDataSourceTypeToDialect rejects unsupported type", t, func() {
		dialect, err := MapDataSourceTypeToDialect("oracle")
		So(dialect, ShouldBeBlank)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "unsupported dataSourceType")
	})
}
