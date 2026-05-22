// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"

	"vega-backend/interfaces"
	vmock "vega-backend/interfaces/mock"

	. "github.com/smartystreets/goconvey/convey"
)

// newTestService 使用 mockgen 生成的 mock 构建 resourceService
func newTestService(t *testing.T) (*resourceService,
	*vmock.MockResourceAccess,
	*vmock.MockPermissionService,
	*vmock.MockDatasetService,
	*vmock.MockUserMgmtService,
	*vmock.MockCatalogService,
	*vmock.MockBuildTaskAccess) {

	ctrl := gomock.NewController(t)
	mockRA := vmock.NewMockResourceAccess(ctrl)
	mockPS := vmock.NewMockPermissionService(ctrl)
	mockDS := vmock.NewMockDatasetService(ctrl)
	mockUMS := vmock.NewMockUserMgmtService(ctrl)
	mockCS := vmock.NewMockCatalogService(ctrl)
	mockBTA := vmock.NewMockBuildTaskAccess(ctrl)

	rs := &resourceService{
		ra:  mockRA,
		ps:  mockPS,
		ds:  mockDS,
		ums: mockUMS,
		cs:  mockCS,
		bta: mockBTA,
	}
	return rs, mockRA, mockPS, mockDS, mockUMS, mockCS, mockBTA
}

// ===== CheckExistByID =====

func TestCheckExistByID_Found(t *testing.T) {
	Convey("Test CheckExistByID found", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().GetByID(gomock.Any(), "r1").
			Return(&interfaces.Resource{ID: "r1"}, nil)

		exists, err := rs.CheckExistByID(context.Background(), "r1")
		So(err, ShouldBeNil)
		So(exists, ShouldBeTrue)
	})
}

func TestCheckExistByID_NotFound(t *testing.T) {
	Convey("Test CheckExistByID not found", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().GetByID(gomock.Any(), "missing").
			Return(nil, nil)

		exists, err := rs.CheckExistByID(context.Background(), "missing")
		So(err, ShouldBeNil)
		So(exists, ShouldBeFalse)
	})
}

func TestCheckExistByID_Error(t *testing.T) {
	Convey("Test CheckExistByID error", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().GetByID(gomock.Any(), "r1").
			Return(nil, fmt.Errorf("db error"))

		_, err := rs.CheckExistByID(context.Background(), "r1")
		So(err, ShouldNotBeNil)
	})
}

// ===== CheckExistByName =====

func TestCheckExistByName_Found(t *testing.T) {
	Convey("Test CheckExistByName found", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().GetByName(gomock.Any(), "cat1", "test").
			Return(&interfaces.Resource{Name: "test"}, nil)

		exists, err := rs.CheckExistByName(context.Background(), "cat1", "test")
		So(err, ShouldBeNil)
		So(exists, ShouldBeTrue)
	})
}

func TestCheckExistByName_NotFound(t *testing.T) {
	Convey("Test CheckExistByName not found", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().GetByName(gomock.Any(), "cat1", "missing").
			Return(nil, nil)

		exists, err := rs.CheckExistByName(context.Background(), "cat1", "missing")
		So(err, ShouldBeNil)
		So(exists, ShouldBeFalse)
	})
}

// ===== GetByID =====

func TestGetByID_Success(t *testing.T) {
	Convey("Test GetByID success", t, func() {
		rs, mockRA, mockPS, _, mockUMS, _, _ := newTestService(t)
		mockRA.EXPECT().GetByID(gomock.Any(), "r1").
			Return(&interfaces.Resource{ID: "r1", Name: "test"}, nil)
		mockPS.EXPECT().FilterResources(gomock.Any(), interfaces.AUTH_RESOURCE_TYPE_RESOURCE,
			[]string{"r1"}, gomock.Any(), true, gomock.Any()).
			Return(map[string]interfaces.PermissionResourceOps{
				"r1": {ResourceID: "r1", Operations: []string{"view_detail"}},
			}, nil)
		mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

		resource, err := rs.GetByID(context.Background(), "r1")
		So(err, ShouldBeNil)
		So(resource.ID, ShouldEqual, "r1")
	})
}

func TestGetByID_NotFound(t *testing.T) {
	Convey("Test GetByID not found", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().GetByID(gomock.Any(), "missing").
			Return(nil, nil)

		_, err := rs.GetByID(context.Background(), "missing")
		So(err, ShouldNotBeNil)
	})
}

func TestGetByID_DBError(t *testing.T) {
	Convey("Test GetByID db error", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().GetByID(gomock.Any(), "r1").
			Return(nil, fmt.Errorf("db error"))

		_, err := rs.GetByID(context.Background(), "r1")
		So(err, ShouldNotBeNil)
	})
}

// ===== GetByIDs =====

func TestGetByIDs_Success(t *testing.T) {
	Convey("Test GetByIDs success", t, func() {
		rs, mockRA, mockPS, _, mockUMS, _, _ := newTestService(t)
		mockRA.EXPECT().GetByIDs(gomock.Any(), []string{"r1", "r2"}).
			Return([]*interfaces.Resource{{ID: "r1"}, {ID: "r2"}}, nil)
		mockRA.EXPECT().AttachListExtensions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockPS.EXPECT().FilterResources(gomock.Any(), interfaces.AUTH_RESOURCE_TYPE_RESOURCE,
			[]string{"r1", "r2"}, gomock.Any(), true, gomock.Any()).
			Return(map[string]interfaces.PermissionResourceOps{
				"r1": {ResourceID: "r1", Operations: []string{"view_detail"}},
				"r2": {ResourceID: "r2", Operations: []string{"view_detail"}},
			}, nil)
		mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

		resources, err := rs.GetByIDs(context.Background(), []string{"r1", "r2"})
		So(err, ShouldBeNil)
		So(resources, ShouldHaveLength, 2)
	})
}

// ===== GetByCatalogID =====

func TestGetByCatalogID_Success(t *testing.T) {
	Convey("Test GetByCatalogID success", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().GetByCatalogID(gomock.Any(), "cat1").
			Return([]*interfaces.Resource{{ID: "r1", CatalogID: "cat1"}}, nil)

		resources, err := rs.GetByCatalogID(context.Background(), "cat1")
		So(err, ShouldBeNil)
		So(resources, ShouldHaveLength, 1)
	})
}

// ===== List 分页逻辑 =====

func TestList_Pagination(t *testing.T) {
	Convey("Test List pagination", t, func() {
		rs, mockRA, mockPS, _, mockUMS, _, _ := newTestService(t)
		ids := []string{"c1", "c2", "c3", "c4"}
		mockRA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(ids, nil)
		mockPS.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true, gomock.Any()).
			Return(map[string]interfaces.PermissionResourceOps{
				"c1": {ResourceID: "c1"}, "c2": {ResourceID: "c2"}, "c3": {ResourceID: "c3"}, "c4": {ResourceID: "c4"},
			}, nil)
		resources := []*interfaces.Resource{{ID: "r2"}, {ID: "r3"}}
		mockRA.EXPECT().GetByIDsBasic(gomock.Any(), gomock.Any()).Return(resources, nil)
		mockRA.EXPECT().AttachListExtensions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

		result, total, err := rs.List(context.Background(), interfaces.ResourcesQueryParams{
			PaginationQueryParams: interfaces.PaginationQueryParams{Offset: 1, Limit: 2},
		})
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 4)
		So(result, ShouldHaveLength, 2)
		So(result[0].ID, ShouldEqual, "r2")
	})
}

func TestList_ReturnAll(t *testing.T) {
	Convey("Test List return all", t, func() {
		rs, mockRA, mockPS, _, mockUMS, _, _ := newTestService(t)
		ids := []string{"c1", "c2"}
		resources := []*interfaces.Resource{{ID: "r1"}, {ID: "r2"}}
		mockRA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(ids, nil)
		mockPS.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true, gomock.Any()).
			Return(map[string]interfaces.PermissionResourceOps{
				"c1": {ResourceID: "c1"}, "c2": {ResourceID: "c2"},
			}, nil)
		mockRA.EXPECT().GetByIDsBasic(gomock.Any(), gomock.Any()).Return(resources, nil)
		mockRA.EXPECT().AttachListExtensions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

		result, total, err := rs.List(context.Background(), interfaces.ResourcesQueryParams{
			PaginationQueryParams: interfaces.PaginationQueryParams{Limit: -1},
		})
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 2)
		So(result, ShouldHaveLength, 2)
	})
}

func TestList_OffsetBeyondTotal(t *testing.T) {
	Convey("Test List offset beyond total", t, func() {
		rs, mockRA, mockPS, _, _, _, _ := newTestService(t)

		ids := []string{"c1"}
		mockRA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(ids, nil)
		mockPS.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true, gomock.Any()).
			Return(map[string]interfaces.PermissionResourceOps{"c1": {ResourceID: "c1"}}, nil)

		result, total, err := rs.List(context.Background(), interfaces.ResourcesQueryParams{
			PaginationQueryParams: interfaces.PaginationQueryParams{Offset: 10, Limit: 5},
		})
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 1)
		So(result, ShouldBeEmpty)
	})
}

func TestCreate_DatasetCategory(t *testing.T) {
	Convey("Test Create dataset category", t, func() {
		rs, mockRA, mockPS, mockDS, _, mockCS, _ := newTestService(t)
		mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockCS.EXPECT().CheckExistByID(gomock.Any(), gomock.Any()).Return(true, nil)
		mockRA.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		mockDS.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		resource, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
			Name:     "test-dataset",
			Category: interfaces.ResourceCategoryDataset,
		})
		So(err, ShouldBeNil)
		So(resource, ShouldNotBeNil)
	})
}

// ===== DeleteByIDs =====

func TestDeleteByIDs_Empty(t *testing.T) {
	Convey("Test DeleteByIDs empty", t, func() {
		rs, _, _, _, _, _, _ := newTestService(t)
		err := rs.DeleteByIDs(context.Background(), []string{})
		So(err, ShouldBeNil)
	})
}

func TestDeleteByIDs_Success(t *testing.T) {
	Convey("Test DeleteByIDs success", t, func() {
		rs, mockRA, mockPS, _, _, _, mockBTA := newTestService(t)
		mockPS.EXPECT().FilterResources(gomock.Any(), interfaces.AUTH_RESOURCE_TYPE_RESOURCE,
			[]string{"r1"}, gomock.Any(), true, gomock.Any()).
			Return(map[string]interfaces.PermissionResourceOps{
				"r1": {ResourceID: "r1"},
			}, nil)
		mockRA.EXPECT().GetByIDs(gomock.Any(), []string{"r1"}).
			Return([]*interfaces.Resource{{ID: "r1", Category: "table"}}, nil)
		mockRA.EXPECT().DeleteByIDs(gomock.Any(), []string{"r1"}).Return(nil)
		mockPS.EXPECT().DeleteResources(gomock.Any(), interfaces.AUTH_RESOURCE_TYPE_RESOURCE, []string{"r1"}).Return(nil)
		mockBTA.EXPECT().GetByResourceID(gomock.Any(), "r1").Return(nil, nil)
		err := rs.DeleteByIDs(context.Background(), []string{"r1"})
		So(err, ShouldBeNil)
	})
}

// ===== Create =====

func TestCreate_Success(t *testing.T) {
	Convey("Test Create success", t, func() {
		rs, mockRA, mockPS, _, _, mockCS, _ := newTestService(t)
		mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockCS.EXPECT().CheckExistByID(gomock.Any(), gomock.Any()).Return(true, nil)
		mockRA.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		resource, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
			Name:     "test-resource",
			Category: "table",
		})
		So(err, ShouldBeNil)
		So(resource, ShouldNotBeNil)
	})
}

func TestCreate_WithExplicitID(t *testing.T) {
	Convey("Test Create with explicit ID", t, func() {
		rs, mockRA, mockPS, _, _, mockCS, _ := newTestService(t)
		mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockCS.EXPECT().CheckExistByID(gomock.Any(), gomock.Any()).Return(true, nil)
		mockRA.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		resource, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
			ID:       "custom-id",
			Name:     "test-resource",
			Category: "table",
		})
		So(err, ShouldBeNil)
		So(resource, ShouldNotBeNil)
		So(resource.ID, ShouldEqual, "custom-id")
	})
}

func TestCreate_DBError(t *testing.T) {
	Convey("Test Create db error", t, func() {
		rs, mockRA, mockPS, _, _, mockCS, _ := newTestService(t)
		mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockCS.EXPECT().CheckExistByID(gomock.Any(), gomock.Any()).Return(true, nil)
		mockRA.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("db error"))

		_, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
			Name: "test-resource",
		})
		So(err, ShouldNotBeNil)
	})
}

// ===== UpdateStatus =====

func TestUpdateStatus_Success(t *testing.T) {
	Convey("Test UpdateStatus success", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().UpdateStatus(gomock.Any(), "r1", "active", "").Return(nil)

		err := rs.UpdateStatus(context.Background(), "r1", "active", "")
		So(err, ShouldBeNil)
	})
}

func TestUpdateStatus_Error(t *testing.T) {
	Convey("Test UpdateStatus error", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().UpdateStatus(gomock.Any(), "r1", "active", "").
			Return(fmt.Errorf("db error"))

		err := rs.UpdateStatus(context.Background(), "r1", "active", "")
		So(err, ShouldNotBeNil)
	})
}

func TestUpdateResource_PreservesDiscoverStatus(t *testing.T) {
	Convey("Test UpdateResource preserves discover status", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		resource := &interfaces.Resource{ID: "r1", LastDiscoverStatus: interfaces.DiscoverStatusUpdated}
		mockRA.EXPECT().Update(gomock.Any(), resource).Return(nil)

		err := rs.UpdateResource(context.Background(), resource)
		So(err, ShouldBeNil)
		So(resource.LastDiscoverStatus, ShouldEqual, interfaces.DiscoverStatusUpdated)
	})
}

func TestUpdateDiscoverStatus_Success(t *testing.T) {
	Convey("Test UpdateDiscoverStatus success", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().UpdateDiscoverStatus(gomock.Any(), "r1", interfaces.DiscoverStatusUpdated).Return(nil)

		err := rs.UpdateDiscoverStatus(context.Background(), "r1", interfaces.DiscoverStatusUpdated)
		So(err, ShouldBeNil)
	})
}

func TestUpdateDiscoverStatus_Error(t *testing.T) {
	Convey("Test UpdateDiscoverStatus error", t, func() {
		rs, mockRA, _, _, _, _, _ := newTestService(t)
		mockRA.EXPECT().UpdateDiscoverStatus(gomock.Any(), "r1", interfaces.DiscoverStatusUpdated).
			Return(fmt.Errorf("db error"))

		err := rs.UpdateDiscoverStatus(context.Background(), "r1", interfaces.DiscoverStatusUpdated)
		So(err, ShouldNotBeNil)
	})
}

// ===== Update =====

func TestUpdate_NilResource(t *testing.T) {
	Convey("Test Update nil resource", t, func() {
		rs, _, _, _, _, _, _ := newTestService(t)
		err := rs.Update(context.Background(), nil, &interfaces.ResourceRequest{})
		So(err, ShouldNotBeNil)
	})
}

func TestUpdate_Success(t *testing.T) {
	Convey("Test Update success", t, func() {
		rs, mockRA, mockPS, _, _, mockCS, _ := newTestService(t)
		mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockCS.EXPECT().CheckExistByID(gomock.Any(), gomock.Any()).Return(true, nil)
		mockRA.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		err := rs.Update(context.Background(), &interfaces.Resource{ID: "r1", Name: "updated"}, &interfaces.ResourceRequest{
			Name: "updated",
		})
		So(err, ShouldBeNil)
	})
}
