package driveradapters

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestValidateConnectorTypeReq(t *testing.T) {
	Convey("Test ValidateConnectorTypeReq", t, func() {
		ctx := context.Background()

		Convey("Should validate local connector type", func() {
			err := ValidateConnectorTypeReq(ctx, &interfaces.ConnectorTypeReq{
				Mode:     interfaces.ConnectorModeLocal,
				Category: interfaces.ConnectorCategoryTable,
			})
			So(err, ShouldBeNil)
		})

		Convey("Should validate remote connector type with endpoint", func() {
			err := ValidateConnectorTypeReq(ctx, &interfaces.ConnectorTypeReq{
				Mode:     interfaces.ConnectorModeRemote,
				Category: interfaces.ConnectorCategoryAPI,
				Endpoint: "http://connector",
			})
			So(err, ShouldBeNil)
		})

		Convey("Should reject empty mode", func() {
			err := ValidateConnectorTypeReq(ctx, &interfaces.ConnectorTypeReq{Category: interfaces.ConnectorCategoryTable})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.ConnectorType.InvalidParameter.Mode")
		})

		Convey("Should reject invalid mode", func() {
			err := ValidateConnectorTypeReq(ctx, &interfaces.ConnectorTypeReq{Mode: "invalid", Category: interfaces.ConnectorCategoryTable})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid mode: invalid")
		})

		Convey("Should reject empty category", func() {
			err := ValidateConnectorTypeReq(ctx, &interfaces.ConnectorTypeReq{Mode: interfaces.ConnectorModeLocal})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.ConnectorType.InvalidParameter.Category")
		})

		Convey("Should reject invalid category", func() {
			err := ValidateConnectorTypeReq(ctx, &interfaces.ConnectorTypeReq{Mode: interfaces.ConnectorModeLocal, Category: "invalid"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid category: invalid")
		})

		Convey("Should reject remote connector type without endpoint", func() {
			err := ValidateConnectorTypeReq(ctx, &interfaces.ConnectorTypeReq{
				Mode:     interfaces.ConnectorModeRemote,
				Category: interfaces.ConnectorCategoryAPI,
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Remote connector requires endpoint URL")
		})
	})
}

func TestValidateConnectorTypeListQueryParams(t *testing.T) {
	Convey("Test ValidateConnectorTypeListQueryParams", t, func() {
		ctx := context.Background()

		Convey("Should accept empty optional params", func() {
			err := ValidateConnectorTypeListQueryParams(ctx, interfaces.ConnectorTypesQueryParams{})
			So(err, ShouldBeNil)
		})

		Convey("Should accept valid optional params", func() {
			err := ValidateConnectorTypeListQueryParams(ctx, interfaces.ConnectorTypesQueryParams{
				Mode:     interfaces.ConnectorModeRemote,
				Category: interfaces.ConnectorCategoryAPI,
			})
			So(err, ShouldBeNil)
		})

		Convey("Should reject invalid optional mode", func() {
			err := ValidateConnectorTypeListQueryParams(ctx, interfaces.ConnectorTypesQueryParams{Mode: "invalid"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid mode: invalid")
		})

		Convey("Should reject invalid optional category", func() {
			err := ValidateConnectorTypeListQueryParams(ctx, interfaces.ConnectorTypesQueryParams{Category: "invalid"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid category: invalid")
		})
	})
}
