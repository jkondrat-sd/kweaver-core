// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package extensions

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

func TestValidateEntityExtensionsMap_Success(t *testing.T) {
	Convey("Test ValidateEntityExtensionsMap success", t, func() {
		err := ValidateEntityExtensionsMap(context.Background(), map[string]string{
			"owner": "data-team",
			"env":   "prod",
		})
		So(err, ShouldBeNil)
	})
}

func TestValidateEntityExtensionsMap_QuotaExceeded(t *testing.T) {
	Convey("Test ValidateEntityExtensionsMap quota exceeded", t, func() {
		err := ValidateEntityExtensionsMap(context.Background(), mapWithPairs(MaxEntityExtensionPairs+1))
		assertExtensionsHTTPError(err, verrors.VegaBackend_Extensions_QuotaExceeded)
	})
}

func TestValidateEntityExtensionsMap_EmptyKey(t *testing.T) {
	Convey("Test ValidateEntityExtensionsMap rejects empty key", t, func() {
		err := ValidateEntityExtensionsMap(context.Background(), map[string]string{"": "value"})
		assertExtensionsHTTPError(err, verrors.VegaBackend_Extensions_InvalidFormat)
	})
}

func TestValidateEntityExtensionsMap_KeyTooLong(t *testing.T) {
	Convey("Test ValidateEntityExtensionsMap rejects long key", t, func() {
		err := ValidateEntityExtensionsMap(context.Background(), map[string]string{
			strings.Repeat("a", MaxExtensionKeyLen+1): "value",
		})
		assertExtensionsHTTPError(err, verrors.VegaBackend_Extensions_InvalidFormat)
	})
}

func TestValidateEntityExtensionsMap_ReservedKey(t *testing.T) {
	Convey("Test ValidateEntityExtensionsMap rejects reserved key", t, func() {
		err := ValidateEntityExtensionsMap(context.Background(), map[string]string{"Vega_Internal": "value"})
		assertExtensionsHTTPError(err, verrors.VegaBackend_Extensions_ReservedKey)
	})
}

func TestValidateEntityExtensionsMap_ValueTooLong(t *testing.T) {
	Convey("Test ValidateEntityExtensionsMap rejects long value", t, func() {
		err := ValidateEntityExtensionsMap(context.Background(), map[string]string{
			"owner": strings.Repeat("a", MaxExtensionValueLen+1),
		})
		assertExtensionsHTTPError(err, verrors.VegaBackend_Extensions_InvalidFormat)
	})
}

func TestValidatePropertyExtensionsMap_Success(t *testing.T) {
	Convey("Test ValidatePropertyExtensionsMap success", t, func() {
		err := ValidatePropertyExtensionsMap(context.Background(), map[string]string{"pii": "false"})
		So(err, ShouldBeNil)
	})
}

func TestValidatePropertyExtensionsMap_QuotaExceeded(t *testing.T) {
	Convey("Test ValidatePropertyExtensionsMap quota exceeded", t, func() {
		err := ValidatePropertyExtensionsMap(context.Background(), mapWithPairs(MaxPropertyExtensionPairs+1))
		assertExtensionsHTTPError(err, verrors.VegaBackend_Extensions_PropertyQuotaExceeded)
	})
}

func TestValidateSchemaPropertiesExtensions_SkipNilAndEmpty(t *testing.T) {
	Convey("Test ValidateSchemaPropertiesExtensions skips nil and empty properties", t, func() {
		err := ValidateSchemaPropertiesExtensions(context.Background(), []*interfaces.Property{
			nil,
			{Extensions: nil},
			{Extensions: map[string]string{}},
		})
		So(err, ShouldBeNil)
	})
}

func TestValidateSchemaPropertiesExtensions_InvalidPropertyExtension(t *testing.T) {
	Convey("Test ValidateSchemaPropertiesExtensions returns property validation error", t, func() {
		err := ValidateSchemaPropertiesExtensions(context.Background(), []*interfaces.Property{
			{Extensions: map[string]string{"vega_reserved": "value"}},
		})
		assertExtensionsHTTPError(err, verrors.VegaBackend_Extensions_ReservedKey)
	})
}

func TestValidateExtensionQueryPairs_Empty(t *testing.T) {
	Convey("Test ValidateExtensionQueryPairs allows empty filters", t, func() {
		err := ValidateExtensionQueryPairs(context.Background(), nil, nil)
		So(err, ShouldBeNil)
	})
}

func TestValidateExtensionQueryPairs_Success(t *testing.T) {
	Convey("Test ValidateExtensionQueryPairs success", t, func() {
		err := ValidateExtensionQueryPairs(context.Background(), []string{"owner"}, []string{"data-team"})
		So(err, ShouldBeNil)
	})
}

func TestValidateExtensionQueryPairs_MismatchedLength(t *testing.T) {
	Convey("Test ValidateExtensionQueryPairs rejects mismatched length", t, func() {
		err := ValidateExtensionQueryPairs(context.Background(), []string{"owner"}, nil)
		assertExtensionsHTTPError(err, verrors.VegaBackend_Extensions_MismatchedQueryPairs)
	})
}

func TestValidateExtensionQueryPairs_TooManyPairs(t *testing.T) {
	Convey("Test ValidateExtensionQueryPairs rejects too many pairs", t, func() {
		err := ValidateExtensionQueryPairs(context.Background(),
			[]string{"k1", "k2", "k3", "k4", "k5", "k6"},
			[]string{"v1", "v2", "v3", "v4", "v5", "v6"},
		)
		assertExtensionsHTTPError(err, verrors.VegaBackend_Extensions_TooManyFilterPairs)
	})
}

func TestValidateExtensionQueryPairs_EmptyKey(t *testing.T) {
	Convey("Test ValidateExtensionQueryPairs rejects empty key", t, func() {
		err := ValidateExtensionQueryPairs(context.Background(), []string{""}, []string{"value"})
		assertExtensionsHTTPError(err, verrors.VegaBackend_Extensions_MismatchedQueryPairs)
	})
}

func TestValidateExtensionQueryPairs_EmptyValue(t *testing.T) {
	Convey("Test ValidateExtensionQueryPairs rejects empty value", t, func() {
		err := ValidateExtensionQueryPairs(context.Background(), []string{"owner"}, []string{""})
		assertExtensionsHTTPError(err, verrors.VegaBackend_Extensions_MismatchedQueryPairs)
	})
}

func assertExtensionsHTTPError(err error, errorCode string) {
	So(err, ShouldNotBeNil)
	httpErr, ok := err.(*rest.HTTPError)
	So(ok, ShouldBeTrue)
	So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
	So(httpErr.BaseError.ErrorCode, ShouldEqual, errorCode)
}

func mapWithPairs(count int) map[string]string {
	extensions := make(map[string]string, count)
	for i := 0; i < count; i++ {
		extensions["key_"+string(rune('a'+i))] = "value"
	}
	return extensions
}
