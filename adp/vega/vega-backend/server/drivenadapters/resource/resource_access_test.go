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

func Test_ResourceAccess_UpdateStatus(t *testing.T) {
	Convey("test UpdateStatus\n", t, func() {
		ra, smock := MockNewResourceAccess(t)
		sqlStr := fmt.Sprintf("UPDATE %s SET f_status = ?, f_status_message = ? WHERE f_id = ?", RESOURCE_TABLE_NAME)

		Convey("UpdateStatus Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(interfaces.ResourceStatusActive, "ok", "resource-1").
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := ra.UpdateStatus(context.Background(), "resource-1", interfaces.ResourceStatusActive, "ok")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateStatus Failed\n", func() {
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

func Test_ResourceAccess_UpdateDiscoverStatus(t *testing.T) {
	Convey("test UpdateDiscoverStatus\n", t, func() {
		ra, smock := MockNewResourceAccess(t)
		sqlStr := fmt.Sprintf("UPDATE %s SET f_last_discover_status = ? WHERE f_id = ?", RESOURCE_TABLE_NAME)

		Convey("UpdateDiscoverStatus Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(interfaces.DiscoverStatusUpdated, "resource-1").WillReturnResult(sqlmock.NewResult(0, 1))

			err := ra.UpdateDiscoverStatus(context.Background(), "resource-1", interfaces.DiscoverStatusUpdated)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ResourceAccess_CheckExistByCategories(t *testing.T) {
	Convey("test CheckExistByCategories\n", t, func() {
		ra, smock := MockNewResourceAccess(t)
		sqlStr := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE f_catalog_id = ? AND f_category IN (?,?)", RESOURCE_TABLE_NAME)

		Convey("CheckExistByCategories Success true\n", func() {
			rows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
			smock.ExpectQuery(sqlStr).WithArgs("catalog-1", interfaces.ResourceCategoryTable, interfaces.ResourceCategoryFile).WillReturnRows(rows)

			exists, err := ra.CheckExistByCategories(context.Background(), "catalog-1", []string{interfaces.ResourceCategoryTable, interfaces.ResourceCategoryFile})
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckExistByCategories Failed\n", func() {
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
