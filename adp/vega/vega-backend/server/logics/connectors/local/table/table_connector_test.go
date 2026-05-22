// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package table

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConvertValue_BytesToString(t *testing.T) {
	Convey("Test convertValue converts bytes to string", t, func() {
		So(convertValue([]byte("hello")), ShouldEqual, "hello")
	})
}

func TestConvertValue_KeepsOtherValues(t *testing.T) {
	Convey("Test convertValue keeps other values", t, func() {
		So(convertValue(123), ShouldEqual, 123)
	})
}

func TestScanRows_Success(t *testing.T) {
	Convey("Test ScanRows success", t, func() {
		db, mock, err := sqlmock.New()
		So(err, ShouldBeNil)
		defer func() { _ = db.Close() }()

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, []byte("Alice")).
			AddRow(2, []byte("Bob"))
		mock.ExpectQuery("SELECT id, name FROM users").WillReturnRows(rows)

		sqlRows, err := db.Query("SELECT id, name FROM users")
		So(err, ShouldBeNil)
		defer func() { _ = sqlRows.Close() }()

		result, err := ScanRows(sqlRows)
		So(err, ShouldBeNil)
		So(result.Columns, ShouldResemble, []string{"id", "name"})
		So(result.Total, ShouldEqual, 2)
		So(result.Rows[0]["name"], ShouldEqual, "Alice")
		So(result.Rows[1]["name"], ShouldEqual, "Bob")
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}
