// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package factory

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
)

func TestConnectorFactory_RegisterConnector(t *testing.T) {
	Convey("Test ConnectorFactory.RegisterConnector", t, func() {
		cf := &ConnectorFactory{connectors: map[string]connectors.Connector{}}

		Convey("registers remote connector", func() {
			err := cf.RegisterConnector(context.Background(), "custom", &interfaces.ConnectorType{
				Type:    "custom",
				Name:    "Custom",
				Mode:    interfaces.ConnectorModeRemote,
				Enabled: true,
			})
			So(err, ShouldBeNil)
			So(cf.connectors["custom"].GetType(), ShouldEqual, "custom")
			So(cf.connectors["custom"].GetEnabled(), ShouldBeTrue)
		})

		Convey("rejects unregistered local connector", func() {
			err := cf.RegisterConnector(context.Background(), "custom", &interfaces.ConnectorType{
				Type: "custom",
				Name: "Custom",
				Mode: interfaces.ConnectorModeLocal,
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "local connector custom:Custom not implemented")
		})
	})
}

func TestConnectorFactory_SetConnectorEnabled(t *testing.T) {
	Convey("Test ConnectorFactory.SetConnectorEnabled", t, func() {
		Convey("toggles enabled flag on registered connector", func() {
			connector := &fakeConnector{enabled: false}
			cf := &ConnectorFactory{connectors: map[string]connectors.Connector{"custom": connector}}
			err := cf.SetConnectorEnabled(context.Background(), "custom", true)
			So(err, ShouldBeNil)
			So(connector.GetEnabled(), ShouldBeTrue)
		})

		Convey("returns error when connector is missing", func() {
			cf := &ConnectorFactory{connectors: map[string]connectors.Connector{}}
			err := cf.SetConnectorEnabled(context.Background(), "missing", true)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "connector missing not implemented")
		})
	})
}

func TestConnectorFactory_CreateConnectorInstance(t *testing.T) {
	Convey("Test ConnectorFactory.CreateConnectorInstance", t, func() {
		Convey("returns instance when connector is enabled", func() {
			cf := &ConnectorFactory{connectors: map[string]connectors.Connector{
				"custom": &fakeConnector{enabled: true},
			}}
			instance, err := cf.CreateConnectorInstance(context.Background(), "custom", interfaces.ConnectorConfig{"token": "secret"})
			So(err, ShouldBeNil)
			So(instance.GetType(), ShouldEqual, "custom")
		})

		Convey("rejects disabled connector", func() {
			cf := &ConnectorFactory{connectors: map[string]connectors.Connector{
				"custom": &fakeConnector{enabled: false},
			}}
			instance, err := cf.CreateConnectorInstance(context.Background(), "custom", nil)
			So(instance, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "connector custom is disabled")
		})
	})
}

func TestConnectorFactory_GetSensitiveFields(t *testing.T) {
	Convey("Test ConnectorFactory.GetSensitiveFields", t, func() {
		cf := &ConnectorFactory{connectors: map[string]connectors.Connector{
			"custom": &fakeConnector{sensitiveFields: []string{"password", "token"}},
		}}

		Convey("returns declared sensitive fields", func() {
			So(cf.GetSensitiveFields("custom"), ShouldResemble, []string{"password", "token"})
		})

		Convey("returns nil for unknown connector", func() {
			So(cf.GetSensitiveFields("missing"), ShouldBeNil)
		})
	})
}

type fakeConnector struct {
	enabled         bool
	sensitiveFields []string
}

func (f *fakeConnector) GetType() string {
	return "custom"
}

func (f *fakeConnector) GetName() string {
	return "Custom"
}

func (f *fakeConnector) GetMode() string {
	return interfaces.ConnectorModeRemote
}

func (f *fakeConnector) GetCategory() string {
	return interfaces.ConnectorCategoryAPI
}

func (f *fakeConnector) GetEnabled() bool {
	return f.enabled
}

func (f *fakeConnector) SetEnabled(enabled bool) {
	f.enabled = enabled
}

func (f *fakeConnector) GetSensitiveFields() []string {
	return f.sensitiveFields
}

func (f *fakeConnector) GetFieldConfig() map[string]interfaces.ConnectorFieldConfig {
	return nil
}

func (f *fakeConnector) New(cfg interfaces.ConnectorConfig) (connectors.Connector, error) {
	return &fakeConnector{enabled: f.enabled, sensitiveFields: f.sensitiveFields}, nil
}

func (f *fakeConnector) Connect(ctx context.Context) error {
	return nil
}

func (f *fakeConnector) Ping(ctx context.Context) error {
	return nil
}

func (f *fakeConnector) Close(ctx context.Context) error {
	return nil
}

func (f *fakeConnector) TestConnection(ctx context.Context) error {
	return nil
}

func (f *fakeConnector) GetMetadata(ctx context.Context) (map[string]any, error) {
	return nil, nil
}
