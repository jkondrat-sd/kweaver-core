// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package filter_condition

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func testFieldsMap() map[string]*interfaces.Property {
	return map[string]*interfaces.Property{
		"name":       {Name: "name", Type: interfaces.DataType_String},
		"age":        {Name: "age", Type: interfaces.DataType_Integer},
		"score":      {Name: "score", Type: interfaces.DataType_Float},
		"created_at": {Name: "created_at", Type: interfaces.DataType_Datetime},
		"is_active":  {Name: "is_active", Type: interfaces.DataType_Boolean},
		"tags":       {Name: "tags", Type: interfaces.DataType_Text},
		"other_name": {Name: "other_name", Type: interfaces.DataType_String},
	}
}

func constCfg(name, op string, value any) *interfaces.FilterCondCfg {
	return &interfaces.FilterCondCfg{
		Name:      name,
		Operation: op,
		ValueOptCfg: interfaces.ValueOptCfg{
			ValueFrom: interfaces.ValueFrom_Const,
			Value:     value,
		},
	}
}

func newTestCondition(cfg *interfaces.FilterCondCfg) (interfaces.FilterCondition, error) {
	return NewFilterCondition(context.Background(), cfg, testFieldsMap())
}

func TestNewFilterCondition_NilConfig(t *testing.T) {
	Convey("Test NewFilterCondition with nil config", t, func() {
		cond, err := newTestCondition(nil)
		So(err, ShouldBeNil)
		So(cond, ShouldBeNil)
	})
}

func TestNewFilterCondition_EmptyConfig(t *testing.T) {
	Convey("Test NewFilterCondition with empty config", t, func() {
		cond, err := newTestCondition(&interfaces.FilterCondCfg{})
		So(err, ShouldBeNil)
		So(cond, ShouldBeNil)
	})
}

func TestNewFilterCondition_UnsupportedOperation(t *testing.T) {
	Convey("Test NewFilterCondition rejects unsupported operation", t, func() {
		_, err := newTestCondition(constCfg("name", "unknown_op", "test"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "unsupported operation")
	})
}

func TestNewFilterCondition_ValidEqual(t *testing.T) {
	Convey("Test NewFilterCondition creates equal condition", t, func() {
		cond, err := newTestCondition(constCfg("name", "==", "alice"))
		So(err, ShouldBeNil)
		So(cond, ShouldNotBeNil)
		So(cond.GetOperation(), ShouldEqual, OperationEqual)
	})
}

func TestIsSlice(t *testing.T) {
	Convey("Test IsSlice", t, func() {
		So(IsSlice([]int{1, 2, 3}), ShouldBeTrue)
		So(IsSlice([]string{"a"}), ShouldBeTrue)
		So(IsSlice("not a slice"), ShouldBeFalse)
		So(IsSlice(42), ShouldBeFalse)
	})
}

func TestIsSameType(t *testing.T) {
	Convey("Test IsSameType", t, func() {
		So(IsSameType([]any{}), ShouldBeTrue)
		So(IsSameType([]any{1, 2, 3}), ShouldBeTrue)
		So(IsSameType([]any{"a", "b"}), ShouldBeTrue)
		So(IsSameType([]any{1, "two", 3}), ShouldBeFalse)
	})
}

func TestEqualCond_Valid(t *testing.T) {
	Convey("Test EqualCond valid const value", t, func() {
		cond, err := newTestCondition(constCfg("name", "==", "alice"))
		So(err, ShouldBeNil)
		eq := cond.(*EqualCond)
		So(eq.Lfield.Name, ShouldEqual, "name")
		So(eq.Value, ShouldEqual, "alice")
	})
}

func TestEqualCond_EmptyFieldName(t *testing.T) {
	Convey("Test EqualCond rejects empty field", t, func() {
		_, err := newTestCondition(constCfg("", "==", "alice"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "left field is empty")
	})
}

func TestEqualCond_FieldNotFound(t *testing.T) {
	Convey("Test EqualCond rejects unknown field", t, func() {
		_, err := newTestCondition(constCfg("nonexistent", "==", "alice"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "not found")
	})
}

func TestEqualCond_RejectsArrayValue(t *testing.T) {
	Convey("Test EqualCond rejects array value", t, func() {
		_, err := newTestCondition(constCfg("name", "==", []any{"a", "b"}))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "single value")
	})
}

func TestEqualCond_FieldToField(t *testing.T) {
	Convey("Test EqualCond field to field", t, func() {
		cfg := &interfaces.FilterCondCfg{
			Name:      "name",
			Operation: "==",
			ValueOptCfg: interfaces.ValueOptCfg{
				ValueFrom: interfaces.ValueFrom_Field,
				Value:     "other_name",
			},
		}
		cond, err := newTestCondition(cfg)
		So(err, ShouldBeNil)
		eq := cond.(*EqualCond)
		So(eq.Rfield, ShouldNotBeNil)
		So(eq.Rfield.Name, ShouldEqual, "other_name")
	})
}

func TestEqualCond_FieldToField_RightNotFound(t *testing.T) {
	Convey("Test EqualCond rejects unknown right field", t, func() {
		cfg := &interfaces.FilterCondCfg{
			Name:      "name",
			Operation: "==",
			ValueOptCfg: interfaces.ValueOptCfg{
				ValueFrom: interfaces.ValueFrom_Field,
				Value:     "nonexistent",
			},
		}
		_, err := newTestCondition(cfg)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "right field")
	})
}

func TestEqualCond_AliasOperations(t *testing.T) {
	Convey("Test EqualCond alias operation", t, func() {
		cond, err := newTestCondition(constCfg("name", "eq", "alice"))
		So(err, ShouldBeNil)
		So(cond.GetOperation(), ShouldEqual, OperationEqual)
	})
}

func TestInCond_Valid(t *testing.T) {
	Convey("Test InCond valid", t, func() {
		cond, err := newTestCondition(constCfg("name", "in", []any{"alice", "bob"}))
		So(err, ShouldBeNil)
		in := cond.(*InCond)
		So(in.Value, ShouldHaveLength, 2)
	})
}

func TestInCond_EmptyArray(t *testing.T) {
	Convey("Test InCond rejects empty array", t, func() {
		_, err := newTestCondition(constCfg("name", "in", []any{}))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "length >= 1")
	})
}

func TestInCond_NonArrayValue(t *testing.T) {
	Convey("Test InCond rejects non-array value", t, func() {
		_, err := newTestCondition(constCfg("name", "in", "single_value"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "should be an array")
	})
}

func TestInCond_RejectsFieldValueFrom(t *testing.T) {
	Convey("Test InCond rejects field value_from", t, func() {
		cfg := &interfaces.FilterCondCfg{
			Name:      "name",
			Operation: "in",
			ValueOptCfg: interfaces.ValueOptCfg{
				ValueFrom: interfaces.ValueFrom_Field,
				Value:     []any{"alice"},
			},
		}
		_, err := newTestCondition(cfg)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "does not support value_from")
	})
}

func TestLikeCond_Valid(t *testing.T) {
	Convey("Test LikeCond valid", t, func() {
		cond, err := newTestCondition(constCfg("name", "like", "ali%"))
		So(err, ShouldBeNil)
		like := cond.(*LikeCond)
		So(like.Value, ShouldEqual, "ali%")
	})
}

func TestLikeCond_NonStringField(t *testing.T) {
	Convey("Test LikeCond rejects non-string field", t, func() {
		_, err := newTestCondition(constCfg("age", "like", "test"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "not a string field")
	})
}

func TestLikeCond_NonStringValue(t *testing.T) {
	Convey("Test LikeCond rejects non-string value", t, func() {
		_, err := newTestCondition(constCfg("name", "like", 123))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "not a string value")
	})
}

func TestRangeCond_Valid(t *testing.T) {
	Convey("Test RangeCond valid", t, func() {
		cond, err := newTestCondition(constCfg("age", "range", []any{18, 65}))
		So(err, ShouldBeNil)
		r := cond.(*RangeCond)
		So(r.Value, ShouldHaveLength, 2)
	})
}

func TestRangeCond_WrongArrayLength(t *testing.T) {
	Convey("Test RangeCond rejects wrong array length", t, func() {
		_, err := newTestCondition(constCfg("age", "range", []any{18}))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "length 2")
	})
}

func TestRangeCond_NonNumericField(t *testing.T) {
	Convey("Test RangeCond rejects non-numeric field", t, func() {
		_, err := newTestCondition(constCfg("name", "range", []any{1, 2}))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "not a date/number field")
	})
}

func TestRangeCond_DateField(t *testing.T) {
	Convey("Test RangeCond accepts date field", t, func() {
		_, err := newTestCondition(constCfg("created_at", "range", []any{"2024-01-01", "2024-12-31"}))
		So(err, ShouldBeNil)
	})
}

func TestAndCond_Valid(t *testing.T) {
	Convey("Test AndCond valid", t, func() {
		cfg := &interfaces.FilterCondCfg{
			Operation: "and",
			SubConds: []*interfaces.FilterCondCfg{
				constCfg("name", "==", "alice"),
				constCfg("age", ">", 18),
			},
		}
		cond, err := newTestCondition(cfg)
		So(err, ShouldBeNil)
		and := cond.(*AndCond)
		So(and.SubConds, ShouldHaveLength, 2)
	})
}

func TestAndCond_EmptySubConds(t *testing.T) {
	Convey("Test AndCond rejects empty sub conditions", t, func() {
		_, err := newTestCondition(&interfaces.FilterCondCfg{Operation: "and", SubConds: []*interfaces.FilterCondCfg{}})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "size is 0")
	})
}

func TestAndCond_InvalidSubCond(t *testing.T) {
	Convey("Test AndCond rejects invalid sub condition", t, func() {
		cfg := &interfaces.FilterCondCfg{
			Operation: "and",
			SubConds: []*interfaces.FilterCondCfg{
				constCfg("nonexistent", "==", "test"),
			},
		}
		_, err := newTestCondition(cfg)
		So(err, ShouldNotBeNil)
	})
}

func TestAndCond_NestedAndOr(t *testing.T) {
	Convey("Test AndCond supports nested or", t, func() {
		cfg := &interfaces.FilterCondCfg{
			Operation: "and",
			SubConds: []*interfaces.FilterCondCfg{
				constCfg("name", "==", "alice"),
				{
					Operation: "or",
					SubConds: []*interfaces.FilterCondCfg{
						constCfg("age", ">", 18),
						constCfg("age", "<", 65),
					},
				},
			},
		}
		cond, err := newTestCondition(cfg)
		So(err, ShouldBeNil)
		and := cond.(*AndCond)
		So(and.SubConds, ShouldHaveLength, 2)
		So(and.SubConds[1].GetOperation(), ShouldEqual, OperationOr)
	})
}

func TestAllOperationsRegistered(t *testing.T) {
	Convey("Test all operations are registered", t, func() {
		assertOpRegistered("and")
		assertOpRegistered("or")
		assertOpRegistered("==")
		assertOpRegistered("eq")
		assertOpRegistered("!=")
		assertOpRegistered("not_eq")
		assertOpRegistered(">")
		assertOpRegistered("gt")
		assertOpRegistered(">=")
		assertOpRegistered("gte")
		assertOpRegistered("<")
		assertOpRegistered("lt")
		assertOpRegistered("<=")
		assertOpRegistered("lte")
		assertOpRegistered("in")
		assertOpRegistered("not_in")
		assertOpRegistered("like")
		assertOpRegistered("not_like")
		assertOpRegistered("contain")
		assertOpRegistered("not_contain")
		assertOpRegistered("range")
		assertOpRegistered("out_range")
		assertOpRegistered("exist")
		assertOpRegistered("not_exist")
		assertOpRegistered("empty")
		assertOpRegistered("not_empty")
		assertOpRegistered("regex")
		assertOpRegistered("match")
		assertOpRegistered("match_phrase")
		assertOpRegistered("prefix")
		assertOpRegistered("not_prefix")
		assertOpRegistered("null")
		assertOpRegistered("not_null")
		assertOpRegistered("true")
		assertOpRegistered("false")
		assertOpRegistered("before")
		assertOpRegistered("current")
		assertOpRegistered("between")
		assertOpRegistered("knn_vector")
		assertOpRegistered("multi_match")
	})
}

func assertOpRegistered(op string) {
	_, exists := OperationMap[op]
	So(exists, ShouldBeTrue)
}

func TestComparisonOps_EmptyField(t *testing.T) {
	Convey("Test comparison ops reject empty field", t, func() {
		assertComparisonEmptyField("==")
		assertComparisonEmptyField("!=")
		assertComparisonEmptyField(">")
		assertComparisonEmptyField(">=")
		assertComparisonEmptyField("<")
		assertComparisonEmptyField("<=")
	})
}

func assertComparisonEmptyField(op string) {
	_, err := newTestCondition(constCfg("", op, "test"))
	So(err, ShouldNotBeNil)
}

func TestComparisonOps_FieldNotFound(t *testing.T) {
	Convey("Test comparison ops reject unknown field", t, func() {
		_, err := newTestCondition(constCfg("nonexistent", "==", "test"))
		So(err, ShouldNotBeNil)
	})
}

func TestNullCond_Valid(t *testing.T) {
	Convey("Test NullCond valid", t, func() {
		cond, err := newTestCondition(&interfaces.FilterCondCfg{Name: "name", Operation: "null"})
		So(err, ShouldBeNil)
		So(cond.GetOperation(), ShouldEqual, OperationNull)
		So(cond.NeedValue(), ShouldBeFalse)
	})
}
