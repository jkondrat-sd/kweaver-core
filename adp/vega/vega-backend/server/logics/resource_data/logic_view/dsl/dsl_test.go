// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package dsl

import (
	"context"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestNewlogicViewDSLGenerator_BuildsNodeAndFieldMaps(t *testing.T) {
	Convey("Test NewlogicViewDSLGenerator builds node and field maps", t, func() {
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
		}

		generator := NewlogicViewDSLGenerator(view)
		So(generator.nodes["resource-node"].ID, ShouldEqual, "resource-node")
		So(generator.outputNode.ID, ShouldEqual, "output-node")
		So(generator.viewFieldMap["title"].Type, ShouldEqual, interfaces.DataType_String)
	})
}

func TestLogicViewDSLGenerator_BuildDSL_RejectsEmptySortField(t *testing.T) {
	Convey("Test logicViewDSLGenerator BuildDSL rejects empty sort field", t, func() {
		view := &interfaces.LogicView{
			Resource: interfaces.Resource{
				LogicDefinition: []*interfaces.LogicDefinitionNode{
					{ID: "output-node", Type: interfaces.LogicDefinitionNodeType_Output},
				},
			},
		}
		generator := NewlogicViewDSLGenerator(view)
		dsl, err := generator.BuildDSL(context.Background(), interfaces.ResourceDataQueryParams{
			Sort: []*interfaces.SortField{{Field: "", Direction: "asc"}},
		}, view, nil)
		So(dsl.Sort, ShouldBeNil)
		So(err, ShouldNotBeNil)
		httpErr, ok := err.(*rest.HTTPError)
		So(ok, ShouldBeTrue)
		So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
	})
}
