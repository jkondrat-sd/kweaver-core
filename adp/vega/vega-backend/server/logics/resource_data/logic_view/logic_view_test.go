// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logic_view

import (
	"context"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

func TestLogicViewService_Query_RejectsUnsupportedLogicType(t *testing.T) {
	Convey("Test LogicViewService Query rejects unsupported logic type", t, func() {
		svc := &logicViewService{}
		rows, total, err := svc.Query(context.Background(), &interfaces.Resource{
			ID:        "view-1",
			LogicType: "unsupported",
		}, &interfaces.ResourceDataQueryParams{})
		So(rows, ShouldBeNil)
		So(total, ShouldEqual, 0)
		So(err, ShouldNotBeNil)
		httpErr, ok := err.(*rest.HTTPError)
		So(ok, ShouldBeTrue)
		So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		So(httpErr.BaseError.ErrorCode, ShouldEqual, verrors.VegaBackend_Resource_InternalError_InvalidCategory)
	})
}

func TestExecutePhysicalQuery_RejectsUnsupportedCategory(t *testing.T) {
	Convey("Test executePhysicalQuery rejects unsupported category", t, func() {
		rows, total, err := executePhysicalQuery(context.Background(), &interfaces.Catalog{}, &interfaces.Resource{
			ID:       "resource-1",
			Category: "unsupported",
		}, &interfaces.ResourceDataQueryParams{})
		So(rows, ShouldBeNil)
		So(total, ShouldEqual, 0)
		So(err, ShouldNotBeNil)
		httpErr, ok := err.(*rest.HTTPError)
		So(ok, ShouldBeTrue)
		So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		So(httpErr.BaseError.ErrorCode, ShouldEqual, verrors.VegaBackend_Resource_InternalError_InvalidCategory)
	})
}
