// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsValidDiscoverStrategy(t *testing.T) {
	Convey("Test IsValidDiscoverStrategy", t, func() {
		Convey("Full sync strategy is valid", func() {
			So(IsValidDiscoverStrategy(DiscoverStrategyFullSync), ShouldBeTrue)
		})

		Convey("Create only strategy is valid", func() {
			So(IsValidDiscoverStrategy(DiscoverStrategyCreateOnly), ShouldBeTrue)
		})

		Convey("Cleanup only strategy is valid", func() {
			So(IsValidDiscoverStrategy(DiscoverStrategyCleanupOnly), ShouldBeTrue)
		})

		Convey("Unknown strategy is invalid", func() {
			So(IsValidDiscoverStrategy("unknown"), ShouldBeFalse)
		})

		Convey("Empty strategy is invalid before driver normalization", func() {
			So(IsValidDiscoverStrategy(""), ShouldBeFalse)
		})
	})
}

func TestActionsFromDiscoverStrategy(t *testing.T) {
	Convey("Test ActionsFromDiscoverStrategy", t, func() {
		Convey("Full sync enables create, refresh, and mark stale", func() {
			actions := ActionsFromDiscoverStrategy(DiscoverStrategyFullSync)
			So(actions.Create, ShouldBeTrue)
			So(actions.Refresh, ShouldBeTrue)
			So(actions.MarkStale, ShouldBeTrue)
		})

		Convey("Empty strategy defaults to full sync", func() {
			actions := ActionsFromDiscoverStrategy("")
			So(actions.Create, ShouldBeTrue)
			So(actions.Refresh, ShouldBeTrue)
			So(actions.MarkStale, ShouldBeTrue)
		})

		Convey("Create only enables create", func() {
			actions := ActionsFromDiscoverStrategy(DiscoverStrategyCreateOnly)
			So(actions.Create, ShouldBeTrue)
			So(actions.Refresh, ShouldBeFalse)
			So(actions.MarkStale, ShouldBeFalse)
		})

		Convey("Cleanup only enables mark stale", func() {
			actions := ActionsFromDiscoverStrategy(DiscoverStrategyCleanupOnly)
			So(actions.Create, ShouldBeFalse)
			So(actions.Refresh, ShouldBeFalse)
			So(actions.MarkStale, ShouldBeTrue)
		})
	})
}
