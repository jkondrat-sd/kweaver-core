// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package remote

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestRemoteConnector_MetadataAccessors(t *testing.T) {
	Convey("Test RemoteConnector metadata accessors", t, func() {
		fieldConfig := map[string]interfaces.ConnectorFieldConfig{
			"endpoint": {Name: "Endpoint", Type: "string", Required: true},
		}
		connector := NewRemoteConnector(&interfaces.ConnectorType{
			Type:        "custom",
			Name:        "Custom",
			Mode:        interfaces.ConnectorModeRemote,
			Category:    interfaces.ConnectorCategoryAPI,
			Enabled:     true,
			FieldConfig: fieldConfig,
		})

		So(connector.GetType(), ShouldEqual, "custom")
		So(connector.GetName(), ShouldEqual, "Custom")
		So(connector.GetMode(), ShouldEqual, interfaces.ConnectorModeRemote)
		So(connector.GetCategory(), ShouldEqual, interfaces.ConnectorCategoryAPI)
		So(connector.GetEnabled(), ShouldBeTrue)
		So(connector.GetFieldConfig(), ShouldResemble, fieldConfig)
		So(connector.GetSensitiveFields(), ShouldResemble, []string{"password"})
	})
}

func TestRemoteConnector_SetEnabled(t *testing.T) {
	Convey("Test RemoteConnector SetEnabled", t, func() {
		connector := NewRemoteConnector(&interfaces.ConnectorType{Enabled: true})
		connector.SetEnabled(false)
		So(connector.GetEnabled(), ShouldBeFalse)
	})
}

func TestRemoteConnector_NewCopiesDefinitionAndConfig(t *testing.T) {
	Convey("Test RemoteConnector New creates configured instance", t, func() {
		connector := NewRemoteConnector(&interfaces.ConnectorType{
			Type:    "custom",
			Name:    "Custom",
			Mode:    interfaces.ConnectorModeRemote,
			Enabled: true,
		})

		instance, err := connector.New(interfaces.ConnectorConfig{"token": "secret"})
		So(err, ShouldBeNil)
		remoteInstance, ok := instance.(*RemoteConnector)
		So(ok, ShouldBeTrue)
		So(remoteInstance.GetType(), ShouldEqual, "custom")
		So(remoteInstance.config["token"], ShouldEqual, "secret")
	})
}

func TestRemoteConnector_NoOpMethods(t *testing.T) {
	Convey("Test RemoteConnector no-op methods", t, func() {
		connector := NewRemoteConnector(&interfaces.ConnectorType{})
		So(connector.Connect(context.Background()), ShouldBeNil)
		So(connector.Ping(context.Background()), ShouldBeNil)
		So(connector.TestConnection(context.Background()), ShouldBeNil)
		So(connector.Close(context.Background()), ShouldBeNil)
		metadata, err := connector.GetMetadata(context.Background())
		So(err, ShouldBeNil)
		So(metadata, ShouldBeNil)
	})
}
