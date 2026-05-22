package entityextension

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	. "github.com/smartystreets/goconvey/convey"
)

func MockNewStore(t *testing.T) (*Store, sqlmock.Sqlmock) {
	db, smock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return &Store{db: db}, smock
}

func TestFilterKeys(t *testing.T) {
	Convey("Test FilterKeys", t, func() {
		input := map[string]string{
			"env":   "prod",
			"owner": "data",
			"team":  "search",
		}

		Convey("Should return original map when keysCSV is empty", func() {
			out := FilterKeys(input, "")
			So(out, ShouldResemble, input)
		})

		Convey("Should trim keys and keep only allowed keys", func() {
			out := FilterKeys(input, " env, team ")
			So(out, ShouldResemble, map[string]string{
				"env":  "prod",
				"team": "search",
			})
		})

		Convey("Should return original map when parsed keys are empty", func() {
			out := FilterKeys(input, " , ")
			So(out, ShouldResemble, input)
		})
	})
}

func TestApplyJoinsForCatalog(t *testing.T) {
	Convey("Test ApplyJoinsForCatalog", t, func() {
		builder := sq.Select("t_catalog.f_id").From("t_catalog")

		sqlStr, args, err := ApplyJoinsForCatalog(
			builder,
			[]string{"env", "team"},
			[]string{"prod", "search"},
		).ToSql()

		So(err, ShouldBeNil)
		So(sqlStr, ShouldContainSubstring, "JOIN t_entity_extension vex0 ON vex0.f_entity_kind = ? AND vex0.f_entity_id = t_catalog.f_id AND vex0.f_key = ? AND vex0.f_value = ?")
		So(sqlStr, ShouldContainSubstring, "JOIN t_entity_extension vex1 ON vex1.f_entity_kind = ? AND vex1.f_entity_id = t_catalog.f_id AND vex1.f_key = ? AND vex1.f_value = ?")
		So(args, ShouldResemble, []any{KindCatalog, "env", "prod", KindCatalog, "team", "search"})
	})
}

func TestApplyJoinsForResource(t *testing.T) {
	Convey("Test ApplyJoinsForResource", t, func() {
		builder := sq.Select("t_resource.f_id").From("t_resource")

		sqlStr, args, err := ApplyJoinsForResource(
			builder,
			[]string{"env"},
			[]string{"prod"},
		).ToSql()

		So(err, ShouldBeNil)
		So(sqlStr, ShouldContainSubstring, "JOIN t_entity_extension vex0 ON vex0.f_entity_kind = ? AND vex0.f_entity_id = t_resource.f_id AND vex0.f_key = ? AND vex0.f_value = ?")
		So(args, ShouldResemble, []any{KindResource, "env", "prod"})
	})
}

func Test_Store_Replace(t *testing.T) {
	Convey("test Replace\n", t, func() {
		store, smock := MockNewStore(t)
		deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE f_entity_id = ? AND f_entity_kind = ?", tableName)
		insertSQL := fmt.Sprintf("INSERT INTO %s (f_entity_kind,f_entity_id,f_key,f_value,f_create_time,f_update_time) VALUES (?,?,?,?,?,?)", tableName)

		Convey("Replace Success\n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(deleteSQL).WithArgs("catalog-1", KindCatalog).WillReturnResult(sqlmock.NewResult(0, 1))
			smock.ExpectExec(insertSQL).WithArgs(KindCatalog, "catalog-1", "env", "prod", sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
			smock.ExpectCommit()

			err := store.Replace(context.Background(), KindCatalog, "catalog-1", map[string]string{"env": "prod"})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Replace delete error\n", func() {
			expectedErr := errors.New("delete error")
			smock.ExpectBegin()
			smock.ExpectExec(deleteSQL).WithArgs("catalog-1", KindCatalog).WillReturnError(expectedErr)
			smock.ExpectRollback()

			err := store.Replace(context.Background(), KindCatalog, "catalog-1", map[string]string{"env": "prod"})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Replace insert error\n", func() {
			expectedErr := errors.New("insert error")
			smock.ExpectBegin()
			smock.ExpectExec(deleteSQL).WithArgs("catalog-1", KindCatalog).WillReturnResult(sqlmock.NewResult(0, 1))
			smock.ExpectExec(insertSQL).WithArgs(KindCatalog, "catalog-1", "env", "prod", sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnError(expectedErr)
			smock.ExpectRollback()

			err := store.Replace(context.Background(), KindCatalog, "catalog-1", map[string]string{"env": "prod"})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_Store_DeleteByEntityIDs(t *testing.T) {
	Convey("test DeleteByEntityIDs\n", t, func() {
		store, smock := MockNewStore(t)
		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_entity_id IN (?,?) AND f_entity_kind = ?", tableName)

		Convey("DeleteByEntityIDs Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs("catalog-1", "catalog-2", KindCatalog).WillReturnResult(sqlmock.NewResult(0, 2))

			err := store.DeleteByEntityIDs(context.Background(), KindCatalog, []string{"catalog-1", "catalog-2"})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteByEntityIDs empty ids\n", func() {
			err := store.DeleteByEntityIDs(context.Background(), KindCatalog, nil)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteByEntityIDs failed\n", func() {
			expectedErr := errors.New("delete error")
			smock.ExpectExec(sqlStr).WithArgs("catalog-1", "catalog-2", KindCatalog).WillReturnError(expectedErr)

			err := store.DeleteByEntityIDs(context.Background(), KindCatalog, []string{"catalog-1", "catalog-2"})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_Store_GetByEntityID(t *testing.T) {
	Convey("test GetByEntityID\n", t, func() {
		store, smock := MockNewStore(t)
		sqlStr := fmt.Sprintf("SELECT f_key, f_value FROM %s WHERE f_entity_id = ? AND f_entity_kind = ? ORDER BY f_key", tableName)

		Convey("GetByEntityID Success\n", func() {
			rows := sqlmock.NewRows([]string{"f_key", "f_value"}).
				AddRow("env", "prod").
				AddRow("team", "search")
			smock.ExpectQuery(sqlStr).WithArgs("catalog-1", KindCatalog).WillReturnRows(rows)

			result, err := store.GetByEntityID(context.Background(), KindCatalog, "catalog-1")
			So(err, ShouldBeNil)
			So(result, ShouldResemble, map[string]string{"env": "prod", "team": "search"})

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetByEntityID query failed\n", func() {
			expectedErr := errors.New("query error")
			smock.ExpectQuery(sqlStr).WithArgs("catalog-1", KindCatalog).WillReturnError(expectedErr)

			result, err := store.GetByEntityID(context.Background(), KindCatalog, "catalog-1")
			So(result, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_Store_GetByEntityIDs(t *testing.T) {
	Convey("test GetByEntityIDs\n", t, func() {
		store, smock := MockNewStore(t)
		sqlStr := fmt.Sprintf("SELECT f_entity_id, f_key, f_value FROM %s WHERE f_entity_id IN (?,?) AND f_entity_kind = ? ORDER BY f_entity_id, f_key", tableName)

		Convey("GetByEntityIDs Success\n", func() {
			rows := sqlmock.NewRows([]string{"f_entity_id", "f_key", "f_value"}).
				AddRow("catalog-1", "env", "prod").
				AddRow("catalog-1", "team", "search").
				AddRow("catalog-2", "env", "dev")
			smock.ExpectQuery(sqlStr).WithArgs("catalog-1", "catalog-2", KindCatalog).WillReturnRows(rows)

			result, err := store.GetByEntityIDs(context.Background(), KindCatalog, []string{"catalog-1", "catalog-2"})
			So(err, ShouldBeNil)
			So(result, ShouldResemble, map[string]map[string]string{
				"catalog-1": {"env": "prod", "team": "search"},
				"catalog-2": {"env": "dev"},
			})

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetByEntityIDs empty ids\n", func() {
			result, err := store.GetByEntityIDs(context.Background(), KindCatalog, nil)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, map[string]map[string]string{})

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetByEntityIDs query failed\n", func() {
			expectedErr := errors.New("query error")
			smock.ExpectQuery(sqlStr).WithArgs("catalog-1", "catalog-2", KindCatalog).WillReturnError(expectedErr)

			result, err := store.GetByEntityIDs(context.Background(), KindCatalog, []string{"catalog-1", "catalog-2"})
			So(result, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
