// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package anyshare

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestAnyShareConnector_MetadataAccessors(t *testing.T) {
	Convey("Test AnyShareConnector metadata accessors", t, func() {
		connector := &AnyShareConnector{}
		So(connector.GetType(), ShouldEqual, interfaces.ConnectorTypeAnyShare)
		So(connector.GetName(), ShouldEqual, interfaces.ConnectorTypeAnyShare)
		So(connector.GetMode(), ShouldEqual, interfaces.ConnectorModeLocal)
		So(connector.GetCategory(), ShouldEqual, interfaces.ConnectorCategoryFileset)
		So(connector.GetSensitiveFields(), ShouldResemble, []string{"token", "app_secret"})
		So(connector.GetFieldConfig()["token"].Encrypted, ShouldBeTrue)
		So(connector.GetFieldConfig()["app_secret"].Encrypted, ShouldBeTrue)
	})
}

func TestAnyShareConnector_SetEnabled(t *testing.T) {
	Convey("Test AnyShareConnector SetEnabled", t, func() {
		connector := &AnyShareConnector{}
		connector.SetEnabled(true)
		So(connector.GetEnabled(), ShouldBeTrue)
	})
}

func TestAnyShareConnector_New_TokenAuthSuccess(t *testing.T) {
	Convey("Test AnyShareConnector New succeeds with token auth", t, func() {
		connector := &AnyShareConnector{}
		instance, err := connector.New(interfaces.ConnectorConfig{
			"protocol":     "https",
			"host":         "anyshare.example.com",
			"port":         443,
			"auth_type":    authTypeToken,
			"token":        "token-1",
			"doc_lib_type": docLibTypeKnowledge,
		})
		So(err, ShouldBeNil)
		So(instance, ShouldNotBeNil)
		anyshareInstance := instance.(*AnyShareConnector)
		So(anyshareInstance.baseURL, ShouldEqual, "https://anyshare.example.com:443")
	})
}

func TestAnyShareConnector_New_AppSecretAuthSuccess(t *testing.T) {
	Convey("Test AnyShareConnector New succeeds with app secret auth", t, func() {
		connector := &AnyShareConnector{}
		instance, err := connector.New(interfaces.ConnectorConfig{
			"protocol":     "http",
			"host":         "anyshare.example.com",
			"port":         8080,
			"auth_type":    authTypeAppSecret,
			"app_id":       "app-1",
			"app_secret":   "secret",
			"doc_lib_type": docLibTypeDocument,
		})
		So(err, ShouldBeNil)
		So(instance, ShouldNotBeNil)
	})
}

func TestAnyShareConnector_New_RejectsInvalidProtocol(t *testing.T) {
	Convey("Test AnyShareConnector New rejects invalid protocol", t, func() {
		connector := &AnyShareConnector{}
		instance, err := connector.New(interfaces.ConnectorConfig{"protocol": "ftp"})
		So(instance, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "protocol must be http or https")
	})
}

func TestAnyShareConnector_New_RejectsMissingHostAndPort(t *testing.T) {
	Convey("Test AnyShareConnector New rejects missing host and port", t, func() {
		connector := &AnyShareConnector{}
		instance, err := connector.New(interfaces.ConnectorConfig{
			"protocol": "https",
		})
		So(instance, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "host and port are required")
	})
}

func TestAnyShareConnector_New_RejectsInvalidAuthType(t *testing.T) {
	Convey("Test AnyShareConnector New rejects invalid auth type", t, func() {
		connector := &AnyShareConnector{}
		instance, err := connector.New(interfaces.ConnectorConfig{
			"protocol":  "https",
			"host":      "anyshare.example.com",
			"port":      443,
			"auth_type": 99,
		})
		So(instance, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "auth_type must be 1")
	})
}

func TestAnyShareConnector_New_RejectsDuplicatePaths(t *testing.T) {
	Convey("Test AnyShareConnector New rejects duplicate paths", t, func() {
		connector := &AnyShareConnector{}
		instance, err := connector.New(interfaces.ConnectorConfig{
			"protocol":     "https",
			"host":         "anyshare.example.com",
			"port":         443,
			"auth_type":    authTypeToken,
			"token":        "token-1",
			"doc_lib_type": docLibTypeKnowledge,
			"paths":        []string{"/team", "/team"},
		})
		So(instance, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "duplicate element found")
	})
}

func TestNormalizeBearer_AddsPrefix(t *testing.T) {
	Convey("Test normalizeBearer adds bearer prefix", t, func() {
		So(normalizeBearer("token-1"), ShouldEqual, "Bearer token-1")
	})
}

func TestNormalizeBearer_KeepsExistingPrefix(t *testing.T) {
	Convey("Test normalizeBearer keeps existing bearer prefix", t, func() {
		So(normalizeBearer("Bearer token-1"), ShouldEqual, "Bearer token-1")
	})
}

func TestTruncateForLog_ShortBody(t *testing.T) {
	Convey("Test truncateForLog returns short body unchanged", t, func() {
		So(truncateForLog([]byte("short")), ShouldEqual, "short")
	})
}

func TestTruncateForLog_LongBody(t *testing.T) {
	Convey("Test truncateForLog truncates long body", t, func() {
		result := truncateForLog([]byte(strings.Repeat("a", 513)))
		So(result, ShouldHaveLength, 515)
		So(result, ShouldEndWith, "...")
	})
}

func TestAnyShareConnector_CloseClearsRuntimeState(t *testing.T) {
	Convey("Test AnyShareConnector Close clears runtime state", t, func() {
		connector := &AnyShareConnector{connected: true, authHeader: "Bearer token"}
		err := connector.Close(context.Background())
		So(err, ShouldBeNil)
		So(connector.connected, ShouldBeFalse)
		So(connector.authHeader, ShouldBeBlank)
	})
}
