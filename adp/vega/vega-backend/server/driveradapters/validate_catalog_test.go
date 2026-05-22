package driveradapters

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestValidateCatalogRequest(t *testing.T) {
	Convey("Test ValidateCatalogRequest", t, func() {
		ctx := context.Background()

		Convey("Should validate catalog request", func() {
			req := &interfaces.CatalogRequest{
				ID:          "catalog-1",
				Name:        "catalog",
				Tags:        []string{"prod"},
				Description: "catalog desc",
				ConnectorCfg: interfaces.ConnectorConfig{
					"databases": []any{"db1", "db2"},
					"schemas":   []any{"public"},
				},
			}

			err := ValidateCatalogRequest(ctx, req)
			So(err, ShouldBeNil)
		})

		Convey("Should reject invalid id", func() {
			err := ValidateCatalogRequest(ctx, &interfaces.CatalogRequest{ID: "_invalid", Name: "catalog"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.InvalidParameter.ID")
		})

		Convey("Should reject empty name", func() {
			err := ValidateCatalogRequest(ctx, &interfaces.CatalogRequest{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.InvalidParameter.Name")
		})

		Convey("Should reject too long description", func() {
			err := ValidateCatalogRequest(ctx, &interfaces.CatalogRequest{
				Name:        "catalog",
				Description: strings.Repeat("a", interfaces.DESCRIPTION_MAX_LENGTH+1),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.InvalidParameter.Description")
		})

		Convey("Should reject duplicated databases in connector config", func() {
			err := ValidateCatalogRequest(ctx, &interfaces.CatalogRequest{
				Name: "catalog",
				ConnectorCfg: interfaces.ConnectorConfig{
					"databases": []any{"db1", "db1"},
				},
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "duplicate element found in 'databases'")
		})
	})
}

func TestValidateCatalogListQueryParams(t *testing.T) {
	Convey("Test ValidateCatalogListQueryParams", t, func() {
		ctx := context.Background()

		Convey("Should accept valid params", func() {
			err := ValidateCatalogListQueryParams(ctx, interfaces.CatalogsQueryParams{
				Type:              interfaces.CatalogTypePhysical,
				HealthCheckStatus: interfaces.CatalogHealthStatusHealthy,
			})
			So(err, ShouldBeNil)
		})

		Convey("Should reject invalid catalog type", func() {
			err := ValidateCatalogListQueryParams(ctx, interfaces.CatalogsQueryParams{Type: "unknown"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid type: unknown")
		})

		Convey("Should reject invalid health check status", func() {
			err := ValidateCatalogListQueryParams(ctx, interfaces.CatalogsQueryParams{HealthCheckStatus: "unknown"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid health_check_status: unknown")
		})
	})
}
