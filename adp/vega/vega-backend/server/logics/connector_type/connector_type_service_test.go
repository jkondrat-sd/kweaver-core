// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package connector_type

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	vmock "vega-backend/interfaces/mock"
)

func TestConnectorTypeService_CheckExistByType(t *testing.T) {
	Convey("Test connectorTypeService.CheckExistByType", t, func() {
		ctrl := gomock.NewController(t)
		mockCTA := vmock.NewMockConnectorTypeAccess(ctrl)
		svc := &connectorTypeService{cta: mockCTA}

		Convey("returns true when type found", func() {
			mockCTA.EXPECT().GetByType(gomock.Any(), "mysql").Return(&interfaces.ConnectorType{Type: "mysql"}, nil)

			exists, err := svc.CheckExistByType(context.Background(), "mysql")
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
		})

		Convey("returns false when type not found", func() {
			mockCTA.EXPECT().GetByType(gomock.Any(), "missing").Return(nil, nil)

			exists, err := svc.CheckExistByType(context.Background(), "missing")
			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
		})

		Convey("wraps access error as HTTP 500", func() {
			mockCTA.EXPECT().GetByType(gomock.Any(), "mysql").Return(nil, errors.New("db failed"))

			exists, err := svc.CheckExistByType(context.Background(), "mysql")
			So(exists, ShouldBeFalse)
			assertConnectorTypeHTTPError(err, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_GetFailed)
		})
	})
}

func TestConnectorTypeService_CheckExistByName(t *testing.T) {
	Convey("Test connectorTypeService.CheckExistByName", t, func() {
		ctrl := gomock.NewController(t)
		mockCTA := vmock.NewMockConnectorTypeAccess(ctrl)
		svc := &connectorTypeService{cta: mockCTA}

		Convey("returns true when name found", func() {
			mockCTA.EXPECT().GetByName(gomock.Any(), "MySQL").Return(&interfaces.ConnectorType{Name: "MySQL"}, nil)

			exists, err := svc.CheckExistByName(context.Background(), "MySQL")
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
		})
	})
}

func TestConnectorTypeService_SetEnabled(t *testing.T) {
	Convey("Test connectorTypeService.SetEnabled", t, func() {
		ctrl := gomock.NewController(t)
		mockCTA := vmock.NewMockConnectorTypeAccess(ctrl)
		mockPS := vmock.NewMockPermissionService(ctrl)
		svc := &connectorTypeService{cta: mockCTA, ps: mockPS}

		Convey("delegates to access after permission check", func() {
			mockPS.EXPECT().CheckPermission(gomock.Any(), interfaces.PermissionResource{
				Type: interfaces.AUTH_RESOURCE_TYPE_CONNECTOR_TYPE,
				ID:   "mysql",
			}, []string{interfaces.OPERATION_TYPE_MODIFY}).Return(nil)
			mockCTA.EXPECT().SetEnabled(gomock.Any(), "mysql", true).Return(nil)

			err := svc.SetEnabled(context.Background(), "mysql", true)
			So(err, ShouldBeNil)
		})
	})
}

func TestConnectorTypeService_ListAuthResources(t *testing.T) {
	Convey("Test connectorTypeService.ListAuthResources", t, func() {
		ctrl := gomock.NewController(t)
		mockCTA := vmock.NewMockConnectorTypeAccess(ctrl)
		mockPS := vmock.NewMockPermissionService(ctrl)
		svc := &connectorTypeService{cta: mockCTA, ps: mockPS}

		Convey("returns empty result", func() {
			params := interfaces.AuthResourceQueryParams{PaginationQueryParams: interfaces.PaginationQueryParams{Offset: 0, Limit: 20}}
			mockCTA.EXPECT().ListAuthResources(gomock.Any(), params).Return([]*interfaces.AuthResourceEntry{}, nil)

			entries, total, err := svc.ListAuthResources(context.Background(), params)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(entries, ShouldBeEmpty)
		})

		Convey("filters authorized entries and paginates", func() {
			params := interfaces.AuthResourceQueryParams{PaginationQueryParams: interfaces.PaginationQueryParams{Offset: 1, Limit: 1}}
			entries := []*interfaces.AuthResourceEntry{
				{ID: "mysql", Type: interfaces.AUTH_RESOURCE_TYPE_CONNECTOR_TYPE, Name: "MySQL"},
				nil,
				{ID: "postgresql", Type: interfaces.AUTH_RESOURCE_TYPE_CONNECTOR_TYPE, Name: "PostgreSQL"},
				{ID: "oracle", Type: interfaces.AUTH_RESOURCE_TYPE_CONNECTOR_TYPE, Name: "Oracle"},
			}
			mockCTA.EXPECT().ListAuthResources(gomock.Any(), params).Return(entries, nil)
			mockPS.EXPECT().FilterResources(gomock.Any(), interfaces.AUTH_RESOURCE_TYPE_CONNECTOR_TYPE,
				[]string{"mysql", "postgresql", "oracle"},
				[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
				false,
				interfaces.COMMON_OPERATIONS,
			).Return(map[string]interfaces.PermissionResourceOps{
				"mysql":      {ResourceID: "mysql"},
				"postgresql": {ResourceID: "postgresql"},
			}, nil)

			result, total, err := svc.ListAuthResources(context.Background(), params)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
			So(result, ShouldHaveLength, 1)
			So(result[0].ID, ShouldEqual, "postgresql")
		})
	})
}

func TestPaginateConnectorTypeAuthResources(t *testing.T) {
	Convey("Test paginateConnectorTypeAuthResources", t, func() {
		Convey("returns all entries when limit is -1", func() {
			entries := []*interfaces.AuthResourceEntry{{ID: "mysql"}, {ID: "postgresql"}}
			result := paginateConnectorTypeAuthResources(entries, 0, -1)
			So(result, ShouldResemble, entries)
		})

		Convey("returns empty when offset is out of range", func() {
			entries := []*interfaces.AuthResourceEntry{{ID: "mysql"}}
			result := paginateConnectorTypeAuthResources(entries, 2, 1)
			So(result, ShouldBeEmpty)
		})
	})
}

func assertConnectorTypeHTTPError(err error, status int, errorCode string) {
	So(err, ShouldNotBeNil)
	httpErr, ok := err.(*rest.HTTPError)
	So(ok, ShouldBeTrue)
	So(httpErr.HTTPCode, ShouldEqual, status)
	So(httpErr.BaseError.ErrorCode, ShouldEqual, errorCode)
}
