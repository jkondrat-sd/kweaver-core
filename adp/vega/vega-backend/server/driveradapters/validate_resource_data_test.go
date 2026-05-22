package driveradapters

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestValidateResourceDataQueryParams(t *testing.T) {
	Convey("Test ValidateResourceDataQueryParams", t, func() {
		ctx := context.Background()

		Convey("Should apply default format and limit", func() {
			params := &interfaces.ResourceDataQueryParams{}

			err := ValidateResourceDataQueryParams(ctx, params)
			So(err, ShouldBeNil)
			So(params.Format, ShouldEqual, interfaces.Format_Original)
			So(params.Limit, ShouldEqual, interfaces.DEFAULT_DATA_LIMIT)
		})

		Convey("Should reject invalid format", func() {
			params := &interfaces.ResourceDataQueryParams{Format: "compact"}

			err := ValidateResourceDataQueryParams(ctx, params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.InvalidParameter.Format")
		})

		Convey("Should reject negative offset", func() {
			params := &interfaces.ResourceDataQueryParams{Offset: -1}

			err := ValidateResourceDataQueryParams(ctx, params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.InvalidParameter.Offset")
		})

		Convey("Should reject invalid limit", func() {
			params := &interfaces.ResourceDataQueryParams{Limit: interfaces.MAX_SEARCH_SIZE + 1}

			err := ValidateResourceDataQueryParams(ctx, params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.InvalidParameter.Limit")
		})

		Convey("Should reject invalid sort direction", func() {
			params := &interfaces.ResourceDataQueryParams{
				Sort: []*interfaces.SortField{{Field: "name", Direction: "up"}},
			}

			err := ValidateResourceDataQueryParams(ctx, params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.InvalidParameter.Direction")
		})

		Convey("Should decode valid filter condition", func() {
			params := &interfaces.ResourceDataQueryParams{
				FilterCondition: map[string]any{
					"field":     "name",
					"operation": "==",
					"value":     "alice",
				},
			}

			err := ValidateResourceDataQueryParams(ctx, params)
			So(err, ShouldBeNil)
			So(params.FilterCondCfg, ShouldNotBeNil)
			So(params.FilterCondCfg.Name, ShouldEqual, "name")
		})

		Convey("Should reject filter condition without operation", func() {
			params := &interfaces.ResourceDataQueryParams{
				FilterCondition: map[string]any{"field": "name", "value": "alice"},
			}

			err := ValidateResourceDataQueryParams(ctx, params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.NullParameter.FilterConditionOperation")
		})

		Convey("Should reject invalid aggregation function", func() {
			params := &interfaces.ResourceDataQueryParams{
				Aggregation: &interfaces.Aggregation{Property: "amount", Aggr: "median"},
			}

			err := ValidateResourceDataQueryParams(ctx, params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.InvalidParameter.Aggregation")
		})

		Convey("Should reject invalid calendar interval", func() {
			params := &interfaces.ResourceDataQueryParams{
				GroupBy: []*interfaces.GroupByItem{{Property: "created_at", CalendarInterval: "decade"}},
			}

			err := ValidateResourceDataQueryParams(ctx, params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.InvalidParameter.CalendarInterval")
		})

		Convey("Should reject having without aggregation", func() {
			params := &interfaces.ResourceDataQueryParams{
				Having: &interfaces.HavingClause{Field: "__value", Operation: ">", Value: 1},
			}

			err := ValidateResourceDataQueryParams(ctx, params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.InvalidParameter.Having")
		})
	})
}
