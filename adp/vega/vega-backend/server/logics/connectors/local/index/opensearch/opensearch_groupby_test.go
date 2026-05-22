package opensearch

import (
	"testing"

	"vega-backend/interfaces"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFlattenNestedGroupByRows_TwoDimensions(t *testing.T) {
	Convey("Test flatten nested group by rows with two dimensions", t, func() {
		conn := &OpenSearchConnector{}
		params := &interfaces.ResourceDataQueryParams{
			Limit: 10,
			Aggregation: &interfaces.Aggregation{
				Property: "id",
				Aggr:     "count",
				Alias:    "__value",
			},
			GroupBy: []*interfaces.GroupByItem{
				{Property: "kn_id"},
				{Property: "module_type"},
			},
		}

		rootAgg := map[string]any{
			"buckets": []any{
				map[string]any{
					"key": "yzm_mock_system",
					"group_by_module_type": map[string]any{
						"buckets": []any{
							map[string]any{
								"key": "a",
								"__value": map[string]any{
									"value": float64(3),
								},
							},
							map[string]any{
								"key": "b",
								"__value": map[string]any{
									"value": float64(2),
								},
							},
						},
					},
				},
			},
		}

		rows := conn.flattenNestedGroupByRows(rootAgg, params, "__value")
		So(rows, ShouldHaveLength, 2)
		So(rows[0]["kn_id"], ShouldEqual, "yzm_mock_system")
		So(rows[0]["module_type"], ShouldEqual, "a")
		So(rows[0]["__value"], ShouldEqual, float64(3))
		So(rows[1]["kn_id"], ShouldEqual, "yzm_mock_system")
		So(rows[1]["module_type"], ShouldEqual, "b")
		So(rows[1]["__value"], ShouldEqual, float64(2))
	})
}
