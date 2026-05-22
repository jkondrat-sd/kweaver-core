package discover_task

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

func MockNewDiscoverTaskAccess(t *testing.T) (*discoverTaskAccess, sqlmock.Sqlmock) {
	db, smock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return &discoverTaskAccess{db: db}, smock
}

func mockDiscoverTaskRows() *sqlmock.Rows {
	return sqlmock.NewRows(discoverTaskColumns()).AddRow(
		"task-1",
		"catalog-1",
		"schedule-1",
		"full_sync",
		interfaces.DiscoverTaskTriggerScheduled,
		interfaces.DiscoverTaskStatusCompleted,
		100,
		"done",
		int64(1000),
		int64(2000),
		sql.NullString{String: `{}`, Valid: true},
		"user-1",
		interfaces.ACCESSOR_TYPE_USER,
		int64(900),
	)
}

func testDiscoverTask() *interfaces.DiscoverTask {
	return &interfaces.DiscoverTask{
		ID:          "task-1",
		CatalogID:   "catalog-1",
		ScheduleID:  "schedule-1",
		Strategy:    "full_sync",
		TriggerType: interfaces.DiscoverTaskTriggerScheduled,
		Status:      interfaces.DiscoverTaskStatusPending,
		Progress:    0,
		Message:     "",
		StartTime:   0,
		FinishTime:  0,
		Creator:     interfaces.AccountInfo{ID: "user-1", Type: interfaces.ACCESSOR_TYPE_USER},
		CreateTime:  900,
	}
}

type fakeDiscoverTaskScanner struct {
	values []any
	err    error
}

func (s fakeDiscoverTaskScanner) Scan(dest ...any) error {
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
		case *sql.NullString:
			*d = value.(sql.NullString)
		}
	}
	return nil
}

func TestDiscoverTaskColumns(t *testing.T) {
	Convey("Test discoverTaskColumns", t, func() {
		So(discoverTaskColumns(), ShouldResemble, []string{
			"f_id",
			"f_catalog_id",
			"f_schedule_id",
			"f_strategy",
			"f_trigger_type",
			"f_status",
			"f_progress",
			"f_message",
			"f_start_time",
			"f_finish_time",
			"f_result",
			"f_creator",
			"f_creator_type",
			"f_create_time",
		})
	})
}

func TestScanDiscoverTask(t *testing.T) {
	Convey("Test scanDiscoverTask", t, func() {
		Convey("Should scan task with result", func() {
			scanner := fakeDiscoverTaskScanner{values: []any{
				"task-1",
				"catalog-1",
				"schedule-1",
				"full_sync",
				interfaces.DiscoverTaskTriggerScheduled,
				interfaces.DiscoverTaskStatusCompleted,
				100,
				"done",
				int64(1000),
				int64(2000),
				sql.NullString{String: `{}`, Valid: true},
				"user-1",
				interfaces.ACCESSOR_TYPE_USER,
				int64(900),
			}}

			task, err := scanDiscoverTask(scanner)
			So(err, ShouldBeNil)
			So(task.ID, ShouldEqual, "task-1")
			So(task.CatalogID, ShouldEqual, "catalog-1")
			So(task.ScheduleID, ShouldEqual, "schedule-1")
			So(task.Progress, ShouldEqual, 100)
			So(task.Result, ShouldNotBeNil)
			So(task.Creator, ShouldResemble, interfaces.AccountInfo{ID: "user-1", Type: interfaces.ACCESSOR_TYPE_USER})
		})

		Convey("Should keep result nil when database value is null", func() {
			scanner := fakeDiscoverTaskScanner{values: []any{
				"task-1",
				"catalog-1",
				"",
				"full_sync",
				interfaces.DiscoverTaskTriggerManual,
				interfaces.DiscoverTaskStatusPending,
				0,
				"",
				int64(0),
				int64(0),
				sql.NullString{},
				"user-1",
				interfaces.ACCESSOR_TYPE_USER,
				int64(900),
			}}

			task, err := scanDiscoverTask(scanner)
			So(err, ShouldBeNil)
			So(task.Result, ShouldBeNil)
		})

		Convey("Should return scan error", func() {
			expectedErr := errors.New("scan failed")
			task, err := scanDiscoverTask(fakeDiscoverTaskScanner{err: expectedErr})
			So(err, ShouldEqual, expectedErr)
			So(task, ShouldBeNil)
		})
	})
}

func Test_DiscoverTaskAccess_GetScheduledTaskStrategy(t *testing.T) {
	Convey("test GetScheduledTaskStrategy\n", t, func() {
		dta, smock := MockNewDiscoverTaskAccess(t)
		sqlStr := "SELECT f_strategy FROM t_discover_schedule WHERE f_id = ?"

		Convey("GetScheduledTaskStrategy Success\n", func() {
			rows := sqlmock.NewRows([]string{"f_strategy"}).AddRow("full_sync")
			smock.ExpectQuery(sqlStr).WithArgs("schedule-1").WillReturnRows(rows)

			strategy, err := dta.GetScheduledTaskStrategy(context.Background(), "schedule-1")
			So(err, ShouldBeNil)
			So(strategy, ShouldEqual, "full_sync")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetScheduledTaskStrategy Success no row\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs("missing").WillReturnError(sql.ErrNoRows)

			strategy, err := dta.GetScheduledTaskStrategy(context.Background(), "missing")
			So(err, ShouldBeNil)
			So(strategy, ShouldEqual, "")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DiscoverTaskAccess_Create(t *testing.T) {
	Convey("test Create\n", t, func() {
		dta, smock := MockNewDiscoverTaskAccess(t)
		task := testDiscoverTask()
		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id,f_catalog_id,f_schedule_id,f_strategy,f_trigger_type,f_status,f_progress,f_message,f_start_time,f_finish_time,f_result,f_creator,f_creator_type,f_create_time) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)", DISCOVER_TASK_TABLE_NAME)

		Convey("Create Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(
				task.ID,
				task.CatalogID,
				task.ScheduleID,
				task.Strategy,
				task.TriggerType,
				task.Status,
				task.Progress,
				task.Message,
				task.StartTime,
				task.FinishTime,
				"",
				task.Creator.ID,
				task.Creator.Type,
				task.CreateTime,
			).WillReturnResult(sqlmock.NewResult(1, 1))

			err := dta.Create(context.Background(), task)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DiscoverTaskAccess_GetByID(t *testing.T) {
	Convey("test GetByID\n", t, func() {
		dta, smock := MockNewDiscoverTaskAccess(t)
		sqlStr := fmt.Sprintf("SELECT f_id, f_catalog_id, f_schedule_id, f_strategy, f_trigger_type, f_status, f_progress, f_message, f_start_time, f_finish_time, f_result, f_creator, f_creator_type, f_create_time FROM %s WHERE f_id = ?", DISCOVER_TASK_TABLE_NAME)

		Convey("GetByID Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs("task-1").WillReturnRows(mockDiscoverTaskRows())

			task, err := dta.GetByID(context.Background(), "task-1")
			So(err, ShouldBeNil)
			So(task.ID, ShouldEqual, "task-1")
			So(task.Result, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetByID Success no row\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs("missing").WillReturnError(sql.ErrNoRows)

			task, err := dta.GetByID(context.Background(), "missing")
			So(err, ShouldBeNil)
			So(task, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DiscoverTaskAccess_Delete(t *testing.T) {
	Convey("test Delete\n", t, func() {
		dta, smock := MockNewDiscoverTaskAccess(t)
		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_id = ?", DISCOVER_TASK_TABLE_NAME)

		Convey("Delete Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs("task-1").WillReturnResult(sqlmock.NewResult(0, 1))

			err := dta.Delete(context.Background(), "task-1")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete no row\n", func() {
			smock.ExpectExec(sqlStr).WithArgs("missing").WillReturnResult(sqlmock.NewResult(0, 0))

			err := dta.Delete(context.Background(), "missing")
			So(err, ShouldEqual, sql.ErrNoRows)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
