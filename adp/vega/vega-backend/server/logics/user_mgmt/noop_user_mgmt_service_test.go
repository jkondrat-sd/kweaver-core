// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package user_mgmt

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestNoopUserMgmtService_GetAccountNames_FillsMissingNames(t *testing.T) {
	Convey("Test NoopUserMgmtService GetAccountNames fills missing names", t, func() {
		svc := NewNoopUserMgmtService(nil)
		accounts := []*interfaces.AccountInfo{
			{ID: "account-1", Type: "user"},
			{ID: "account-2", Type: "user", Name: "Alice"},
		}

		err := svc.GetAccountNames(context.Background(), accounts)
		So(err, ShouldBeNil)
		So(accounts[0].Name, ShouldEqual, "account-1")
		So(accounts[1].Name, ShouldEqual, "Alice")
	})
}

func TestNoopUserMgmtService_GetAccountNames_EmptyInput(t *testing.T) {
	Convey("Test NoopUserMgmtService GetAccountNames accepts empty input", t, func() {
		svc := NewNoopUserMgmtService(nil)
		err := svc.GetAccountNames(context.Background(), nil)
		So(err, ShouldBeNil)
	})
}
