// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package user_mgmt

import (
	"context"
	"errors"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"vega-backend/common"
	"vega-backend/interfaces"
	mock_interfaces "vega-backend/interfaces/mock"
	"vega-backend/logics"
)

func TestUserMgmtServiceImpl_GetAccountNames_Success(t *testing.T) {
	Convey("Test UserMgmtServiceImpl GetAccountNames delegates to access", t, func() {
		ctrl := gomock.NewController(t)
		mockUMA := mock_interfaces.NewMockUserMgmtAccess(ctrl)
		accounts := []*interfaces.AccountInfo{{ID: "account-1"}}
		mockUMA.EXPECT().GetAccountNames(gomock.Any(), accounts).Return(nil)

		svc := &UserMgmtServiceImpl{uma: mockUMA}
		err := svc.GetAccountNames(context.Background(), accounts)
		So(err, ShouldBeNil)
	})
}

func TestUserMgmtServiceImpl_GetAccountNames_Error(t *testing.T) {
	Convey("Test UserMgmtServiceImpl GetAccountNames returns access error", t, func() {
		ctrl := gomock.NewController(t)
		mockUMA := mock_interfaces.NewMockUserMgmtAccess(ctrl)
		expectedErr := errors.New("access failed")
		accounts := []*interfaces.AccountInfo{{ID: "account-1"}}
		mockUMA.EXPECT().GetAccountNames(gomock.Any(), accounts).Return(expectedErr)

		svc := &UserMgmtServiceImpl{uma: mockUMA}
		err := svc.GetAccountNames(context.Background(), accounts)
		So(err, ShouldEqual, expectedErr)
	})
}

func TestNewUserMgmtService_DisabledAuthUsesNoop(t *testing.T) {
	Convey("Test NewUserMgmtService returns noop service when auth is disabled", t, func() {
		t.Setenv("AUTH_ENABLED", "false")
		resetUserMgmtServiceSingleton()

		svc := NewUserMgmtService(&common.AppSetting{})
		_, ok := svc.(*NoopUserMgmtService)
		So(ok, ShouldBeTrue)
	})
}

func TestNewUserMgmtService_EnabledAuthUsesImpl(t *testing.T) {
	Convey("Test NewUserMgmtService returns impl service when auth is enabled", t, func() {
		t.Setenv("AUTH_ENABLED", "true")
		resetUserMgmtServiceSingleton()
		oldUMA := logics.UMA
		logics.SetUserMgmtAccess(nil)
		t.Cleanup(func() {
			logics.SetUserMgmtAccess(oldUMA)
			resetUserMgmtServiceSingleton()
		})

		svc := NewUserMgmtService(&common.AppSetting{})
		_, ok := svc.(*UserMgmtServiceImpl)
		So(ok, ShouldBeTrue)
	})
}

func resetUserMgmtServiceSingleton() {
	umServiceOnce = sync.Once{}
	umService = nil
}
