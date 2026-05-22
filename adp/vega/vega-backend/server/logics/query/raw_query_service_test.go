// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package query

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

// NewRawQueryServiceWithDeps 创建SQL查询服务（用于测试）
func NewRawQueryServiceWithDeps(cs interfaces.CatalogService, rs interfaces.ResourceService) interfaces.RawQueryService {
	return &rawQueryService{cs: cs, rs: rs}
}

func TestExecuteRejectsDisabledCatalogForOpenSearchQuery(t *testing.T) {
	Convey("Test Execute rejects disabled catalog for OpenSearch query", t, func() {
		ctrl := gomock.NewController(t)
		mockCS := mock_interfaces.NewMockCatalogService(ctrl)
		mockRS := mock_interfaces.NewMockResourceService(ctrl)
		service := NewRawQueryServiceWithDeps(mockCS, mockRS)

		mockRS.EXPECT().GetByID(gomock.Any(), "resource-1").
			Return(&interfaces.Resource{ID: "resource-1", CatalogID: "catalog-1"}, nil)
		mockCS.EXPECT().GetByID(gomock.Any(), "catalog-1", true).
			Return(&interfaces.Catalog{ID: "catalog-1", Enabled: false, ConnectorType: interfaces.ConnectorTypeOpenSearch}, nil)

		_, err := service.Execute(context.Background(), &interfaces.RawQueryRequest{
			ResourceType: interfaces.ConnectorTypeOpenSearch,
			Query:        map[string]any{"resource_id": "resource-1"},
		})
		assertCatalogDisabledError(err)
	})
}

func TestExecuteRejectsDisabledCatalogForExistingStreamSession(t *testing.T) {
	Convey("Test Execute rejects disabled catalog for existing stream session", t, func() {
		ctrl := gomock.NewController(t)
		mockCS := mock_interfaces.NewMockCatalogService(ctrl)
		mockRS := mock_interfaces.NewMockResourceService(ctrl)
		service := NewRawQueryServiceWithDeps(mockCS, mockRS)

		session, err := GetStreamQueryManager().CreateSession(
			interfaces.ConnectorTypeMariaDB,
			"catalog",
			"catalog-1",
			&interfaces.Catalog{ID: "catalog-1", Enabled: true, ConnectorType: interfaces.ConnectorTypeMariaDB},
			100,
			"select * from {{resource-1}}",
			[]string{"resource-1"},
		)
		So(err, ShouldBeNil)
		defer GetStreamQueryManager().RemoveSession(session.QueryID)

		mockCS.EXPECT().GetByID(gomock.Any(), "catalog-1", true).
			Return(&interfaces.Catalog{ID: "catalog-1", Enabled: false, ConnectorType: interfaces.ConnectorTypeMariaDB}, nil)

		_, err = service.Execute(context.Background(), &interfaces.RawQueryRequest{
			QueryType: interfaces.QueryType_Stream,
			QueryID:   session.QueryID,
		})
		assertCatalogDisabledError(err)
	})
}

func assertCatalogDisabledError(err error) {
	So(err, ShouldNotBeNil)
	httpErr, ok := err.(*rest.HTTPError)
	So(ok, ShouldBeTrue)
	So(httpErr.HTTPCode, ShouldEqual, http.StatusConflict)
	So(httpErr.BaseError.ErrorCode, ShouldEqual, verrors.VegaBackend_Catalog_IsDisabled)
}
