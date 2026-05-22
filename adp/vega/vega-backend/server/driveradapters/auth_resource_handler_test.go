// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"vega-backend/common"
	"vega-backend/interfaces"
	vmock "vega-backend/interfaces/mock"
)

func TestAuthResourceRestHandler_ListAuthResourcesRoute(t *testing.T) {
	Convey("Test authResourceRestHandler.ListAuthResources routes by type", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		as := vmock.NewMockAuthService(mockCtrl)
		handler := MockNewRestHandler(&common.AppSetting{}, as, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().
			Return(hydra.Visitor{ID: "u1", Type: hydra.VisitorType_User}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/vega-backend/v1/auth-resources", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "resource_type is invalid")
	})
}

func TestAuthResourceRestHandler_ListConnectorTypeResources(t *testing.T) {
	Convey("Test authResourceRestHandler.ListAuthResources for connector-type", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		as := vmock.NewMockAuthService(mockCtrl)
		cts := vmock.NewMockConnectorTypeService(mockCtrl)
		handler := MockNewRestHandler(&common.AppSetting{}, as, nil, nil, nil, nil, cts, nil, nil, nil, nil)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().
			Return(hydra.Visitor{ID: "u1", Type: hydra.VisitorType_User}, nil)

		cts.EXPECT().ListAuthResources(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, params interfaces.AuthResourceQueryParams) ([]*interfaces.AuthResourceEntry, int64, error) {
				So(params.Keyword, ShouldEqual, "mysql")
				return []*interfaces.AuthResourceEntry{
					{ID: interfaces.ConnectorTypeMySQL, Type: interfaces.AuthResourceTypeConnectorType, Name: "MySQL"},
				}, int64(1), nil
			})

		req := httptest.NewRequest(http.MethodGet, "/api/vega-backend/v1/auth-resources?resource_type=connector-type&keyword=mysql", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		So(w.Body.String(), ShouldContainSubstring, `"id":"mysql"`)
		So(w.Body.String(), ShouldContainSubstring, `"type":"connector-type"`)
	})
}

func TestAuthResourceRestHandler_RejectUnsupportedSort(t *testing.T) {
	Convey("Test authResourceRestHandler.ListAuthResources rejects unsupported sort", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		as := vmock.NewMockAuthService(mockCtrl)
		handler := MockNewRestHandler(&common.AppSetting{}, as, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().
			Return(hydra.Visitor{ID: "u1", Type: hydra.VisitorType_User}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/vega-backend/v1/auth-resources?resource_type=resource&sort=update_time", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldContainSubstring, "VegaBackend.InvalidParameter.Sort")
	})
}
