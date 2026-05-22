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
	mock_interfaces "vega-backend/interfaces/mock"
)

func TestConnectorTypeService_CheckExistByType_Found(t *testing.T) {
	Convey("Test ConnectorTypeService CheckExistByType found", t, func() {
		ctrl := gomock.NewController(t)
		mockCTA := mock_interfaces.NewMockConnectorTypeAccess(ctrl)
		mockCTA.EXPECT().GetByType(gomock.Any(), "mysql").Return(&interfaces.ConnectorType{Type: "mysql"}, nil)

		svc := &connectorTypeService{cta: mockCTA}
		exists, err := svc.CheckExistByType(context.Background(), "mysql")
		So(err, ShouldBeNil)
		So(exists, ShouldBeTrue)
	})
}

func TestConnectorTypeService_CheckExistByType_NotFound(t *testing.T) {
	Convey("Test ConnectorTypeService CheckExistByType not found", t, func() {
		ctrl := gomock.NewController(t)
		mockCTA := mock_interfaces.NewMockConnectorTypeAccess(ctrl)
		mockCTA.EXPECT().GetByType(gomock.Any(), "missing").Return(nil, nil)

		svc := &connectorTypeService{cta: mockCTA}
		exists, err := svc.CheckExistByType(context.Background(), "missing")
		So(err, ShouldBeNil)
		So(exists, ShouldBeFalse)
	})
}

func TestConnectorTypeService_CheckExistByType_Error(t *testing.T) {
	Convey("Test ConnectorTypeService CheckExistByType returns HTTP error", t, func() {
		ctrl := gomock.NewController(t)
		mockCTA := mock_interfaces.NewMockConnectorTypeAccess(ctrl)
		mockCTA.EXPECT().GetByType(gomock.Any(), "mysql").Return(nil, errors.New("db failed"))

		svc := &connectorTypeService{cta: mockCTA}
		exists, err := svc.CheckExistByType(context.Background(), "mysql")
		So(exists, ShouldBeFalse)
		assertConnectorTypeHTTPError(err, http.StatusInternalServerError, verrors.VegaBackend_ConnectorType_InternalError_GetFailed)
	})
}

func TestConnectorTypeService_CheckExistByName_Found(t *testing.T) {
	Convey("Test ConnectorTypeService CheckExistByName found", t, func() {
		ctrl := gomock.NewController(t)
		mockCTA := mock_interfaces.NewMockConnectorTypeAccess(ctrl)
		mockCTA.EXPECT().GetByName(gomock.Any(), "MySQL").Return(&interfaces.ConnectorType{Name: "MySQL"}, nil)

		svc := &connectorTypeService{cta: mockCTA}
		exists, err := svc.CheckExistByName(context.Background(), "MySQL")
		So(err, ShouldBeNil)
		So(exists, ShouldBeTrue)
	})
}

func TestConnectorTypeService_SetEnabled_Success(t *testing.T) {
	Convey("Test ConnectorTypeService SetEnabled success", t, func() {
		ctrl := gomock.NewController(t)
		mockCTA := mock_interfaces.NewMockConnectorTypeAccess(ctrl)
		mockPS := mock_interfaces.NewMockPermissionService(ctrl)
		mockPS.EXPECT().CheckPermission(gomock.Any(), interfaces.PermissionResource{
			Type: interfaces.AUTH_RESOURCE_TYPE_CONNECTOR_TYPE,
			ID:   "mysql",
		}, []string{interfaces.OPERATION_TYPE_MODIFY}).Return(nil)
		mockCTA.EXPECT().SetEnabled(gomock.Any(), "mysql", true).Return(nil)

		svc := &connectorTypeService{cta: mockCTA, ps: mockPS}
		err := svc.SetEnabled(context.Background(), "mysql", true)
		So(err, ShouldBeNil)
	})
}

func TestConnectorTypeService_ListAuthResources_Empty(t *testing.T) {
	Convey("Test ConnectorTypeService ListAuthResources empty result", t, func() {
		ctrl := gomock.NewController(t)
		mockCTA := mock_interfaces.NewMockConnectorTypeAccess(ctrl)
		params := interfaces.AuthResourceQueryParams{PaginationQueryParams: interfaces.PaginationQueryParams{Offset: 0, Limit: 20}}
		mockCTA.EXPECT().ListAuthResources(gomock.Any(), params).Return([]*interfaces.AuthResourceEntry{}, nil)

		svc := &connectorTypeService{cta: mockCTA}
		entries, total, err := svc.ListAuthResources(context.Background(), params)
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 0)
		So(entries, ShouldBeEmpty)
	})
}

func TestConnectorTypeService_ListAuthResources_FiltersAndPaginates(t *testing.T) {
	Convey("Test ConnectorTypeService ListAuthResources filters authorized entries and paginates", t, func() {
		ctrl := gomock.NewController(t)
		mockCTA := mock_interfaces.NewMockConnectorTypeAccess(ctrl)
		mockPS := mock_interfaces.NewMockPermissionService(ctrl)
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

		svc := &connectorTypeService{cta: mockCTA, ps: mockPS}
		result, total, err := svc.ListAuthResources(context.Background(), params)
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 2)
		So(result, ShouldHaveLength, 1)
		So(result[0].ID, ShouldEqual, "postgresql")
	})
}

func TestPaginateConnectorTypeAuthResources_All(t *testing.T) {
	Convey("Test paginateConnectorTypeAuthResources returns all entries when limit is -1", t, func() {
		entries := []*interfaces.AuthResourceEntry{{ID: "mysql"}, {ID: "postgresql"}}
		result := paginateConnectorTypeAuthResources(entries, 0, -1)
		So(result, ShouldResemble, entries)
	})
}

func TestPaginateConnectorTypeAuthResources_OffsetOutOfRange(t *testing.T) {
	Convey("Test paginateConnectorTypeAuthResources returns empty when offset is out of range", t, func() {
		entries := []*interfaces.AuthResourceEntry{{ID: "mysql"}}
		result := paginateConnectorTypeAuthResources(entries, 2, 1)
		So(result, ShouldBeEmpty)
	})
}

func assertConnectorTypeHTTPError(err error, status int, errorCode string) {
	So(err, ShouldNotBeNil)
	httpErr, ok := err.(*rest.HTTPError)
	So(ok, ShouldBeTrue)
	So(httpErr.HTTPCode, ShouldEqual, status)
	So(httpErr.BaseError.ErrorCode, ShouldEqual, errorCode)
}
