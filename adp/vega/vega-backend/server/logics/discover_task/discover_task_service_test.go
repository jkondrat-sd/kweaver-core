// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package discover_task

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	mock_interfaces "vega-backend/interfaces/mock"
)

func TestDiscoverTaskService_Create_DebugMode(t *testing.T) {
	Convey("Test DiscoverTaskService Create writes task and enqueues debug task", t, func() {
		t.Setenv("DEBUG_MODE", "true")
		ctrl := gomock.NewController(t)
		mockDTA := mock_interfaces.NewMockDiscoverTaskAccess(ctrl)
		mockDTA.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, task *interfaces.DiscoverTask) error {
				So(task.CatalogID, ShouldEqual, "catalog-1")
				So(task.ScheduleID, ShouldEqual, "schedule-1")
				So(task.Strategy, ShouldEqual, "full_sync")
				So(task.TriggerType, ShouldEqual, interfaces.DiscoverTaskTriggerManual)
				So(task.Status, ShouldEqual, interfaces.DiscoverTaskStatusPending)
				So(task.Creator.ID, ShouldEqual, "account-1")
				return nil
			},
		)

		svc := &discoverTaskService{
			dta:            mockDTA,
			debugTaskQueue: make(chan *asynq.Task, debugDiscoverTaskQueueSize),
		}

		ctx := context.WithValue(context.Background(), interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{ID: "account-1", Type: "user"})
		taskID, err := svc.Create(ctx, &interfaces.CreateDiscoverTaskRequest{
			CatalogID:   "catalog-1",
			ScheduleID:  "schedule-1",
			Strategy:    "full_sync",
			TriggerType: interfaces.DiscoverTaskTriggerManual,
		})
		So(err, ShouldBeNil)
		So(taskID, ShouldNotBeBlank)

		queuedTask := <-svc.DebugTaskQueue()
		So(queuedTask.Type(), ShouldEqual, interfaces.DiscoverTaskType)
		var message interfaces.DiscoverTaskMessage
		err = sonic.Unmarshal(queuedTask.Payload(), &message)
		So(err, ShouldBeNil)
		So(message.TaskID, ShouldEqual, taskID)
	})
}

func TestDiscoverTaskService_GetByID_EnrichesCreatorName(t *testing.T) {
	Convey("Test DiscoverTaskService GetByID enriches creator name", t, func() {
		ctrl := gomock.NewController(t)
		mockDTA := mock_interfaces.NewMockDiscoverTaskAccess(ctrl)
		mockUMS := mock_interfaces.NewMockUserMgmtService(ctrl)
		task := &interfaces.DiscoverTask{ID: "task-1", Creator: interfaces.AccountInfo{ID: "account-1"}}
		mockDTA.EXPECT().GetByID(gomock.Any(), "task-1").Return(task, nil)
		mockUMS.EXPECT().GetAccountNames(gomock.Any(), []*interfaces.AccountInfo{&task.Creator}).DoAndReturn(
			func(ctx context.Context, accountInfos []*interfaces.AccountInfo) error {
				accountInfos[0].Name = "Alice"
				return nil
			},
		)

		svc := &discoverTaskService{dta: mockDTA, ums: mockUMS}
		result, err := svc.GetByID(context.Background(), "task-1")
		So(err, ShouldBeNil)
		So(result.Creator.Name, ShouldEqual, "Alice")
	})
}

func TestDiscoverTaskService_GetByID_GetAccountNamesError(t *testing.T) {
	Convey("Test DiscoverTaskService GetByID wraps GetAccountNames error", t, func() {
		ctrl := gomock.NewController(t)
		mockDTA := mock_interfaces.NewMockDiscoverTaskAccess(ctrl)
		mockUMS := mock_interfaces.NewMockUserMgmtService(ctrl)
		task := &interfaces.DiscoverTask{ID: "task-1", Creator: interfaces.AccountInfo{ID: "account-1"}}
		mockDTA.EXPECT().GetByID(gomock.Any(), "task-1").Return(task, nil)
		mockUMS.EXPECT().GetAccountNames(gomock.Any(), []*interfaces.AccountInfo{&task.Creator}).Return(errors.New("iam failed"))

		svc := &discoverTaskService{dta: mockDTA, ums: mockUMS}
		result, err := svc.GetByID(context.Background(), "task-1")
		So(result, ShouldBeNil)
		assertDiscoverTaskHTTPError(err, http.StatusInternalServerError, verrors.VegaBackend_DiscoverTask_InternalError_GetAccountNamesFailed)
	})
}

func TestDiscoverTaskService_List_EnrichesCreatorNames(t *testing.T) {
	Convey("Test DiscoverTaskService List enriches creator names", t, func() {
		ctrl := gomock.NewController(t)
		mockDTA := mock_interfaces.NewMockDiscoverTaskAccess(ctrl)
		mockUMS := mock_interfaces.NewMockUserMgmtService(ctrl)
		params := interfaces.DiscoverTaskQueryParams{CatalogID: "catalog-1"}
		task1 := &interfaces.DiscoverTask{ID: "task-1", Creator: interfaces.AccountInfo{ID: "account-1"}}
		task2 := &interfaces.DiscoverTask{ID: "task-2", Creator: interfaces.AccountInfo{ID: "account-2"}}
		mockDTA.EXPECT().List(gomock.Any(), params).Return([]*interfaces.DiscoverTask{task1, task2}, int64(2), nil)
		mockUMS.EXPECT().GetAccountNames(gomock.Any(), []*interfaces.AccountInfo{&task1.Creator, &task2.Creator}).Return(nil)

		svc := &discoverTaskService{dta: mockDTA, ums: mockUMS}
		tasks, total, err := svc.List(context.Background(), params)
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 2)
		So(tasks, ShouldHaveLength, 2)
	})
}

func TestDiscoverTaskService_UpdateStatus_DelegatesToAccess(t *testing.T) {
	Convey("Test DiscoverTaskService UpdateStatus delegates to access", t, func() {
		ctrl := gomock.NewController(t)
		mockDTA := mock_interfaces.NewMockDiscoverTaskAccess(ctrl)
		mockDTA.EXPECT().UpdateStatus(gomock.Any(), "task-1", interfaces.DiscoverTaskStatusRunning, "running", int64(100)).Return(nil)

		svc := &discoverTaskService{dta: mockDTA}
		err := svc.UpdateStatus(context.Background(), "task-1", interfaces.DiscoverTaskStatusRunning, "running", 100)
		So(err, ShouldBeNil)
	})
}

func TestDiscoverTaskService_Delete_RejectsRunningTask(t *testing.T) {
	Convey("Test DiscoverTaskService Delete rejects running task", t, func() {
		ctrl := gomock.NewController(t)
		mockDTA := mock_interfaces.NewMockDiscoverTaskAccess(ctrl)
		mockDTA.EXPECT().GetByID(gomock.Any(), "task-1").Return(&interfaces.DiscoverTask{
			ID:     "task-1",
			Status: interfaces.DiscoverTaskStatusRunning,
		}, nil)

		svc := &discoverTaskService{dta: mockDTA}
		err := svc.Delete(context.Background(), []string{"task-1"}, false)
		assertDiscoverTaskHTTPError(err, http.StatusConflict, verrors.VegaBackend_DiscoverTask_HasRunningExecution)
	})
}

func TestDiscoverTaskService_Delete_RejectsMissingTask(t *testing.T) {
	Convey("Test DiscoverTaskService Delete rejects missing task", t, func() {
		ctrl := gomock.NewController(t)
		mockDTA := mock_interfaces.NewMockDiscoverTaskAccess(ctrl)
		mockDTA.EXPECT().GetByID(gomock.Any(), "missing").Return(nil, nil)

		svc := &discoverTaskService{dta: mockDTA}
		err := svc.Delete(context.Background(), []string{"missing"}, false)
		assertDiscoverTaskHTTPError(err, http.StatusNotFound, verrors.VegaBackend_DiscoverTask_NotFound)
	})
}

func TestDiscoverTaskService_Delete_SuccessDeduplicatesIDs(t *testing.T) {
	Convey("Test DiscoverTaskService Delete succeeds and deduplicates IDs", t, func() {
		ctrl := gomock.NewController(t)
		mockDTA := mock_interfaces.NewMockDiscoverTaskAccess(ctrl)
		mockDTA.EXPECT().GetByID(gomock.Any(), "task-1").Return(&interfaces.DiscoverTask{
			ID:     "task-1",
			Status: interfaces.DiscoverTaskStatusCompleted,
		}, nil)
		mockDTA.EXPECT().Delete(gomock.Any(), "task-1").Return(nil)

		svc := &discoverTaskService{dta: mockDTA}
		err := svc.Delete(context.Background(), []string{"task-1", "task-1"}, false)
		So(err, ShouldBeNil)
	})
}

func assertDiscoverTaskHTTPError(err error, status int, errorCode string) {
	So(err, ShouldNotBeNil)
	httpErr, ok := err.(*rest.HTTPError)
	So(ok, ShouldBeTrue)
	So(httpErr.HTTPCode, ShouldEqual, status)
	So(httpErr.BaseError.ErrorCode, ShouldEqual, errorCode)
}
