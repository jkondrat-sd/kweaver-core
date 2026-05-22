// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mariadb

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

func TestQuoteColumnName_Simple(t *testing.T) {
	Convey("Test quoteColumnName with simple column", t, func() {
		So(quoteColumnName("name"), ShouldEqual, "`name`")
	})
}

func TestQuoteColumnName_WithAlias(t *testing.T) {
	Convey("Test quoteColumnName with alias", t, func() {
		So(quoteColumnName("t1.name"), ShouldEqual, "`t1`.`name`")
	})
}

func TestQuoteColumnName_Empty(t *testing.T) {
	Convey("Test quoteColumnName with empty column", t, func() {
		So(quoteColumnName(""), ShouldEqual, "``")
	})
}

func TestQuoteColumnName_WithBacktick(t *testing.T) {
	Convey("Test quoteColumnName escapes backticks", t, func() {
		So(quoteColumnName("col`name"), ShouldEqual, "`col``name`")
	})
}

func TestQuoteColumnName_AliasWithSpaces(t *testing.T) {
	Convey("Test quoteColumnName trims alias spaces", t, func() {
		So(quoteColumnName(" t1 . name "), ShouldEqual, "`t1`.`name`")
	})
}

func TestSpecialReplacer(t *testing.T) {
	Convey("Test Special replacer", t, func() {
		Convey("Leaves normal text unchanged", func() {
			So(Special.Replace(`hello`), ShouldEqual, `hello`)
		})

		Convey("Escapes percent", func() {
			So(Special.Replace(`%`), ShouldEqual, `\%`)
		})

		Convey("Escapes underscore", func() {
			So(Special.Replace(`_`), ShouldEqual, `\_`)
		})

		Convey("Escapes quote", func() {
			So(Special.Replace(`'`), ShouldEqual, `\'`)
		})

		Convey("Escapes backslash", func() {
			So(Special.Replace(`\`), ShouldEqual, `\\\\`)
		})

		Convey("Escapes mixed special characters", func() {
			So(Special.Replace(`100%_done`), ShouldEqual, `100\%\_done`)
		})
	})
}

func testFieldsMap() map[string]*interfaces.Property {
	return map[string]*interfaces.Property{
		"name":       {Name: "name", OriginalName: "name", Type: interfaces.DataType_String},
		"age":        {Name: "age", OriginalName: "age", Type: interfaces.DataType_Integer},
		"score":      {Name: "score", OriginalName: "score", Type: interfaces.DataType_Float},
		"created_at": {Name: "created_at", OriginalName: "created_at", Type: interfaces.DataType_Datetime},
		"is_active":  {Name: "is_active", OriginalName: "is_active", Type: interfaces.DataType_Boolean},
		"tags":       {Name: "tags", OriginalName: "tags", Type: interfaces.DataType_Text},
		"alias_col":  {Name: "alias_col", OriginalName: "t1.col", Type: interfaces.DataType_String},
	}
}

func toSQL(connector *MariaDBConnector, cond interfaces.FilterCondition) (string, []interface{}) {
	sqlizer, err := connector.ConvertFilterCondition(context.Background(), cond, testFieldsMap())
	So(err, ShouldBeNil)
	sql, args, err := sqlizer.ToSql()
	So(err, ShouldBeNil)
	return sql, args
}

func mustNewCond(name, op string, value any) interfaces.FilterCondition {
	cfg := &interfaces.FilterCondCfg{
		Name:      name,
		Operation: op,
		ValueOptCfg: interfaces.ValueOptCfg{
			ValueFrom: interfaces.ValueFrom_Const,
			Value:     value,
		},
	}
	cond, err := filter_condition.NewFilterCondition(context.Background(), cfg, testFieldsMap())
	So(err, ShouldBeNil)
	return cond
}

func TestConvertEqual_Const(t *testing.T) {
	Convey("Test convert equal const", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("name", "==", "alice")
		sql, args := toSQL(c, cond)
		So(sql, ShouldEqual, "`name` = ?")
		So(args, ShouldResemble, []interface{}{"alice"})
	})
}

func TestConvertEqual_FieldToField(t *testing.T) {
	Convey("Test convert equal field to field", t, func() {
		c := &MariaDBConnector{}
		cfg := &interfaces.FilterCondCfg{
			Name:      "name",
			Operation: "==",
			ValueOptCfg: interfaces.ValueOptCfg{
				ValueFrom: interfaces.ValueFrom_Field,
				Value:     "tags",
			},
		}
		cond, err := filter_condition.NewFilterCondition(context.Background(), cfg, testFieldsMap())
		So(err, ShouldBeNil)
		sql, _ := toSQL(c, cond)
		So(sql, ShouldEqual, "`name` = `tags`")
	})
}

func TestConvertNotEqual(t *testing.T) {
	Convey("Test convert not equal", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("name", "!=", "bob")
		sql, args := toSQL(c, cond)
		So(sql, ShouldEqual, "`name` <> ?")
		So(args, ShouldResemble, []interface{}{"bob"})
	})
}

func TestConvertGt(t *testing.T) {
	Convey("Test convert greater than", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("age", ">", 18)
		sql, args := toSQL(c, cond)
		So(sql, ShouldEqual, "`age` > ?")
		So(args, ShouldResemble, []interface{}{18})
	})
}

func TestConvertGte(t *testing.T) {
	Convey("Test convert greater than or equal", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("age", ">=", 18)
		sql, _ := toSQL(c, cond)
		So(sql, ShouldEqual, "`age` >= ?")
	})
}

func TestConvertLt(t *testing.T) {
	Convey("Test convert less than", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("age", "<", 65)
		sql, _ := toSQL(c, cond)
		So(sql, ShouldEqual, "`age` < ?")
	})
}

func TestConvertLte(t *testing.T) {
	Convey("Test convert less than or equal", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("age", "<=", 65)
		sql, _ := toSQL(c, cond)
		So(sql, ShouldEqual, "`age` <= ?")
	})
}

func TestConvertIn(t *testing.T) {
	Convey("Test convert in", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("name", "in", []any{"alice", "bob"})
		sql, args := toSQL(c, cond)
		So(sql, ShouldEqual, "`name` IN (?,?)")
		So(args, ShouldHaveLength, 2)
	})
}

func TestConvertNotIn(t *testing.T) {
	Convey("Test convert not in", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("name", "not_in", []any{"alice"})
		sql, _ := toSQL(c, cond)
		So(sql, ShouldEqual, "`name` NOT IN (?)")
	})
}

func TestConvertLike(t *testing.T) {
	Convey("Test convert like", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("name", "like", "ali")
		sql, args := toSQL(c, cond)
		So(sql, ShouldEqual, "`name` LIKE ?")
		So(args, ShouldResemble, []interface{}{"%ali%"})
	})
}

func TestConvertLike_SpecialChars(t *testing.T) {
	Convey("Test convert like escapes special characters", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("name", "like", "100%")
		_, args := toSQL(c, cond)
		argStr, ok := args[0].(string)
		So(ok, ShouldBeTrue)
		So(argStr, ShouldContainSubstring, `\%`)
	})
}

func TestConvertNull(t *testing.T) {
	Convey("Test convert null", t, func() {
		c := &MariaDBConnector{}
		cfg := &interfaces.FilterCondCfg{Name: "name", Operation: "null"}
		cond, err := filter_condition.NewFilterCondition(context.Background(), cfg, testFieldsMap())
		So(err, ShouldBeNil)
		sql, _ := toSQL(c, cond)
		So(sql, ShouldEqual, "`name` IS NULL")
	})
}

func TestConvertNotNull(t *testing.T) {
	Convey("Test convert not null", t, func() {
		c := &MariaDBConnector{}
		cfg := &interfaces.FilterCondCfg{Name: "name", Operation: "not_null"}
		cond, err := filter_condition.NewFilterCondition(context.Background(), cfg, testFieldsMap())
		So(err, ShouldBeNil)
		sql, _ := toSQL(c, cond)
		So(sql, ShouldEqual, "`name` IS NOT NULL")
	})
}

func TestConvertEmpty(t *testing.T) {
	Convey("Test convert empty", t, func() {
		c := &MariaDBConnector{}
		cfg := &interfaces.FilterCondCfg{Name: "name", Operation: "empty"}
		cond, err := filter_condition.NewFilterCondition(context.Background(), cfg, testFieldsMap())
		So(err, ShouldBeNil)
		sql, args := toSQL(c, cond)
		So(sql, ShouldEqual, "`name` = ?")
		So(args, ShouldResemble, []interface{}{""})
	})
}

func TestConvertRange(t *testing.T) {
	Convey("Test convert range", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("age", "range", []any{18, 65})
		sql, args := toSQL(c, cond)
		So(sql, ShouldContainSubstring, "`age` >= ?")
		So(sql, ShouldContainSubstring, "`age` <= ?")
		So(args, ShouldHaveLength, 2)
	})
}

func TestConvertRegex(t *testing.T) {
	Convey("Test convert regex", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("name", "regex", "^ali.*")
		sql, args := toSQL(c, cond)
		So(sql, ShouldEqual, "`name` REGEXP ?")
		So(args, ShouldResemble, []interface{}{"^ali.*"})
	})
}

func TestConvertTrue(t *testing.T) {
	Convey("Test convert true", t, func() {
		c := &MariaDBConnector{}
		cfg := &interfaces.FilterCondCfg{Name: "is_active", Operation: "true"}
		cond, err := filter_condition.NewFilterCondition(context.Background(), cfg, testFieldsMap())
		So(err, ShouldBeNil)
		sql, args := toSQL(c, cond)
		So(sql, ShouldEqual, "`is_active` = ?")
		So(args, ShouldResemble, []interface{}{true})
	})
}

func TestConvertPrefix(t *testing.T) {
	Convey("Test convert prefix", t, func() {
		c := &MariaDBConnector{}
		cond := mustNewCond("name", "prefix", "ali")
		sql, args := toSQL(c, cond)
		So(sql, ShouldEqual, "`name` LIKE ?")
		So(args, ShouldResemble, []interface{}{"ali%"})
	})
}

func TestConvertAnd(t *testing.T) {
	Convey("Test convert and", t, func() {
		c := &MariaDBConnector{}
		cfg := &interfaces.FilterCondCfg{
			Operation: "and",
			SubConds: []*interfaces.FilterCondCfg{
				{Name: "name", Operation: "==", ValueOptCfg: interfaces.ValueOptCfg{ValueFrom: "const", Value: "alice"}},
				{Name: "age", Operation: ">", ValueOptCfg: interfaces.ValueOptCfg{ValueFrom: "const", Value: 18}},
			},
		}
		cond, err := filter_condition.NewFilterCondition(context.Background(), cfg, testFieldsMap())
		So(err, ShouldBeNil)
		sql, args := toSQL(c, cond)
		So(sql, ShouldContainSubstring, "`name` = ?")
		So(sql, ShouldContainSubstring, "`age` > ?")
		So(sql, ShouldContainSubstring, " AND ")
		So(args, ShouldHaveLength, 2)
	})
}

func TestConvertOr(t *testing.T) {
	Convey("Test convert or", t, func() {
		c := &MariaDBConnector{}
		cfg := &interfaces.FilterCondCfg{
			Operation: "or",
			SubConds: []*interfaces.FilterCondCfg{
				{Name: "name", Operation: "==", ValueOptCfg: interfaces.ValueOptCfg{ValueFrom: "const", Value: "alice"}},
				{Name: "name", Operation: "==", ValueOptCfg: interfaces.ValueOptCfg{ValueFrom: "const", Value: "bob"}},
			},
		}
		cond, err := filter_condition.NewFilterCondition(context.Background(), cfg, testFieldsMap())
		So(err, ShouldBeNil)
		sql, _ := toSQL(c, cond)
		So(sql, ShouldContainSubstring, " OR ")
	})
}

func TestConvertEqual_AliasColumn(t *testing.T) {
	Convey("Test convert equal alias column", t, func() {
		c := &MariaDBConnector{}
		cfg := &interfaces.FilterCondCfg{
			Name:      "alias_col",
			Operation: "==",
			ValueOptCfg: interfaces.ValueOptCfg{
				ValueFrom: interfaces.ValueFrom_Const,
				Value:     "test",
			},
		}
		cond, err := filter_condition.NewFilterCondition(context.Background(), cfg, testFieldsMap())
		So(err, ShouldBeNil)
		sql, _ := toSQL(c, cond)
		So(sql, ShouldEqual, "`t1`.`col` = ?")
	})
}
