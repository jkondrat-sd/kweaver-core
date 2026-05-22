// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package oracle

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestOracleConnector_MetadataAccessors(t *testing.T) {
	Convey("Test OracleConnector metadata accessors", t, func() {
		connector := &OracleConnector{}
		So(connector.GetType(), ShouldEqual, interfaces.ConnectorTypeOracle)
		So(connector.GetName(), ShouldEqual, interfaces.ConnectorTypeOracle)
		So(connector.GetMode(), ShouldEqual, interfaces.ConnectorModeLocal)
		So(connector.GetCategory(), ShouldEqual, interfaces.ConnectorCategoryTable)
		So(connector.GetSensitiveFields(), ShouldResemble, []string{"password"})
		So(connector.GetFieldConfig()["password"].Encrypted, ShouldBeTrue)
	})
}

func TestOracleConnector_SetEnabled(t *testing.T) {
	Convey("Test OracleConnector SetEnabled", t, func() {
		connector := &OracleConnector{}
		connector.SetEnabled(true)
		So(connector.GetEnabled(), ShouldBeTrue)
	})
}

func TestOracleConnector_MapType_Integer(t *testing.T) {
	Convey("Test OracleConnector MapType maps integer", t, func() {
		connector := &OracleConnector{}
		So(connector.MapType("integer"), ShouldEqual, "integer")
	})
}

func TestOracleConnector_MapType_Timestamp(t *testing.T) {
	Convey("Test OracleConnector MapType maps timestamp", t, func() {
		connector := &OracleConnector{}
		So(connector.MapType("timestamp with time zone"), ShouldEqual, "datetime")
	})
}

func TestOracleConnector_MapType_Unsupported(t *testing.T) {
	Convey("Test OracleConnector MapType returns unsupported for unknown type", t, func() {
		connector := &OracleConnector{}
		So(connector.MapType("custom_type"), ShouldEqual, "unsupported")
	})
}

func TestOracleConnector_NewRejectsIncompleteConfig(t *testing.T) {
	Convey("Test OracleConnector New rejects incomplete config", t, func() {
		connector := &OracleConnector{}
		instance, err := connector.New(interfaces.ConnectorConfig{"host": "localhost"})
		So(instance, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "oracle connector config is incomplete")
	})
}

func TestOracleConnector_NewRejectsInvalidPort(t *testing.T) {
	Convey("Test OracleConnector New rejects invalid port", t, func() {
		connector := &OracleConnector{}
		instance, err := connector.New(interfaces.ConnectorConfig{
			"host":         "localhost",
			"port":         70000,
			"service_name": "ORCL",
			"username":     "user",
			"password":     "pass",
		})
		So(instance, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "out of valid range")
	})
}
