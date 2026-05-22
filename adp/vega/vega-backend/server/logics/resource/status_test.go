// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource

import (
	"context"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

func TestEnsureResourceQueryable(t *testing.T) {
	Convey("Test EnsureResourceQueryable", t, func() {
		ctx := context.Background()

		Convey("Nil resource passes", func() {
			w, err := EnsureResourceQueryable(ctx, nil)
			So(err, ShouldBeNil)
			So(w, ShouldBeEmpty)
		})

		Convey("Active resource passes silently", func() {
			w, err := EnsureResourceQueryable(ctx, &interfaces.Resource{ID: "r1", Status: interfaces.ResourceStatusActive})
			So(err, ShouldBeNil)
			So(w, ShouldBeEmpty)
		})

		Convey("Deprecated resource returns a warning", func() {
			w, err := EnsureResourceQueryable(ctx, &interfaces.Resource{ID: "r1", Name: "n1", Status: interfaces.ResourceStatusDeprecated})
			So(err, ShouldBeNil)
			So(w, ShouldNotBeEmpty)
		})

		Convey("Disabled resource blocks query", func() {
			_, err := EnsureResourceQueryable(ctx, &interfaces.Resource{ID: "r1", Status: interfaces.ResourceStatusDisabled})
			SoResourceNotQueryableError(err)
		})

		Convey("Stale resource blocks query", func() {
			_, err := EnsureResourceQueryable(ctx, &interfaces.Resource{ID: "r1", Status: interfaces.ResourceStatusStale})
			SoResourceNotQueryableError(err)
		})

		Convey("Unknown status passes for forward compatibility", func() {
			w, err := EnsureResourceQueryable(ctx, &interfaces.Resource{ID: "r1", Status: "unknown_future_status"})
			So(err, ShouldBeNil)
			So(w, ShouldBeEmpty)
		})
	})
}

func SoResourceNotQueryableError(err error) {
	So(err, ShouldNotBeNil)
	he, ok := err.(*rest.HTTPError)
	So(ok, ShouldBeTrue)
	So(he.HTTPCode, ShouldEqual, http.StatusConflict)
	So(he.BaseError.ErrorCode, ShouldEqual, verrors.VegaBackend_Resource_NotQueryable)
}

func TestEnsureResourcesQueryable(t *testing.T) {
	Convey("Test EnsureResourcesQueryable", t, func() {
		ctx := context.Background()

		Convey("All active produces no warnings", func() {
			ws, err := EnsureResourcesQueryable(ctx, []*interfaces.Resource{
				{ID: "a", Status: interfaces.ResourceStatusActive},
				{ID: "b", Status: interfaces.ResourceStatusActive},
			})
			So(err, ShouldBeNil)
			So(ws, ShouldBeEmpty)
		})

		Convey("Mixed active and deprecated returns deprecated warning", func() {
			ws, err := EnsureResourcesQueryable(ctx, []*interfaces.Resource{
				{ID: "a", Status: interfaces.ResourceStatusActive},
				{ID: "b", Name: "n", Status: interfaces.ResourceStatusDeprecated},
			})
			So(err, ShouldBeNil)
			So(ws, ShouldHaveLength, 1)
		})

		Convey("Any disabled in slice fails fast", func() {
			_, err := EnsureResourcesQueryable(ctx, []*interfaces.Resource{
				{ID: "a", Status: interfaces.ResourceStatusActive},
				{ID: "b", Status: interfaces.ResourceStatusDisabled},
				{ID: "c", Status: interfaces.ResourceStatusActive},
			})
			So(err, ShouldNotBeNil)
		})

		Convey("Stale also blocks", func() {
			_, err := EnsureResourcesQueryable(ctx, []*interfaces.Resource{
				{ID: "a", Status: interfaces.ResourceStatusStale},
			})
			So(err, ShouldNotBeNil)
		})
	})
}
