package connector_type

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func MockNewConnectorTypeAccess(t *testing.T) (*connectorTypeAccess, sqlmock.Sqlmock) {
	db, smock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	cta := &connectorTypeAccess{
		db: db,
	}
	return cta, smock
}

func testConnectorType() *interfaces.ConnectorType {
	return &interfaces.ConnectorType{
		Type:        "mysql",
		Name:        "MySQL",
		Tags:        []string{"database", "source"},
		Description: "MySQL connector",
		Mode:        interfaces.ConnectorModeLocal,
		Category:    interfaces.ConnectorCategoryTable,
		Endpoint:    "http://connector",
		FieldConfig: map[string]interfaces.ConnectorFieldConfig{
			"host": {Name: "Host", Type: "string", Required: true},
		},
		Enabled: true,
	}
}

func mockConnectorTypeRows() *sqlmock.Rows {
	return sqlmock.NewRows(connectorTypeColumns()).AddRow(
		"mysql",
		"MySQL",
		`"database","source"`,
		"MySQL connector",
		interfaces.ConnectorModeLocal,
		interfaces.ConnectorCategoryTable,
		"http://connector",
		`{"host":{"name":"Host","type":"string","required":true}}`,
		true,
	)
}

type fakeConnectorTypeScanner struct {
	values []any
	err    error
}

func (s fakeConnectorTypeScanner) Scan(dest ...any) error {
	if s.err != nil {
		return s.err
	}
	for i, value := range s.values {
		switch d := dest[i].(type) {
		case *string:
			*d = value.(string)
		case *bool:
			*d = value.(bool)
		}
	}
	return nil
}

func TestConnectorTypeColumns(t *testing.T) {
	Convey("Test connectorTypeColumns", t, func() {
		So(connectorTypeColumns(), ShouldResemble, []string{
			"f_type",
			"f_name",
			"f_tags",
			"f_description",
			"f_mode",
			"f_category",
			"f_endpoint",
			"f_field_config",
			"f_enabled",
		})
	})
}

func TestScanConnectorType(t *testing.T) {
	Convey("Test scanConnectorType", t, func() {
		Convey("Should scan connector type with field config", func() {
			scanner := fakeConnectorTypeScanner{values: []any{
				"mysql",
				"MySQL",
				"database,source",
				"MySQL connector",
				interfaces.ConnectorModeLocal,
				interfaces.ConnectorCategoryTable,
				"http://connector",
				`{"host":{"name":"Host","type":"string","required":true}}`,
				true,
			}}

			ct, err := scanConnectorType(scanner)
			So(err, ShouldBeNil)
			So(ct.Type, ShouldEqual, "mysql")
			So(ct.Name, ShouldEqual, "MySQL")
			So(ct.Tags, ShouldResemble, []string{"database", "source"})
			So(ct.Description, ShouldEqual, "MySQL connector")
			So(ct.Mode, ShouldEqual, interfaces.ConnectorModeLocal)
			So(ct.Category, ShouldEqual, interfaces.ConnectorCategoryTable)
			So(ct.Endpoint, ShouldEqual, "http://connector")
			So(ct.FieldConfig["host"].Name, ShouldEqual, "Host")
			So(ct.FieldConfig["host"].Required, ShouldBeTrue)
			So(ct.Enabled, ShouldBeTrue)
		})

		Convey("Should scan connector type with empty field config", func() {
			scanner := fakeConnectorTypeScanner{values: []any{
				"api",
				"API",
				"",
				"API connector",
				interfaces.ConnectorModeRemote,
				interfaces.ConnectorCategoryAPI,
				"http://api",
				"",
				false,
			}}

			ct, err := scanConnectorType(scanner)
			So(err, ShouldBeNil)
			So(ct.Tags, ShouldBeEmpty)
			So(ct.FieldConfig, ShouldBeNil)
			So(ct.Enabled, ShouldBeFalse)
		})

		Convey("Should return scan error", func() {
			scanner := fakeConnectorTypeScanner{err: sql.ErrNoRows}

			ct, err := scanConnectorType(scanner)
			So(err, ShouldEqual, sql.ErrNoRows)
			So(ct, ShouldBeNil)
		})

		Convey("Should return error when field config JSON is invalid", func() {
			scanner := fakeConnectorTypeScanner{values: []any{
				"mysql",
				"MySQL",
				"",
				"MySQL connector",
				interfaces.ConnectorModeLocal,
				interfaces.ConnectorCategoryTable,
				"",
				"invalid json",
				true,
			}}

			ct, err := scanConnectorType(scanner)
			So(err, ShouldNotBeNil)
			So(ct, ShouldBeNil)
		})

		Convey("Should propagate generic scan error", func() {
			expectedErr := errors.New("scan failed")
			scanner := fakeConnectorTypeScanner{err: expectedErr}

			ct, err := scanConnectorType(scanner)
			So(err, ShouldEqual, expectedErr)
			So(ct, ShouldBeNil)
		})
	})
}

func Test_ConnectorTypeAccess_Create(t *testing.T) {
	Convey("test Create\n", t, func() {
		cta, smock := MockNewConnectorTypeAccess(t)
		ct := testConnectorType()
		sqlStr := fmt.Sprintf("INSERT INTO %s (f_type,f_name,f_tags,f_description,f_mode,f_category,f_endpoint,f_field_config,f_enabled) VALUES (?,?,?,?,?,?,?,?,?)", CONNECTOR_TYPE_TABLE_NAME)

		Convey("Create Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(
				ct.Type,
				ct.Name,
				`"database","source"`,
				ct.Description,
				ct.Mode,
				ct.Category,
				ct.Endpoint,
				sqlmock.AnyArg(),
				ct.Enabled,
			).WillReturnResult(sqlmock.NewResult(1, 1))

			err := cta.Create(context.Background(), ct)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create Exec sql error\n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs(
				ct.Type,
				ct.Name,
				`"database","source"`,
				ct.Description,
				ct.Mode,
				ct.Category,
				ct.Endpoint,
				sqlmock.AnyArg(),
				ct.Enabled,
			).WillReturnError(expectedErr)

			err := cta.Create(context.Background(), ct)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ConnectorTypeAccess_GetByType(t *testing.T) {
	Convey("test GetByType\n", t, func() {
		cta, smock := MockNewConnectorTypeAccess(t)
		sqlStr := fmt.Sprintf("SELECT f_type, f_name, f_tags, f_description, f_mode, f_category, f_endpoint, f_field_config, f_enabled FROM %s WHERE f_type = ?", CONNECTOR_TYPE_TABLE_NAME)

		Convey("GetByType Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs("mysql").WillReturnRows(mockConnectorTypeRows())

			ct, err := cta.GetByType(context.Background(), "mysql")
			So(err, ShouldBeNil)
			So(ct.Type, ShouldEqual, "mysql")
			So(ct.Tags, ShouldResemble, []string{"database", "source"})

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetByType Success no row\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs("missing").WillReturnError(sql.ErrNoRows)

			ct, err := cta.GetByType(context.Background(), "missing")
			So(err, ShouldBeNil)
			So(ct, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetByType Failed\n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs("mysql").WillReturnError(expectedErr)

			ct, err := cta.GetByType(context.Background(), "mysql")
			So(ct, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ConnectorTypeAccess_Update(t *testing.T) {
	Convey("test Update\n", t, func() {
		cta, smock := MockNewConnectorTypeAccess(t)
		ct := testConnectorType()
		sqlStr := fmt.Sprintf("UPDATE %s SET f_name = ?, f_tags = ?, f_description = ?, f_mode = ?, f_category = ?, f_endpoint = ?, f_field_config = ?, f_enabled = ? WHERE f_type = ?", CONNECTOR_TYPE_TABLE_NAME)

		Convey("Update Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(
				ct.Name,
				`"database","source"`,
				ct.Description,
				ct.Mode,
				ct.Category,
				ct.Endpoint,
				sqlmock.AnyArg(),
				ct.Enabled,
				ct.Type,
			).WillReturnResult(sqlmock.NewResult(0, 1))

			err := cta.Update(context.Background(), ct)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update Exec sql error\n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs(
				ct.Name,
				`"database","source"`,
				ct.Description,
				ct.Mode,
				ct.Category,
				ct.Endpoint,
				sqlmock.AnyArg(),
				ct.Enabled,
				ct.Type,
			).WillReturnError(expectedErr)

			err := cta.Update(context.Background(), ct)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ConnectorTypeAccess_DeleteByType(t *testing.T) {
	Convey("test DeleteByType\n", t, func() {
		cta, smock := MockNewConnectorTypeAccess(t)
		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_type = ?", CONNECTOR_TYPE_TABLE_NAME)

		Convey("DeleteByType Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs("mysql").WillReturnResult(sqlmock.NewResult(0, 1))

			err := cta.DeleteByType(context.Background(), "mysql")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteByType Failed\n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs("mysql").WillReturnError(expectedErr)

			err := cta.DeleteByType(context.Background(), "mysql")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ConnectorTypeAccess_SetEnabled(t *testing.T) {
	Convey("test SetEnabled\n", t, func() {
		cta, smock := MockNewConnectorTypeAccess(t)
		sqlStr := fmt.Sprintf("UPDATE %s SET f_enabled = ? WHERE f_type = ?", CONNECTOR_TYPE_TABLE_NAME)

		Convey("SetEnabled Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(false, "mysql").WillReturnResult(sqlmock.NewResult(0, 1))

			err := cta.SetEnabled(context.Background(), "mysql", false)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("SetEnabled Failed\n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs(false, "mysql").WillReturnError(expectedErr)

			err := cta.SetEnabled(context.Background(), "mysql", false)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
