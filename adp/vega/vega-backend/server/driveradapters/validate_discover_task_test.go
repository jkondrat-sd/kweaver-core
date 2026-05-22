// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestValidateDiscoverTaskQueryParams(t *testing.T) {
	Convey("Test ValidateDiscoverTaskQueryParams", t, func() {
		ctx := context.Background()

		Convey("Valid empty params", func() {
			err := ValidateDiscoverTaskQueryParams(ctx, interfaces.DiscoverTaskQueryParams{})
			So(err, ShouldBeNil)
		})

		Convey("Valid status and trigger type", func() {
			err := ValidateDiscoverTaskQueryParams(ctx, interfaces.DiscoverTaskQueryParams{
				Status:      interfaces.DiscoverTaskStatusCompleted,
				TriggerType: interfaces.DiscoverTaskTriggerScheduled,
			})
			So(err, ShouldBeNil)
		})

		Convey("Invalid status", func() {
			err := ValidateDiscoverTaskQueryParams(ctx, interfaces.DiscoverTaskQueryParams{
				Status: "unknown",
			})
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid trigger type", func() {
			err := ValidateDiscoverTaskQueryParams(ctx, interfaces.DiscoverTaskQueryParams{
				TriggerType: "unknown",
			})
			So(err, ShouldNotBeNil)
		})
	})
}
