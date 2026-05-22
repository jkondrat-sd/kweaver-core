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

func TestValidateDiscoverScheduleRequest(t *testing.T) {
	Convey("Test ValidateDiscoverScheduleRequest", t, func() {
		validReq := func() *interfaces.DiscoverScheduleRequest {
			return &interfaces.DiscoverScheduleRequest{
				Name:      "schedule-1",
				CatalogID: "catalog-1",
				CronExpr:  "*/5 * * * *",
				StartTime: 1000,
				EndTime:   2000,
				Strategy:  interfaces.DiscoverStrategyFullSync,
			}
		}

		Convey("Valid request", func() {
			err := ValidateDiscoverScheduleRequest(context.Background(), validReq())
			So(err, ShouldBeNil)
		})

		Convey("Missing name", func() {
			req := validReq()
			req.Name = ""
			err := ValidateDiscoverScheduleRequest(context.Background(), req)
			So(err, ShouldNotBeNil)
		})

		Convey("Missing catalog ID", func() {
			req := validReq()
			req.CatalogID = ""
			err := ValidateDiscoverScheduleRequest(context.Background(), req)
			So(err, ShouldNotBeNil)
		})

		Convey("Missing cron expression", func() {
			req := validReq()
			req.CronExpr = ""
			err := ValidateDiscoverScheduleRequest(context.Background(), req)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid cron expression", func() {
			req := validReq()
			req.CronExpr = "invalid"
			err := ValidateDiscoverScheduleRequest(context.Background(), req)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid strategy", func() {
			req := validReq()
			req.Strategy = "unknown"
			err := ValidateDiscoverScheduleRequest(context.Background(), req)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid time range", func() {
			req := validReq()
			req.StartTime = 2000
			req.EndTime = 1000
			err := ValidateDiscoverScheduleRequest(context.Background(), req)
			So(err, ShouldNotBeNil)
		})

		Convey("Valid time range without end time", func() {
			req := validReq()
			req.StartTime = 0
			req.EndTime = 0
			err := ValidateDiscoverScheduleRequest(context.Background(), req)
			So(err, ShouldBeNil)
		})

		Convey("Negative start time", func() {
			req := validReq()
			req.StartTime = -1
			err := ValidateDiscoverScheduleRequest(context.Background(), req)
			So(err, ShouldNotBeNil)
		})

		Convey("Negative end time", func() {
			req := validReq()
			req.EndTime = -1
			err := ValidateDiscoverScheduleRequest(context.Background(), req)
			So(err, ShouldNotBeNil)
		})
	})
}
