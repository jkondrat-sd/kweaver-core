// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

var testSortTypes = map[string]string{
	"name":        "f_name",
	"create_time": "f_create_time",
}

func TestValidate_Name(t *testing.T) {
	Convey("Test validateName", t, func() {
		Convey("Valid name", func() {
			err := validateName(context.Background(), "test-catalog")
			So(err, ShouldBeNil)
		})

		Convey("Empty name", func() {
			err := validateName(context.Background(), "")
			So(err, ShouldNotBeNil)
		})

		Convey("Max length name", func() {
			name := strings.Repeat("a", interfaces.NAME_MAX_LENGTH)
			err := validateName(context.Background(), name)
			So(err, ShouldBeNil)
		})

		Convey("Exceeds max length", func() {
			name := strings.Repeat("a", interfaces.NAME_MAX_LENGTH+1)
			err := validateName(context.Background(), name)
			So(err, ShouldNotBeNil)
		})

		Convey("UTF-8 name length", func() {
			name := strings.Repeat("中", interfaces.NAME_MAX_LENGTH)
			err := validateName(context.Background(), name)
			So(err, ShouldBeNil)

			name = strings.Repeat("中", interfaces.NAME_MAX_LENGTH+1)
			err = validateName(context.Background(), name)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestValidate_ID(t *testing.T) {
	Convey("Test validateID", t, func() {
		Convey("Empty ID", func() {
			err := validateID(context.Background(), "")
			So(err, ShouldBeNil)
		})

		Convey("Valid ID", func() {
			err := validateID(context.Background(), "test-id_123")
			So(err, ShouldBeNil)
		})

		Convey("Max length ID", func() {
			id := strings.Repeat("a", 40)
			err := validateID(context.Background(), id)
			So(err, ShouldBeNil)
		})

		Convey("Exceeds max length", func() {
			id := strings.Repeat("a", 41)
			err := validateID(context.Background(), id)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid character", func() {
			err := validateID(context.Background(), "test.id")
			So(err, ShouldNotBeNil)
		})

		Convey("Starts with underscore", func() {
			err := validateID(context.Background(), "_test_id")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestValidate_CatalogRequest_ID(t *testing.T) {
	Convey("Test ValidateCatalogRequest ID", t, func() {
		Convey("Invalid ID", func() {
			req := &interfaces.CatalogRequest{
				ID:   strings.Repeat("a", 41),
				Name: "test-catalog",
			}
			err := ValidateCatalogRequest(context.Background(), req)
			So(err, ShouldNotBeNil)
		})

		Convey("Valid ID", func() {
			req := &interfaces.CatalogRequest{
				ID:   "test-catalog_1",
				Name: "test-catalog",
			}
			err := ValidateCatalogRequest(context.Background(), req)
			So(err, ShouldBeNil)
		})
	})
}

func TestValidate_ResourceRequest_ID(t *testing.T) {
	Convey("Test ValidateResourceRequest ID", t, func() {
		Convey("Invalid ID", func() {
			req := &interfaces.ResourceRequest{
				ID:   "test.resource",
				Name: "test-resource",
			}
			err := ValidateResourceRequest(context.Background(), req)
			So(err, ShouldNotBeNil)
		})

		Convey("Valid ID", func() {
			req := &interfaces.ResourceRequest{
				ID:   "test-resource_1",
				Name: "test-resource",
			}
			err := ValidateResourceRequest(context.Background(), req)
			So(err, ShouldBeNil)
		})
	})
}

func TestValidate_CreateResourceCategory(t *testing.T) {
	Convey("Test validateCreateResourceCategory", t, func() {
		Convey("Allow dataset", func() {
			err := validateCreateResourceCategory(context.Background(), interfaces.ResourceCategoryDataset)
			So(err, ShouldBeNil)
		})

		Convey("Allow logicview", func() {
			err := validateCreateResourceCategory(context.Background(), interfaces.ResourceCategoryLogicView)
			So(err, ShouldBeNil)
		})

		Convey("Reject discover-owned categories", func() {
			rejected := []string{
				interfaces.ResourceCategoryTable,
				interfaces.ResourceCategoryFile,
				interfaces.ResourceCategoryFileset,
				interfaces.ResourceCategoryAPI,
				interfaces.ResourceCategoryMetric,
				interfaces.ResourceCategoryTopic,
				interfaces.ResourceCategoryIndex,
			}
			for _, category := range rejected {
				err := validateCreateResourceCategory(context.Background(), category)
				So(err, ShouldNotBeNil)
			}
		})

		Convey("Reject empty category", func() {
			err := validateCreateResourceCategory(context.Background(), "")
			So(err, ShouldNotBeNil)
		})

		Convey("Reject unknown category", func() {
			err := validateCreateResourceCategory(context.Background(), "foo")
			So(err, ShouldNotBeNil)
		})

		Convey("Category match is case sensitive", func() {
			err := validateCreateResourceCategory(context.Background(), "Dataset")
			So(err, ShouldNotBeNil)
		})
	})
}
func TestValidate_DatasetRequest(t *testing.T) {
	baseReq := func(props []*interfaces.Property) *interfaces.ResourceRequest {
		return &interfaces.ResourceRequest{
			Name:             "ds",
			Category:         interfaces.ResourceCategoryDataset,
			SchemaDefinition: props,
		}
	}

	Convey("Test ValidateResourceRequest for dataset", t, func() {
		Convey("Reject nil schema_definition", func() {
			err := ValidateResourceRequest(context.Background(), baseReq(nil))
			So(err, ShouldNotBeNil)
		})

		Convey("Reject empty schema_definition", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{}))
			So(err, ShouldNotBeNil)
		})

		Convey("Reject empty field name", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{
				{Name: "", Type: interfaces.DataType_String},
			}))
			So(err, ShouldNotBeNil)
		})

		Convey("Reject field name length exceeded", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{
				{Name: strings.Repeat("a", interfaces.MaxLength_PropertyName+1), Type: interfaces.DataType_String},
			}))
			So(err, ShouldNotBeNil)
		})

		Convey("Reject display name length exceeded", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{
				{Name: "f1", DisplayName: strings.Repeat("a", interfaces.MaxLength_PropertyDisplayName+1), Type: interfaces.DataType_String},
			}))
			So(err, ShouldNotBeNil)
		})

		Convey("Reject description length exceeded", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{
				{Name: "f1", Description: strings.Repeat("a", interfaces.MaxLength_PropertyDescription+1), Type: interfaces.DataType_String},
			}))
			So(err, ShouldNotBeNil)
		})

		Convey("Reject duplicate field name", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{
				{Name: "f1", Type: interfaces.DataType_String},
				{Name: "f1", Type: interfaces.DataType_Integer},
			}))
			So(err, ShouldNotBeNil)
		})

		Convey("Reject duplicate display name", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{
				{Name: "f1", DisplayName: "same", Type: interfaces.DataType_String},
				{Name: "f2", DisplayName: "same", Type: interfaces.DataType_String},
			}))
			So(err, ShouldNotBeNil)
		})

		Convey("Reject invalid feature type", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{
				{Name: "f1", Type: interfaces.DataType_Text, Features: []interfaces.PropertyFeature{
					{FeatureName: "feat1", FeatureType: "bogus", RefProperty: "f1", IsNative: true},
				}},
			}))
			So(err, ShouldNotBeNil)
		})

		Convey("Reject non-native feature with missing ref_property", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{
				{Name: "f1", Type: interfaces.DataType_Text, Features: []interfaces.PropertyFeature{
					{FeatureName: "feat1", FeatureType: interfaces.PropertyFeatureType_Fulltext, RefProperty: "missing"},
				}},
			}))
			So(err, ShouldNotBeNil)
		})

		Convey("Reject feature with mismatched ref type", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{
				{Name: "f1", Type: interfaces.DataType_Integer},
				{Name: "f2", Type: interfaces.DataType_Text, Features: []interfaces.PropertyFeature{
					{FeatureName: "feat1", FeatureType: interfaces.PropertyFeatureType_Keyword, RefProperty: "f1"},
				}},
			}))
			So(err, ShouldNotBeNil)
		})

		Convey("Accept minimal valid dataset", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{
				{Name: "id", Type: interfaces.DataType_Integer},
				{Name: "name", Type: interfaces.DataType_String},
			}))
			So(err, ShouldBeNil)
		})

		Convey("Accept dataset with native fulltext feature", func() {
			err := ValidateResourceRequest(context.Background(), baseReq([]*interfaces.Property{
				{Name: "body", Type: interfaces.DataType_Text, Features: []interfaces.PropertyFeature{
					{FeatureName: "body.ft", FeatureType: interfaces.PropertyFeatureType_Fulltext, RefProperty: "body", IsNative: true},
				}},
			}))
			So(err, ShouldBeNil)
		})
	})
}

func TestValidate_Tags(t *testing.T) {
	Convey("Test ValidateTags", t, func() {
		Convey("Valid tags", func() {
			err := ValidateTags(context.Background(), []string{"tag1", "tag2"})
			So(err, ShouldBeNil)
		})

		Convey("Empty tags", func() {
			err := ValidateTags(context.Background(), []string{})
			So(err, ShouldBeNil)
		})

		Convey("Exceeds max number", func() {
			tags := make([]string, interfaces.TAGS_MAX_NUMBER+1)
			for i := range tags {
				tags[i] = "tag"
			}
			err := ValidateTags(context.Background(), tags)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid tag in list", func() {
			err := ValidateTags(context.Background(), []string{"good-tag", "bad/tag"})
			So(err, ShouldNotBeNil)
		})
	})
}

func TestValidate_Tag(t *testing.T) {
	Convey("Test validateTag", t, func() {
		Convey("Valid tag", func() {
			err := validateTag(context.Background(), "my-tag")
			So(err, ShouldBeNil)
		})

		Convey("Empty tag", func() {
			err := validateTag(context.Background(), "")
			So(err, ShouldNotBeNil)
		})

		Convey("Only spaces", func() {
			err := validateTag(context.Background(), "   ")
			So(err, ShouldNotBeNil)
		})

		Convey("Exceeds max length", func() {
			tag := strings.Repeat("a", interfaces.TAG_MAX_LENGTH+1)
			err := validateTag(context.Background(), tag)
			So(err, ShouldNotBeNil)
		})

		Convey("Special chars", func() {
			invalidChars := []string{"/", ":", "?", "\\", "\"", "<", ">", "|", "#", "%", "&", "*", "$", "^", "!", "=", "."}
			for _, ch := range invalidChars {
				err := validateTag(context.Background(), "tag"+ch+"name")
				So(err, ShouldNotBeNil)
			}
		})

		Convey("Trim spaces", func() {
			err := validateTag(context.Background(), "  valid-tag  ")
			So(err, ShouldBeNil)
		})
	})
}

func TestValidate_Description(t *testing.T) {
	Convey("Test validateDescription", t, func() {
		Convey("Valid description", func() {
			err := validateDescription(context.Background(), "A valid description")
			So(err, ShouldBeNil)
		})

		Convey("Empty description", func() {
			err := validateDescription(context.Background(), "")
			So(err, ShouldBeNil)
		})

		Convey("Exceeds max length", func() {
			desc := strings.Repeat("a", interfaces.DESCRIPTION_MAX_LENGTH+1)
			err := validateDescription(context.Background(), desc)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestValidate_PaginationQueryParams(t *testing.T) {
	Convey("Test validatePaginationQueryParams", t, func() {
		Convey("Valid pagination", func() {
			params, err := validatePaginationQueryParams(context.Background(),
				"0", "10", "name", "asc", testSortTypes)
			So(err, ShouldBeNil)
			So(params.Offset, ShouldEqual, 0)
			So(params.Limit, ShouldEqual, 10)
			So(params.Sort, ShouldEqual, "f_name")
			So(params.Direction, ShouldEqual, "asc")
		})

		Convey("No limit", func() {
			params, err := validatePaginationQueryParams(context.Background(),
				"0", "-1", "name", "desc", testSortTypes)
			So(err, ShouldBeNil)
			So(params.Limit, ShouldEqual, -1)
		})

		Convey("Invalid offset", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"abc", "10", "name", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Negative offset", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"-1", "10", "name", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid limit", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"0", "abc", "name", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Limit too small", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"0", "0", "name", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Limit too large", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"0", "1001", "name", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid sort", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"0", "10", "unknown_sort", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid direction", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"0", "10", "name", "up", testSortTypes)
			So(err, ShouldNotBeNil)
		})
	})
}
