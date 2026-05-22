// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package sql

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestNewlogicDefinitionSQLGenerator_BuildsNodeAndFieldMaps(t *testing.T) {
	Convey("Test NewlogicDefinitionSQLGenerator builds node and field maps", t, func() {
		view := &interfaces.LogicView{
			Resource: interfaces.Resource{
				LogicDefinition: []*interfaces.LogicDefinitionNode{
					{ID: "resource-node", Type: interfaces.LogicDefinitionNodeType_Resource},
					{ID: "output-node", Type: interfaces.LogicDefinitionNodeType_Output},
				},
				SchemaDefinition: []*interfaces.Property{
					{Name: "title", Type: interfaces.DataType_String},
				},
			},
			RefResources: map[string]*interfaces.Resource{
				"resource-1": {ID: "resource-1"},
			},
		}

		generator := NewlogicDefinitionSQLGenerator(view)
		So(generator.nodes["resource-node"].ID, ShouldEqual, "resource-node")
		So(generator.outputNode.ID, ShouldEqual, "output-node")
		So(generator.viewFieldMap["title"].Type, ShouldEqual, interfaces.DataType_String)
		So(generator.RefResources["resource-1"].ID, ShouldEqual, "resource-1")
	})
}

func TestLogicViewSQLGenerator_BuildLogicDefinitionSQL_RejectsEmptyDefinition(t *testing.T) {
	Convey("Test logicViewSQLGenerator BuildLogicDefinitionSQL rejects empty definition", t, func() {
		view := &interfaces.LogicView{
			Resource: interfaces.Resource{Name: "view-1"},
		}
		generator := NewlogicDefinitionSQLGenerator(view)
		query, err := generator.BuildLogicDefinitionSQL(context.Background(), view)
		So(query, ShouldBeBlank)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "logic definition is empty")
	})
}

func TestLogicViewSQLGenerator_BuildLogicDefinitionSQL_RejectsMissingOutputNode(t *testing.T) {
	Convey("Test logicViewSQLGenerator BuildLogicDefinitionSQL rejects missing output node", t, func() {
		view := &interfaces.LogicView{
			Resource: interfaces.Resource{
				Name: "view-1",
				LogicDefinition: []*interfaces.LogicDefinitionNode{
					{ID: "resource-node", Type: interfaces.LogicDefinitionNodeType_Resource},
				},
			},
		}
		generator := NewlogicDefinitionSQLGenerator(view)
		query, err := generator.BuildLogicDefinitionSQL(context.Background(), view)
		So(query, ShouldBeBlank)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "output node not found")
	})
}
