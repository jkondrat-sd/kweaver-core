package catalog

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/drivenadapters/entityextension"
	"vega-backend/interfaces"
)

func TestCatalogExtensionHelpers(t *testing.T) {
	Convey("Test catalog extension helpers", t, func() {
		Convey("Should prefix column when extension filters exist", func() {
			params := interfaces.CatalogsQueryParams{ExtensionKeys: []string{"env"}}
			So(catalogExtCol(params, "f_update_time"), ShouldEqual, "t_catalog.f_update_time")
			So(catalogListOrderExpr(params), ShouldEqual, "t_catalog.f_update_time ")
		})

		Convey("Should keep column unprefixed without extension filters", func() {
			params := interfaces.CatalogsQueryParams{PaginationQueryParams: interfaces.PaginationQueryParams{
				Sort:      "f_name",
				Direction: "DESC",
			}}
			So(catalogExtCol(params, "f_update_time"), ShouldEqual, "f_update_time")
			So(catalogListOrderExpr(params), ShouldEqual, "f_name DESC")
		})

		Convey("Should apply extension joins", func() {
			params := interfaces.CatalogsQueryParams{
				ExtensionKeys:   []string{"env"},
				ExtensionValues: []string{"prod"},
			}
			builder := sq.Select("t_catalog.f_id").From("t_catalog")

			sqlStr, args, err := applyCatalogExtensionJoins(builder, params).ToSql()
			So(err, ShouldBeNil)
			So(sqlStr, ShouldContainSubstring, "JOIN t_entity_extension vex0")
			So(args, ShouldResemble, []any{entityextension.KindCatalog, "env", "prod"})
		})
	})
}
