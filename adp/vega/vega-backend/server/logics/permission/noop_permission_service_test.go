// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package permission

import (
	"context"
	"testing"

	"vega-backend/interfaces"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNoopPermission_CheckPermission(t *testing.T) {
	Convey("Test NoopPermission CheckPermission", t, func() {
		svc := NewNoopPermissionService(nil)
		err := svc.CheckPermission(context.Background(), interfaces.PermissionResource{
			Type: "catalog",
			ID:   "all",
		}, []string{"create"})
		So(err, ShouldBeNil)
	})
}

func TestNoopPermission_CreateResources(t *testing.T) {
	Convey("Test NoopPermission CreateResources", t, func() {
		svc := NewNoopPermissionService(nil)
		err := svc.CreateResources(context.Background(), []interfaces.PermissionResource{
			{ID: "r1", Type: "catalog", Name: "test"},
		}, []string{"view", "modify"})
		So(err, ShouldBeNil)
	})
}

func TestNoopPermission_DeleteResources(t *testing.T) {
	Convey("Test NoopPermission DeleteResources", t, func() {
		svc := NewNoopPermissionService(nil)
		err := svc.DeleteResources(context.Background(), "catalog", []string{"r1", "r2"})
		So(err, ShouldBeNil)
	})
}

func TestNoopPermission_FilterResources(t *testing.T) {
	Convey("Test NoopPermission FilterResources", t, func() {
		svc := NewNoopPermissionService(nil)
		ids := []string{"r1", "r2", "r3"}
		ops := []string{"view_detail", "modify"}

		result, err := svc.FilterResources(context.Background(), "catalog",
			ids, ops, true, interfaces.COMMON_OPERATIONS)
		So(err, ShouldBeNil)
		So(result, ShouldHaveLength, 3)

		r1, ok := result["r1"]
		So(ok, ShouldBeTrue)
		So(r1.ResourceID, ShouldEqual, "r1")
		So(r1.Operations, ShouldHaveLength, len(interfaces.COMMON_OPERATIONS))

		r2, ok := result["r2"]
		So(ok, ShouldBeTrue)
		So(r2.ResourceID, ShouldEqual, "r2")
		So(r2.Operations, ShouldHaveLength, len(interfaces.COMMON_OPERATIONS))

		r3, ok := result["r3"]
		So(ok, ShouldBeTrue)
		So(r3.ResourceID, ShouldEqual, "r3")
		So(r3.Operations, ShouldHaveLength, len(interfaces.COMMON_OPERATIONS))
	})
}

func TestNoopPermission_UpdateResource(t *testing.T) {
	Convey("Test NoopPermission UpdateResource", t, func() {
		svc := NewNoopPermissionService(nil)
		err := svc.UpdateResource(context.Background(), interfaces.PermissionResource{
			ID:   "r1",
			Type: "catalog",
			Name: "updated",
		})
		So(err, ShouldBeNil)
	})
}
