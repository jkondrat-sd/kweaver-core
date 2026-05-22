package catalog

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

func MockNewCatalogAccess(t *testing.T) (*catalogAccess, sqlmock.Sqlmock) {
	db, smock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return &catalogAccess{appSetting: &common.AppSetting{}, db: db}, smock
}

func Test_CatalogAccess_UpdateHealthCheckStatus(t *testing.T) {
	Convey("test UpdateHealthCheckStatus\n", t, func() {
		ca, smock := MockNewCatalogAccess(t)
		sqlStr := fmt.Sprintf("UPDATE %s SET f_health_check_status = ?, f_last_check_time = ?, f_health_check_result = ? WHERE f_id = ?", CATALOG_TABLE_NAME)
		status := interfaces.CatalogHealthCheckStatus{
			HealthCheckStatus: interfaces.CatalogHealthStatusHealthy,
			LastCheckTime:     1000,
			HealthCheckResult: "ok",
		}

		Convey("UpdateHealthCheckStatus Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(status.HealthCheckStatus, status.LastCheckTime, status.HealthCheckResult, "catalog-1").
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := ca.UpdateHealthCheckStatus(context.Background(), "catalog-1", status)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateHealthCheckStatus Failed\n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs(status.HealthCheckStatus, status.LastCheckTime, status.HealthCheckResult, "catalog-1").
				WillReturnError(expectedErr)

			err := ca.UpdateHealthCheckStatus(context.Background(), "catalog-1", status)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_CatalogAccess_UpdateEnabled(t *testing.T) {
	Convey("test UpdateEnabled\n", t, func() {
		ca, smock := MockNewCatalogAccess(t)
		sqlStr := fmt.Sprintf("UPDATE %s SET f_enabled = ?, f_health_check_status = ?, f_last_check_time = ?, f_health_check_result = ?, f_updater = ?, f_updater_type = ?, f_update_time = ? WHERE f_id = ?", CATALOG_TABLE_NAME)
		status := interfaces.CatalogHealthCheckStatus{
			HealthCheckStatus: interfaces.CatalogHealthStatusHealthy,
			LastCheckTime:     1000,
			HealthCheckResult: "ok",
		}
		updater := interfaces.AccountInfo{ID: "user-1", Type: interfaces.ACCESSOR_TYPE_USER}

		Convey("UpdateEnabled Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(true, status.HealthCheckStatus, status.LastCheckTime, status.HealthCheckResult, updater.ID, updater.Type, int64(2000), "catalog-1").
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := ca.UpdateEnabled(context.Background(), "catalog-1", true, status, 2000, updater)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_CatalogAccess_UpdateMetadata(t *testing.T) {
	Convey("test UpdateMetadata\n", t, func() {
		ca, smock := MockNewCatalogAccess(t)
		sqlStr := fmt.Sprintf("UPDATE %s SET f_metadata = ? WHERE f_id = ?", CATALOG_TABLE_NAME)

		Convey("UpdateMetadata Success\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(sqlmock.AnyArg(), "catalog-1").
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := ca.UpdateMetadata(context.Background(), "catalog-1", map[string]any{"region": "cn"})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
