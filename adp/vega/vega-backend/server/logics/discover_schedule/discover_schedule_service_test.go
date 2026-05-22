// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package discover_schedule

import (
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"vega-backend/interfaces"
	mock_interfaces "vega-backend/interfaces/mock"
)

func TestDiscoverScheduleService_Create_RejectsEmptyCron(t *testing.T) {
	Convey("Test DiscoverScheduleService Create rejects empty cron", t, func() {
		svc := &discoverScheduleService{}
		id, err := svc.Create(context.Background(), &interfaces.DiscoverScheduleRequest{CatalogID: "catalog-1"})
		So(id, ShouldBeBlank)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "cron_expr is required")
	})
}

func TestDiscoverScheduleService_Create_Success(t *testing.T) {
	Convey("Test DiscoverScheduleService Create success", t, func() {
		ctrl := gomock.NewController(t)
		mockDSA := mock_interfaces.NewMockDiscoverScheduleAccess(ctrl)
		mockDSA.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, schedule *interfaces.DiscoverSchedule) error {
				So(schedule.ID, ShouldNotBeBlank)
				So(schedule.Name, ShouldEqual, "daily")
				So(schedule.CatalogID, ShouldEqual, "catalog-1")
				So(schedule.CronExpr, ShouldEqual, "0 0 * * *")
				So(schedule.Creator.ID, ShouldEqual, "account-1")
				So(schedule.Updater.ID, ShouldEqual, "account-1")
				return nil
			},
		)

		svc := &discoverScheduleService{dsa: mockDSA}
		ctx := context.WithValue(context.Background(), interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{ID: "account-1"})
		id, err := svc.Create(ctx, &interfaces.DiscoverScheduleRequest{
			Name:      "daily",
			CatalogID: "catalog-1",
			CronExpr:  "0 0 * * *",
			Strategy:  "full_sync",
			Enabled:   true,
		})
		So(err, ShouldBeNil)
		So(id, ShouldNotBeBlank)
	})
}

func TestDiscoverScheduleService_GetByID_EnrichesAccountNames(t *testing.T) {
	Convey("Test DiscoverScheduleService GetByID enriches account names", t, func() {
		ctrl := gomock.NewController(t)
		mockDSA := mock_interfaces.NewMockDiscoverScheduleAccess(ctrl)
		mockUMS := mock_interfaces.NewMockUserMgmtService(ctrl)
		schedule := &interfaces.DiscoverSchedule{
			ID:      "schedule-1",
			Creator: interfaces.AccountInfo{ID: "creator"},
			Updater: interfaces.AccountInfo{ID: "updater"},
		}
		mockDSA.EXPECT().GetByID(gomock.Any(), "schedule-1").Return(schedule, nil)
		mockUMS.EXPECT().GetAccountNames(gomock.Any(), []*interfaces.AccountInfo{&schedule.Creator, &schedule.Updater}).Return(nil)

		svc := &discoverScheduleService{dsa: mockDSA, ums: mockUMS}
		result, err := svc.GetByID(context.Background(), "schedule-1")
		So(err, ShouldBeNil)
		So(result.ID, ShouldEqual, "schedule-1")
	})
}

func TestDiscoverScheduleService_Update_RejectsNilSchedule(t *testing.T) {
	Convey("Test DiscoverScheduleService Update rejects nil schedule", t, func() {
		svc := &discoverScheduleService{}
		err := svc.Update(context.Background(), nil, &interfaces.DiscoverScheduleRequest{})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "discover schedule not found")
	})
}

func TestDiscoverScheduleService_Update_Success(t *testing.T) {
	Convey("Test DiscoverScheduleService Update success", t, func() {
		ctrl := gomock.NewController(t)
		mockDSA := mock_interfaces.NewMockDiscoverScheduleAccess(ctrl)
		schedule := &interfaces.DiscoverSchedule{ID: "schedule-1"}
		mockDSA.EXPECT().Update(gomock.Any(), schedule).DoAndReturn(
			func(ctx context.Context, updated *interfaces.DiscoverSchedule) error {
				So(updated.Name, ShouldEqual, "weekly")
				So(updated.CronExpr, ShouldEqual, "0 0 * * 1")
				So(updated.Updater.ID, ShouldEqual, "account-1")
				return nil
			},
		)

		svc := &discoverScheduleService{dsa: mockDSA}
		ctx := context.WithValue(context.Background(), interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{ID: "account-1"})
		err := svc.Update(ctx, schedule, &interfaces.DiscoverScheduleRequest{
			Name:     "weekly",
			CronExpr: "0 0 * * 1",
			Strategy: "full_sync",
		})
		So(err, ShouldBeNil)
	})
}

func TestDiscoverScheduleService_Enable_DelegatesToAccess(t *testing.T) {
	Convey("Test DiscoverScheduleService Enable delegates to access", t, func() {
		ctrl := gomock.NewController(t)
		mockDSA := mock_interfaces.NewMockDiscoverScheduleAccess(ctrl)
		mockDSA.EXPECT().Enable(gomock.Any(), "schedule-1").Return(nil)

		svc := &discoverScheduleService{dsa: mockDSA}
		err := svc.Enable(context.Background(), "schedule-1")
		So(err, ShouldBeNil)
	})
}

func TestDiscoverScheduleService_Disable_DelegatesToAccess(t *testing.T) {
	Convey("Test DiscoverScheduleService Disable delegates to access", t, func() {
		ctrl := gomock.NewController(t)
		mockDSA := mock_interfaces.NewMockDiscoverScheduleAccess(ctrl)
		mockDSA.EXPECT().Disable(gomock.Any(), "schedule-1").Return(nil)

		svc := &discoverScheduleService{dsa: mockDSA}
		err := svc.Disable(context.Background(), "schedule-1")
		So(err, ShouldBeNil)
	})
}

func TestDiscoverScheduleService_ExecuteSchedule_RejectsMissingTaskService(t *testing.T) {
	Convey("Test DiscoverScheduleService ExecuteSchedule rejects missing task service", t, func() {
		svc := &discoverScheduleService{}
		err := svc.ExecuteSchedule(context.Background(), &interfaces.DiscoverSchedule{ID: "schedule-1"})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "DiscoverTaskService not set")
	})
}

func TestDiscoverScheduleService_ExecuteSchedule_SkipsWhenRunningTaskExists(t *testing.T) {
	Convey("Test DiscoverScheduleService ExecuteSchedule skips when running task exists", t, func() {
		ctrl := gomock.NewController(t)
		mockDTS := mock_interfaces.NewMockDiscoverTaskService(ctrl)
		mockDTS.EXPECT().List(gomock.Any(), interfaces.DiscoverTaskQueryParams{
			CatalogID:   "catalog-1",
			Status:      interfaces.DiscoverTaskStatusRunning,
			TriggerType: interfaces.DiscoverTaskTriggerScheduled,
		}).Return(nil, int64(1), nil)

		svc := &discoverScheduleService{dts: mockDTS}
		err := svc.ExecuteSchedule(context.Background(), &interfaces.DiscoverSchedule{ID: "schedule-1", CatalogID: "catalog-1"})
		So(err, ShouldBeNil)
	})
}

func TestDiscoverScheduleService_ExecuteSchedule_Success(t *testing.T) {
	Convey("Test DiscoverScheduleService ExecuteSchedule creates task and updates last run", t, func() {
		ctrl := gomock.NewController(t)
		mockDTS := mock_interfaces.NewMockDiscoverTaskService(ctrl)
		mockDSA := mock_interfaces.NewMockDiscoverScheduleAccess(ctrl)
		schedule := &interfaces.DiscoverSchedule{
			ID:        "schedule-1",
			CatalogID: "catalog-1",
			Strategy:  "full_sync",
			Creator:   interfaces.AccountInfo{ID: "account-1"},
		}
		mockDTS.EXPECT().List(gomock.Any(), interfaces.DiscoverTaskQueryParams{
			CatalogID:   "catalog-1",
			Status:      interfaces.DiscoverTaskStatusRunning,
			TriggerType: interfaces.DiscoverTaskTriggerScheduled,
		}).Return(nil, int64(0), nil)
		mockDTS.EXPECT().Create(gomock.Any(), &interfaces.CreateDiscoverTaskRequest{
			CatalogID:   "catalog-1",
			TriggerType: interfaces.DiscoverTaskTriggerScheduled,
			ScheduleID:  "schedule-1",
			Strategy:    "full_sync",
		}).Return("task-1", nil)
		mockDSA.EXPECT().UpdateLastRun(gomock.Any(), "schedule-1", gomock.Any()).Return(nil)

		svc := &discoverScheduleService{dsa: mockDSA, dts: mockDTS}
		err := svc.ExecuteSchedule(context.Background(), schedule)
		So(err, ShouldBeNil)
	})
}

func TestDiscoverScheduleService_ExecuteSchedule_ReturnsCreateError(t *testing.T) {
	Convey("Test DiscoverScheduleService ExecuteSchedule returns create error", t, func() {
		ctrl := gomock.NewController(t)
		mockDTS := mock_interfaces.NewMockDiscoverTaskService(ctrl)
		expectedErr := errors.New("create failed")
		mockDTS.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, int64(0), nil)
		mockDTS.EXPECT().Create(gomock.Any(), gomock.Any()).Return("", expectedErr)

		svc := &discoverScheduleService{dts: mockDTS}
		err := svc.ExecuteSchedule(context.Background(), &interfaces.DiscoverSchedule{ID: "schedule-1", CatalogID: "catalog-1"})
		So(err, ShouldEqual, expectedErr)
	})
}
