// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package build_task

import (
	"context"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.uber.org/mock/gomock"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	vmock "vega-backend/interfaces/mock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildTaskService_CreateBuildTask(t *testing.T) {
	Convey("Test buildTaskService.CreateBuildTask", t, func() {
		ctrl := gomock.NewController(t)
		mockCS := vmock.NewMockCatalogService(ctrl)
		mockRA := vmock.NewMockResourceAccess(ctrl)
		service := &buildTaskService{cs: mockCS, ra: mockRA}

		Convey("rejects when catalog is disabled", func() {
			mockRA.EXPECT().GetByID(gomock.Any(), "resource-1").
				Return(&interfaces.Resource{
					ID:        "resource-1",
					CatalogID: "catalog-1",
					Category:  interfaces.ResourceCategoryTable,
				}, nil)
			mockCS.EXPECT().GetByID(gomock.Any(), "catalog-1", false).
				Return(&interfaces.Catalog{ID: "catalog-1", Enabled: false}, nil)

			_, err := service.CreateBuildTask(context.Background(), &interfaces.CreateBuildTaskRequest{ResourceID: "resource-1"})
			assertCatalogDisabledError(err)
		})
	})
}

func TestBuildTaskService_StartBuildTask(t *testing.T) {
	Convey("Test buildTaskService.StartBuildTask", t, func() {
		ctrl := gomock.NewController(t)
		mockCS := vmock.NewMockCatalogService(ctrl)
		mockBTA := vmock.NewMockBuildTaskAccess(ctrl)
		service := &buildTaskService{cs: mockCS, bta: mockBTA}

		Convey("rejects when catalog is disabled", func() {
			mockBTA.EXPECT().GetByID(gomock.Any(), "task-1").
				Return(&interfaces.BuildTask{
					ID:        "task-1",
					CatalogID: "catalog-1",
					Status:    interfaces.BuildTaskStatusInit,
				}, nil)
			mockCS.EXPECT().GetByID(gomock.Any(), "catalog-1", false).
				Return(&interfaces.Catalog{ID: "catalog-1", Enabled: false}, nil)

			err := service.StartBuildTask(context.Background(), "task-1", interfaces.BuildTaskExecuteTypeIncremental)
			assertCatalogDisabledError(err)
		})
	})
}

func assertCatalogDisabledError(err error) {
	So(err, ShouldNotBeNil)
	httpErr, ok := err.(*rest.HTTPError)
	So(ok, ShouldBeTrue)
	So(httpErr.HTTPCode, ShouldEqual, http.StatusConflict)
	So(httpErr.BaseError.ErrorCode, ShouldEqual, verrors.VegaBackend_Catalog_IsDisabled)
}
