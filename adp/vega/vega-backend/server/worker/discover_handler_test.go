// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package worker

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"vega-backend/interfaces"
	vmock "vega-backend/interfaces/mock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReconcileTableResourcesMarksNew(t *testing.T) {
	Convey("Test reconcileTableResources marks new resources", t, func() {
		ctrl := gomock.NewController(t)
		rs := vmock.NewMockResourceService(ctrl)
		dh := &DiscoverHandler{rs: rs}

		table := &interfaces.TableMeta{Name: "users"}
		created := &interfaces.Resource{ID: "r1", SourceIdentifier: "users", Status: interfaces.ResourceStatusActive}
		rs.EXPECT().Create(gomock.Any(), gomock.Any()).Return(created, nil)
		rs.EXPECT().UpdateDiscoverStatus(gomock.Any(), "r1", interfaces.DiscoverStatusNew).Return(nil)
		actions := interfaces.ActionsFromDiscoverStrategy(interfaces.DiscoverStrategyFullSync)

		result, items, err := dh.reconcileTableResources(context.Background(), &interfaces.Catalog{ID: "cat1"},
			[]*interfaces.TableMeta{table}, nil, &actions)
		So(err, ShouldBeNil)
		So(result.NewCount, ShouldEqual, 1)
		So(items, ShouldHaveLength, 1)
		So(items[0].markAfterEnrich, ShouldBeFalse)
	})
}

func TestReconcileTableResourcesRefreshesMissingWhenAlreadyStale(t *testing.T) {
	Convey("Test reconcileTableResources refreshes missing status for already stale resources", t, func() {
		ctrl := gomock.NewController(t)
		rs := vmock.NewMockResourceService(ctrl)
		dh := &DiscoverHandler{rs: rs}

		rs.EXPECT().UpdateDiscoverStatus(gomock.Any(), "r1", interfaces.DiscoverStatusMissing).Return(nil)
		actions := interfaces.ActionsFromDiscoverStrategy(interfaces.DiscoverStrategyFullSync)

		result, _, err := dh.reconcileTableResources(context.Background(), &interfaces.Catalog{ID: "cat1"}, nil,
			[]*interfaces.Resource{{
				ID:               "r1",
				SourceIdentifier: "users",
				Category:         interfaces.ResourceCategoryTable,
				Status:           interfaces.ResourceStatusStale,
			}}, &actions)
		So(err, ShouldBeNil)
		So(result.StaleCount, ShouldEqual, 0)
	})
}

func TestReconcileTableResourcesDoesNotDisableUserDisabledResource(t *testing.T) {
	Convey("Test reconcileTableResources does not stale user-disabled resources", t, func() {
		ctrl := gomock.NewController(t)
		rs := vmock.NewMockResourceService(ctrl)
		dh := &DiscoverHandler{rs: rs}

		rs.EXPECT().UpdateDiscoverStatus(gomock.Any(), "r1", interfaces.DiscoverStatusMissing).Return(nil)
		actions := interfaces.ActionsFromDiscoverStrategy(interfaces.DiscoverStrategyFullSync)

		result, _, err := dh.reconcileTableResources(context.Background(), &interfaces.Catalog{ID: "cat1"}, nil,
			[]*interfaces.Resource{{
				ID:               "r1",
				SourceIdentifier: "users",
				Category:         interfaces.ResourceCategoryTable,
				Status:           interfaces.ResourceStatusDisabled,
			}}, &actions)
		So(err, ShouldBeNil)
		So(result.StaleCount, ShouldEqual, 0)
	})
}

func TestReconcileTableResourcesMarksRestored(t *testing.T) {
	Convey("Test reconcileTableResources marks restored resources", t, func() {
		ctrl := gomock.NewController(t)
		rs := vmock.NewMockResourceService(ctrl)
		dh := &DiscoverHandler{rs: rs}

		rs.EXPECT().UpdateStatus(gomock.Any(), "r1", interfaces.ResourceStatusActive, "").Return(nil)
		rs.EXPECT().UpdateDiscoverStatus(gomock.Any(), "r1", interfaces.DiscoverStatusRestored).Return(nil)
		actions := interfaces.ActionsFromDiscoverStrategy(interfaces.DiscoverStrategyFullSync)

		result, items, err := dh.reconcileTableResources(context.Background(), &interfaces.Catalog{ID: "cat1"},
			[]*interfaces.TableMeta{{Name: "users"}},
			[]*interfaces.Resource{{
				ID:               "r1",
				SourceIdentifier: "users",
				Category:         interfaces.ResourceCategoryTable,
				Status:           interfaces.ResourceStatusStale,
			}}, &actions)
		So(err, ShouldBeNil)
		So(items, ShouldHaveLength, 1)
		So(items[0].markAfterEnrich, ShouldBeFalse)
		So(result.RestoredCount, ShouldEqual, 1)
		So(result.UnchangedCount, ShouldEqual, 0)
	})
}

func TestUpdateDiscoverResultForEnrichStatus(t *testing.T) {
	Convey("Test updateDiscoverResultForEnrichStatus", t, func() {
		result := &interfaces.DiscoverResult{}

		updateDiscoverResultForEnrichStatus(result, interfaces.DiscoverStatusUnchanged)
		updateDiscoverResultForEnrichStatus(result, interfaces.DiscoverStatusUpdated)

		So(result.UnchangedCount, ShouldEqual, 1)
		So(result.UpdatedCount, ShouldEqual, 1)
	})
}

func TestSourceSnapshotHashIgnoresDerivedAndUserEditableFields(t *testing.T) {
	Convey("Test sourceSnapshotHash ignores derived and user-editable fields", t, func() {
		resource := &interfaces.Resource{
			Description:      "user text",
			Tags:             []string{"a"},
			Name:             "users",
			SchemaDefinition: []*interfaces.Property{{Name: "id", Type: "int", Description: "derived"}},
			SourceMetadata:   map[string]any{"original_name": "users"},
		}
		before := sourceSnapshotHash(resource)

		resource.Description = "edited by user"
		resource.Tags = []string{"b"}
		resource.Name = "display name"
		resource.SchemaDefinition = append(resource.SchemaDefinition, &interfaces.Property{Name: "name", Type: "string"})

		So(sourceSnapshotHash(resource), ShouldEqual, before)
	})
}

func TestSourceSnapshotHashChangesForSourceMetadata(t *testing.T) {
	Convey("Test sourceSnapshotHash changes for source metadata", t, func() {
		resource := &interfaces.Resource{
			SchemaDefinition: []*interfaces.Property{{Name: "id", Type: "int"}},
			SourceMetadata:   map[string]any{"original_name": "users", "columns": []interfaces.TableColumnMeta{{Name: "id", Type: "int"}}},
		}
		before := sourceSnapshotHash(resource)

		resource.SourceMetadata["columns"] = []interfaces.TableColumnMeta{{Name: "id", Type: "int"}, {Name: "name", Type: "varchar"}}

		So(sourceSnapshotHash(resource), ShouldNotEqual, before)
	})
}
