package driveradapters

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

func TestValidateResourceRequest(t *testing.T) {
	Convey("Test ValidateResourceRequest", t, func() {
		ctx := context.Background()

		Convey("Should validate dataset request", func() {
			req := &interfaces.ResourceRequest{
				Name:     "dataset",
				Category: interfaces.ResourceCategoryDataset,
				SchemaDefinition: []*interfaces.Property{
					{Name: "title", Type: interfaces.DataType_String},
				},
			}

			err := ValidateResourceRequest(ctx, req)
			So(err, ShouldBeNil)
			So(req.SchemaDefinition[0].DisplayName, ShouldEqual, "title")
		})

		Convey("Should reject dataset without schema definition", func() {
			req := &interfaces.ResourceRequest{
				Name:     "dataset",
				Category: interfaces.ResourceCategoryDataset,
			}

			err := ValidateResourceRequest(ctx, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.Dataset.InvalidParameter.SchemaDefinition")
		})

		Convey("Should reject duplicated dataset field name", func() {
			req := &interfaces.ResourceRequest{
				Name:     "dataset",
				Category: interfaces.ResourceCategoryDataset,
				SchemaDefinition: []*interfaces.Property{
					{Name: "title", Type: interfaces.DataType_String},
					{Name: "title", Type: interfaces.DataType_String},
				},
			}

			err := ValidateResourceRequest(ctx, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.Dataset.Duplicated.FieldName")
		})

		Convey("Should reject invalid dataset field type", func() {
			req := &interfaces.ResourceRequest{
				Name:     "dataset",
				Category: interfaces.ResourceCategoryDataset,
				SchemaDefinition: []*interfaces.Property{
					{Name: "title", Type: "invalid"},
				},
			}

			err := ValidateResourceRequest(ctx, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.Dataset.InvalidParameter.FieldType")
		})

		Convey("Should reject feature referencing missing field", func() {
			req := &interfaces.ResourceRequest{
				Name:     "dataset",
				Category: interfaces.ResourceCategoryDataset,
				SchemaDefinition: []*interfaces.Property{
					{
						Name: "title",
						Type: interfaces.DataType_String,
						Features: []interfaces.PropertyFeature{{
							FeatureName: "kw",
							FeatureType: interfaces.PropertyFeatureType_Keyword,
							RefProperty: "missing",
						}},
					},
				},
			}

			err := ValidateResourceRequest(ctx, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.Dataset.InvalidParameter.FieldFeatureRef")
		})

		Convey("Should reject field name length exceeded", func() {
			req := &interfaces.ResourceRequest{
				Name:     "dataset",
				Category: interfaces.ResourceCategoryDataset,
				SchemaDefinition: []*interfaces.Property{
					{Name: strings.Repeat("a", interfaces.MaxLength_PropertyName+1), Type: interfaces.DataType_String},
				},
			}

			err := ValidateResourceRequest(ctx, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VegaBackend.Dataset.LengthExceeded.FieldName")
		})
	})
}

func TestValidateResourceListQueryParams(t *testing.T) {
	Convey("Test ValidateResourceListQueryParams", t, func() {
		ctx := context.Background()

		Convey("Should accept valid category and status", func() {
			err := ValidateResourceListQueryParams(ctx, interfaces.ResourcesQueryParams{
				Category: interfaces.ResourceCategoryDataset,
				Status:   interfaces.ResourceStatusActive,
			})
			So(err, ShouldBeNil)
		})

		Convey("Should reject invalid category", func() {
			err := ValidateResourceListQueryParams(ctx, interfaces.ResourcesQueryParams{Category: "unknown"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid category: unknown")
		})

		Convey("Should reject invalid status", func() {
			err := ValidateResourceListQueryParams(ctx, interfaces.ResourcesQueryParams{Status: "unknown"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid status: unknown")
		})
	})
}

func TestValidateCreateResourceCategory(t *testing.T) {
	Convey("Test validateCreateResourceCategory", t, func() {
		ctx := context.Background()

		So(validateCreateResourceCategory(ctx, interfaces.ResourceCategoryDataset), ShouldBeNil)
		So(validateCreateResourceCategory(ctx, interfaces.ResourceCategoryLogicView), ShouldBeNil)

		err := validateCreateResourceCategory(ctx, interfaces.ResourceCategoryTable)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "cannot be created via API")
	})
}
