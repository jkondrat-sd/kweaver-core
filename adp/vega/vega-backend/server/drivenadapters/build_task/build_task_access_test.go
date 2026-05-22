package build_task

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

func MockNewBuildTaskAccess(t *testing.T) (*buildTaskAccess, sqlmock.Sqlmock) {
	db, smock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	bta := &buildTaskAccess{
		db: db,
	}
	return bta, smock
}

func mockBuildTaskRows() *sqlmock.Rows {
	return sqlmock.NewRows(buildTaskColumns()).AddRow(
		"task-1",
		"resource-1",
		"catalog-1",
		interfaces.BuildTaskStatusRunning,
		interfaces.BuildTaskModeBatch,
		int64(100),
		int64(40),
		int64(30),
		"mark-1",
		"",
		"user-1",
		interfaces.ACCESSOR_TYPE_USER,
		int64(1000),
		"app-1",
		interfaces.ACCESSOR_TYPE_APP,
		int64(2000),
		"title,content",
		"id",
		"embedding-model",
		1536,
	)
}

func testBuildTask() *interfaces.BuildTask {
	return &interfaces.BuildTask{
		ID:              "task-1",
		ResourceID:      "resource-1",
		CatalogID:       "catalog-1",
		Status:          interfaces.BuildTaskStatusRunning,
		Mode:            interfaces.BuildTaskModeBatch,
		TotalCount:      100,
		SyncedCount:     40,
		VectorizedCount: 30,
		SyncedMark:      "mark-1",
		ErrorMsg:        "",
		Creator:         interfaces.AccountInfo{ID: "user-1", Type: interfaces.ACCESSOR_TYPE_USER},
		CreateTime:      1000,
		Updater:         interfaces.AccountInfo{ID: "app-1", Type: interfaces.ACCESSOR_TYPE_APP},
		UpdateTime:      2000,
		EmbeddingFields: "title,content",
		BuildKeyFields:  "id",
		EmbeddingModel:  "embedding-model",
		ModelDimensions: 1536,
	}
}

type fakeBuildTaskScanner struct {
	values []any
	err    error
}

func (s fakeBuildTaskScanner) Scan(dest ...any) error {
	if s.err != nil {
		return s.err
	}
	for i, value := range s.values {
		switch d := dest[i].(type) {
		case *string:
			*d = value.(string)
		case *int:
			*d = value.(int)
		case *int64:
			*d = value.(int64)
		}
	}
	return nil
}

func TestBuildTaskColumns(t *testing.T) {
	Convey("Test buildTaskColumns", t, func() {
		So(buildTaskColumns(), ShouldResemble, []string{
			"f_id",
			"f_resource_id",
			"f_catalog_id",
			"f_status",
			"f_mode",
			"f_total_count",
			"f_synced_count",
			"f_vectorized_count",
			"f_synced_mark",
			"f_error_msg",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_updater",
			"f_updater_type",
			"f_update_time",
			"f_embedding_fields",
			"f_build_key_fields",
			"f_embedding_model",
			"f_model_dimensions",
		})
	})
}

func TestScanBuildTask(t *testing.T) {
	Convey("Test scanBuildTask", t, func() {
		Convey("Should scan build task", func() {
			scanner := fakeBuildTaskScanner{values: []any{
				"task-1",
				"resource-1",
				"catalog-1",
				interfaces.BuildTaskStatusRunning,
				interfaces.BuildTaskModeBatch,
				int64(100),
				int64(40),
				int64(30),
				"mark-1",
				"",
				"user-1",
				interfaces.ACCESSOR_TYPE_USER,
				int64(1000),
				"app-1",
				interfaces.ACCESSOR_TYPE_APP,
				int64(2000),
				"title,content",
				"id",
				"embedding-model",
				1536,
			}}

			task, err := scanBuildTask(scanner)
			So(err, ShouldBeNil)
			So(task.ID, ShouldEqual, "task-1")
			So(task.ResourceID, ShouldEqual, "resource-1")
			So(task.CatalogID, ShouldEqual, "catalog-1")
			So(task.Status, ShouldEqual, interfaces.BuildTaskStatusRunning)
			So(task.Creator, ShouldResemble, interfaces.AccountInfo{ID: "user-1", Type: interfaces.ACCESSOR_TYPE_USER})
			So(task.Updater, ShouldResemble, interfaces.AccountInfo{ID: "app-1", Type: interfaces.ACCESSOR_TYPE_APP})
			So(task.ModelDimensions, ShouldEqual, 1536)
		})

		Convey("Should return sql no rows error", func() {
			task, err := scanBuildTask(fakeBuildTaskScanner{err: sql.ErrNoRows})
			So(err, ShouldEqual, sql.ErrNoRows)
			So(task, ShouldBeNil)
		})

		Convey("Should return generic scan error", func() {
			expectedErr := errors.New("scan failed")
			task, err := scanBuildTask(fakeBuildTaskScanner{err: expectedErr})
			So(err, ShouldEqual, expectedErr)
			So(task, ShouldBeNil)
		})
	})
}

func Test_BuildTaskAccess_Create(t *testing.T) {
	Convey("test Create\n", t, func() {
		bta, smock := MockNewBuildTaskAccess(t)
		task := testBuildTask()

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id,f_resource_id,f_catalog_id,f_status,f_mode,f_total_count,f_synced_count,"+
			"f_vectorized_count,f_synced_mark,f_error_msg,f_creator,f_creator_type,f_create_time,f_updater,f_updater_type,"+
			"f_update_time,f_embedding_fields,f_build_key_fields,f_embedding_model,f_model_dimensions) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", BUILD_TASK_TABLE_NAME)

		Convey("Create Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(
				task.ID,
				task.ResourceID,
				task.CatalogID,
				task.Status,
				task.Mode,
				task.TotalCount,
				task.SyncedCount,
				task.VectorizedCount,
				task.SyncedMark,
				task.ErrorMsg,
				task.Creator.ID,
				task.Creator.Type,
				task.CreateTime,
				task.Updater.ID,
				task.Updater.Type,
				task.UpdateTime,
				task.EmbeddingFields,
				task.BuildKeyFields,
				task.EmbeddingModel,
				task.ModelDimensions,
			).WillReturnResult(sqlmock.NewResult(1, 1))

			err := bta.Create(context.Background(), task)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create Exec sql error\n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs(
				task.ID,
				task.ResourceID,
				task.CatalogID,
				task.Status,
				task.Mode,
				task.TotalCount,
				task.SyncedCount,
				task.VectorizedCount,
				task.SyncedMark,
				task.ErrorMsg,
				task.Creator.ID,
				task.Creator.Type,
				task.CreateTime,
				task.Updater.ID,
				task.Updater.Type,
				task.UpdateTime,
				task.EmbeddingFields,
				task.BuildKeyFields,
				task.EmbeddingModel,
				task.ModelDimensions,
			).WillReturnError(expectedErr)

			err := bta.Create(context.Background(), task)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_BuildTaskAccess_GetByID(t *testing.T) {
	Convey("test GetByID\n", t, func() {
		bta, smock := MockNewBuildTaskAccess(t)
		sqlStr := fmt.Sprintf("SELECT f_id, f_resource_id, f_catalog_id, f_status, f_mode, f_total_count, f_synced_count, "+
			"f_vectorized_count, f_synced_mark, f_error_msg, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, "+
			"f_update_time, f_embedding_fields, f_build_key_fields, f_embedding_model, f_model_dimensions FROM %s WHERE f_id = ?", BUILD_TASK_TABLE_NAME)

		Convey("GetByID Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs("task-1").WillReturnRows(mockBuildTaskRows())

			task, err := bta.GetByID(context.Background(), "task-1")
			So(err, ShouldBeNil)
			So(task.ID, ShouldEqual, "task-1")
			So(task.Creator, ShouldResemble, interfaces.AccountInfo{ID: "user-1", Type: interfaces.ACCESSOR_TYPE_USER})

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetByID Success no row\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs("missing").WillReturnError(sql.ErrNoRows)

			task, err := bta.GetByID(context.Background(), "missing")
			So(err, ShouldBeNil)
			So(task, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetByID Failed\n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs("task-1").WillReturnError(expectedErr)

			task, err := bta.GetByID(context.Background(), "task-1")
			So(task, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_BuildTaskAccess_GetStatus(t *testing.T) {
	Convey("test GetStatus\n", t, func() {
		bta, smock := MockNewBuildTaskAccess(t)
		sqlStr := fmt.Sprintf("SELECT f_status FROM %s WHERE f_id = ?", BUILD_TASK_TABLE_NAME)

		Convey("GetStatus Success\n", func() {
			rows := sqlmock.NewRows([]string{"f_status"}).AddRow(interfaces.BuildTaskStatusRunning)
			smock.ExpectQuery(sqlStr).WithArgs("task-1").WillReturnRows(rows)

			status, err := bta.GetStatus(context.Background(), "task-1")
			So(err, ShouldBeNil)
			So(status, ShouldEqual, interfaces.BuildTaskStatusRunning)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetStatus Success no row\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs("missing").WillReturnError(sql.ErrNoRows)

			status, err := bta.GetStatus(context.Background(), "missing")
			So(status, ShouldEqual, "")
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_BuildTaskAccess_Delete(t *testing.T) {
	Convey("test Delete\n", t, func() {
		bta, smock := MockNewBuildTaskAccess(t)
		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_id = ?", BUILD_TASK_TABLE_NAME)

		Convey("Delete Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs("task-1").WillReturnResult(sqlmock.NewResult(0, 1))

			err := bta.Delete(context.Background(), "task-1")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete not found\n", func() {
			smock.ExpectExec(sqlStr).WithArgs("missing").WillReturnResult(sqlmock.NewResult(0, 0))

			err := bta.Delete(context.Background(), "missing")
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
