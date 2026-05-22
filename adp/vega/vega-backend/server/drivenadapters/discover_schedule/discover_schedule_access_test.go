package discover_schedule

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func MockNewDiscoverScheduleAccess(t *testing.T) (*discoverScheduleAccess, sqlmock.Sqlmock) {
	db, smock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return &discoverScheduleAccess{db: db}, smock
}

func mockDiscoverScheduleRows() *sqlmock.Rows {
	return sqlmock.NewRows(discoverScheduleColumns()).AddRow(
		"schedule-1",
		"daily discovery",
		"catalog-1",
		"0 0 * * *",
		int64(1000),
		int64(0),
		true,
		"full_sync",
		int64(2000),
		int64(3000),
		"user-1",
		interfaces.ACCESSOR_TYPE_USER,
		int64(900),
		"app-1",
		interfaces.ACCESSOR_TYPE_APP,
		int64(950),
	)
}

func testDiscoverSchedule() *interfaces.DiscoverSchedule {
	return &interfaces.DiscoverSchedule{
		ID:         "schedule-1",
		Name:       "daily discovery",
		CatalogID:  "catalog-1",
		CronExpr:   "0 0 * * *",
		StartTime:  1000,
		EndTime:    0,
		Enabled:    true,
		Strategy:   "full_sync",
		LastRun:    0,
		Creator:    interfaces.AccountInfo{ID: "user-1", Type: interfaces.ACCESSOR_TYPE_USER},
		CreateTime: 900,
		Updater:    interfaces.AccountInfo{ID: "app-1", Type: interfaces.ACCESSOR_TYPE_APP},
		UpdateTime: 950,
	}
}

type fakeDiscoverScheduleScanner struct {
	values []any
	err    error
}

func (s fakeDiscoverScheduleScanner) Scan(dest ...any) error {
	if s.err != nil {
		return s.err
	}
	for i, value := range s.values {
		switch d := dest[i].(type) {
		case *string:
			*d = value.(string)
		case *bool:
			*d = value.(bool)
		case *int64:
			*d = value.(int64)
		}
	}
	return nil
}

func TestDiscoverScheduleColumns(t *testing.T) {
	Convey("Test discoverScheduleColumns", t, func() {
		So(discoverScheduleColumns(), ShouldResemble, []string{
			"f_id",
			"f_name",
			"f_catalog_id",
			"f_cron_expr",
			"f_start_time",
			"f_end_time",
			"f_enabled",
			"f_strategy",
			"f_last_run",
			"f_next_run",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_updater",
			"f_updater_type",
			"f_update_time",
		})
	})
}

func TestScanDiscoverSchedule(t *testing.T) {
	Convey("Test scanDiscoverSchedule", t, func() {
		Convey("Should scan discover schedule", func() {
			scanner := fakeDiscoverScheduleScanner{values: []any{
				"schedule-1",
				"daily discovery",
				"catalog-1",
				"0 0 * * *",
				int64(1000),
				int64(0),
				true,
				"full_sync",
				int64(2000),
				int64(3000),
				"user-1",
				interfaces.ACCESSOR_TYPE_USER,
				int64(900),
				"app-1",
				interfaces.ACCESSOR_TYPE_APP,
				int64(950),
			}}

			schedule, err := scanDiscoverSchedule(scanner)
			So(err, ShouldBeNil)
			So(schedule.ID, ShouldEqual, "schedule-1")
			So(schedule.Name, ShouldEqual, "daily discovery")
			So(schedule.Enabled, ShouldBeTrue)
			So(schedule.Creator, ShouldResemble, interfaces.AccountInfo{ID: "user-1", Type: interfaces.ACCESSOR_TYPE_USER})
			So(schedule.Updater, ShouldResemble, interfaces.AccountInfo{ID: "app-1", Type: interfaces.ACCESSOR_TYPE_APP})
		})

		Convey("Should return sql no rows error", func() {
			schedule, err := scanDiscoverSchedule(fakeDiscoverScheduleScanner{err: sql.ErrNoRows})
			So(err, ShouldEqual, sql.ErrNoRows)
			So(schedule, ShouldBeNil)
		})

		Convey("Should return generic scan error", func() {
			expectedErr := errors.New("scan failed")
			schedule, err := scanDiscoverSchedule(fakeDiscoverScheduleScanner{err: expectedErr})
			So(err, ShouldEqual, expectedErr)
			So(schedule, ShouldBeNil)
		})
	})
}

func TestCalculateNextRun(t *testing.T) {
	Convey("Test calculateNextRun", t, func() {
		from := time.Date(2026, 5, 22, 10, 15, 0, 0, time.UTC)

		Convey("Should calculate next run for valid cron", func() {
			next, err := calculateNextRun("0 * * * *", from)
			So(err, ShouldBeNil)
			So(next, ShouldResemble, time.Date(2026, 5, 22, 11, 0, 0, 0, time.UTC))
		})

		Convey("Should return error for invalid cron", func() {
			next, err := calculateNextRun("invalid", from)
			So(err, ShouldNotBeNil)
			So(next.IsZero(), ShouldBeTrue)
		})
	})
}

func Test_DiscoverScheduleAccess_Create(t *testing.T) {
	Convey("test Create\n", t, func() {
		dsa, smock := MockNewDiscoverScheduleAccess(t)
		schedule := testDiscoverSchedule()
		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id,f_name,f_catalog_id,f_cron_expr,f_start_time,f_end_time,f_enabled,f_strategy,f_last_run,f_next_run,f_creator,f_creator_type,f_create_time) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)", DISCOVER_SCHEDULE_TABLE_NAME)

		Convey("Create Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(
				schedule.ID,
				schedule.Name,
				schedule.CatalogID,
				schedule.CronExpr,
				schedule.StartTime,
				schedule.EndTime,
				schedule.Enabled,
				schedule.Strategy,
				schedule.LastRun,
				sqlmock.AnyArg(),
				schedule.Creator.ID,
				schedule.Creator.Type,
				schedule.CreateTime,
			).WillReturnResult(sqlmock.NewResult(1, 1))

			err := dsa.Create(context.Background(), schedule)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create invalid cron\n", func() {
			schedule.CronExpr = "invalid"
			err := dsa.Create(context.Background(), schedule)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DiscoverScheduleAccess_GetByID(t *testing.T) {
	Convey("test GetByID\n", t, func() {
		dsa, smock := MockNewDiscoverScheduleAccess(t)
		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_catalog_id, f_cron_expr, f_start_time, f_end_time, f_enabled, f_strategy, f_last_run, f_next_run, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time FROM %s WHERE f_id = ?", DISCOVER_SCHEDULE_TABLE_NAME)

		Convey("GetByID Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs("schedule-1").WillReturnRows(mockDiscoverScheduleRows())

			schedule, err := dsa.GetByID(context.Background(), "schedule-1")
			So(err, ShouldBeNil)
			So(schedule.ID, ShouldEqual, "schedule-1")
			So(schedule.Creator.ID, ShouldEqual, "user-1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetByID Success no row\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs("missing").WillReturnError(sql.ErrNoRows)

			schedule, err := dsa.GetByID(context.Background(), "missing")
			So(err, ShouldBeNil)
			So(schedule, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DiscoverScheduleAccess_Disable(t *testing.T) {
	Convey("test Disable\n", t, func() {
		dsa, smock := MockNewDiscoverScheduleAccess(t)
		sqlStr := fmt.Sprintf("UPDATE %s SET f_enabled = ? WHERE f_id = ?", DISCOVER_SCHEDULE_TABLE_NAME)

		Convey("Disable Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(0, "schedule-1").WillReturnResult(sqlmock.NewResult(0, 1))

			err := dsa.Disable(context.Background(), "schedule-1")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_DiscoverScheduleAccess_Delete(t *testing.T) {
	Convey("test Delete\n", t, func() {
		dsa, smock := MockNewDiscoverScheduleAccess(t)
		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_id = ?", DISCOVER_SCHEDULE_TABLE_NAME)

		Convey("Delete Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs("schedule-1").WillReturnResult(sqlmock.NewResult(0, 1))

			err := dsa.Delete(context.Background(), "schedule-1")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete failed\n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs("schedule-1").WillReturnError(expectedErr)

			err := dsa.Delete(context.Background(), "schedule-1")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
