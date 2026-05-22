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
	vmock "vega-backend/interfaces/mock"
	"vega-backend/logics"
)

func TestUserMgmtServiceImpl_GetAccountNames(t *testing.T) {
	Convey("Test UserMgmtServiceImpl.GetAccountNames", t, func() {
		ctrl := gomock.NewController(t)
		mockUMA := vmock.NewMockUserMgmtAccess(ctrl)
		svc := &UserMgmtServiceImpl{uma: mockUMA}
		accounts := []*interfaces.AccountInfo{{ID: "account-1"}}

		Convey("delegates to access layer", func() {
			mockUMA.EXPECT().GetAccountNames(gomock.Any(), accounts).Return(nil)

			err := svc.GetAccountNames(context.Background(), accounts)
			So(err, ShouldBeNil)
		})

		Convey("returns access error as-is", func() {
			expectedErr := errors.New("access failed")
			mockUMA.EXPECT().GetAccountNames(gomock.Any(), accounts).Return(expectedErr)

			err := svc.GetAccountNames(context.Background(), accounts)
			So(err, ShouldEqual, expectedErr)
		})
	})
}

func TestNewUserMgmtService(t *testing.T) {
	Convey("Test NewUserMgmtService", t, func() {
		Convey("returns noop service when auth is disabled", func() {
			t.Setenv("AUTH_ENABLED", "false")
			resetUserMgmtServiceSingleton()

			svc := NewUserMgmtService(&common.AppSetting{})
			_, ok := svc.(*NoopUserMgmtService)
			So(ok, ShouldBeTrue)
		})

		Convey("returns impl service when auth is enabled", func() {
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
	})
}

func resetUserMgmtServiceSingleton() {
	umServiceOnce = sync.Once{}
	umService = nil
}
