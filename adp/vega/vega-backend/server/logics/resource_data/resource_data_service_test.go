// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource_data

import (
	"context"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.uber.org/mock/gomock"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	mock_interfaces "vega-backend/interfaces/mock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPrepareOutputFieldsParams_FiltersUndefinedFields(t *testing.T) {
	Convey("Test prepareOutputFieldsParams filters undefined fields", t, func() {
		rds := &resourceDataService{}
		resource := &interfaces.Resource{
			Category: interfaces.ResourceCategoryTable,
			SchemaDefinition: []*interfaces.Property{
				{Name: "name"},
				{Name: "age"},
			},
		}
		params := &interfaces.ResourceDataQueryParams{
			OutputFields: []string{"name", "missing", "age"},
		}

		rds.prepareOutputFieldsParams(resource, params)

		So(params.OutputFields, ShouldResemble, []string{"name", "age"})
	})
}

func TestPrepareOutputFieldsParams_IndexKeepsScore(t *testing.T) {
	Convey("Test prepareOutputFieldsParams keeps _score for index resources", t, func() {
		rds := &resourceDataService{}
		resource := &interfaces.Resource{
			Category: interfaces.ResourceCategoryIndex,
			SchemaDefinition: []*interfaces.Property{
				{Name: "name"},
			},
		}
		params := &interfaces.ResourceDataQueryParams{
			OutputFields: []string{"name", "_score", "missing"},
		}

		rds.prepareOutputFieldsParams(resource, params)

		So(params.OutputFields, ShouldResemble, []string{"name", "_score"})
	})
}

func TestQueryRejectsDisabledCatalog(t *testing.T) {
	Convey("Test Query rejects disabled catalog", t, func() {
		ctrl := gomock.NewController(t)
		mockCS := mock_interfaces.NewMockCatalogService(ctrl)
		rds := &resourceDataService{cs: mockCS}
		resource := &interfaces.Resource{
			ID:        "resource-1",
			CatalogID: "catalog-1",
			Category:  interfaces.ResourceCategoryTable,
		}
		mockCS.EXPECT().GetByID(gomock.Any(), "catalog-1", true).
			Return(&interfaces.Catalog{ID: "catalog-1", Enabled: false}, nil)

		_, _, err := rds.Query(context.Background(), resource, &interfaces.ResourceDataQueryParams{})
		So(err, ShouldNotBeNil)
		httpErr, ok := err.(*rest.HTTPError)
		So(ok, ShouldBeTrue)
		So(httpErr.HTTPCode, ShouldEqual, http.StatusConflict)
		So(httpErr.BaseError.ErrorCode, ShouldEqual, verrors.VegaBackend_Catalog_IsDisabled)
	})
}
