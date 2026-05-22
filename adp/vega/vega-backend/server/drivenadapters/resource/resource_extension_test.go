package resource

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/drivenadapters/entityextension"
	"vega-backend/interfaces"
)

func TestResourceExtensionHelpers(t *testing.T) {
	Convey("Test resource extension helpers", t, func() {
		Convey("prefixes column when extension filters exist", func() {
			params := interfaces.ResourcesQueryParams{ExtensionKeys: []string{"env"}}
			So(resourceExtCol(params, "f_update_time"), ShouldEqual, "t_resource.f_update_time")
			So(resourceListOrderExpr(params), ShouldEqual, "t_resource.f_update_time ")
		})

		Convey("keeps column unprefixed without extension filters", func() {
			params := interfaces.ResourcesQueryParams{PaginationQueryParams: interfaces.PaginationQueryParams{
				Sort:      "f_name",
				Direction: "DESC",
			}}
			So(resourceExtCol(params, "f_update_time"), ShouldEqual, "f_update_time")
			So(resourceListOrderExpr(params), ShouldEqual, "f_name DESC")
		})

		Convey("applies extension joins", func() {
			params := interfaces.ResourcesQueryParams{
				ExtensionKeys:   []string{"env"},
				ExtensionValues: []string{"prod"},
			}
			builder := sq.Select("t_resource.f_id").From("t_resource")

			sqlStr, args, err := applyResourceExtensionJoins(builder, params).ToSql()
			So(err, ShouldBeNil)
			So(sqlStr, ShouldContainSubstring, "JOIN t_entity_extension vex0")
			So(args, ShouldResemble, []any{entityextension.KindResource, "env", "prod"})
		})
	})
}
