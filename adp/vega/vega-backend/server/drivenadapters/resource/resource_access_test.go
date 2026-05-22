package resource

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/common"
	"vega-backend/interfaces"
)

func MockNewResourceAccess(t *testing.T) (*resourceAccess, sqlmock.Sqlmock) {
	db, smock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return &resourceAccess{appSetting: &common.AppSetting{}, db: db}, smock
}

func TestResourceAccess_UpdateStatus(t *testing.T) {
	Convey("Test resourceAccess.UpdateStatus", t, func() {
		ra, smock := MockNewResourceAccess(t)
		sqlStr := fmt.Sprintf("UPDATE %s SET f_status = ?, f_status_message = ? WHERE f_id = ?", RESOURCE_TABLE_NAME)

		Convey("updates status successfully", func() {
			smock.ExpectExec(sqlStr).WithArgs(interfaces.ResourceStatusActive, "ok", "resource-1").
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := ra.UpdateStatus(context.Background(), "resource-1", interfaces.ResourceStatusActive, "ok")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("propagates db error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs(interfaces.ResourceStatusActive, "ok", "resource-1").WillReturnError(expectedErr)

			err := ra.UpdateStatus(context.Background(), "resource-1", interfaces.ResourceStatusActive, "ok")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func TestResourceAccess_UpdateDiscoverStatus(t *testing.T) {
	Convey("Test resourceAccess.UpdateDiscoverStatus", t, func() {
		ra, smock := MockNewResourceAccess(t)
		sqlStr := fmt.Sprintf("UPDATE %s SET f_last_discover_status = ? WHERE f_id = ?", RESOURCE_TABLE_NAME)

		Convey("updates discover status successfully", func() {
			smock.ExpectExec(sqlStr).WithArgs(interfaces.DiscoverStatusUpdated, "resource-1").WillReturnResult(sqlmock.NewResult(0, 1))

			err := ra.UpdateDiscoverStatus(context.Background(), "resource-1", interfaces.DiscoverStatusUpdated)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func TestResourceAccess_CheckExistByCategories(t *testing.T) {
	Convey("Test resourceAccess.CheckExistByCategories", t, func() {
		ra, smock := MockNewResourceAccess(t)
		sqlStr := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE f_catalog_id = ? AND f_category IN (?,?)", RESOURCE_TABLE_NAME)

		Convey("returns true when matching rows exist", func() {
			rows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
			smock.ExpectQuery(sqlStr).WithArgs("catalog-1", interfaces.ResourceCategoryTable, interfaces.ResourceCategoryFile).WillReturnRows(rows)

			exists, err := ra.CheckExistByCategories(context.Background(), "catalog-1", []string{interfaces.ResourceCategoryTable, interfaces.ResourceCategoryFile})
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("propagates db error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs("catalog-1", interfaces.ResourceCategoryTable, interfaces.ResourceCategoryFile).WillReturnError(expectedErr)

			exists, err := ra.CheckExistByCategories(context.Background(), "catalog-1", []string{interfaces.ResourceCategoryTable, interfaces.ResourceCategoryFile})
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
