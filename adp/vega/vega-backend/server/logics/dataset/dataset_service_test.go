// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package dataset

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
)

func TestDatasetService_Create_Success(t *testing.T) {
	Convey("Test DatasetService Create success", t, func() {
		connector := &fakeIndexConnector{
			createFunc: func(ctx context.Context, name string, schemaDefinition []*interfaces.Property) error {
				So(name, ShouldEqual, "resource-1")
				So(schemaDefinition, ShouldHaveLength, 1)
				return nil
			},
		}
		svc := &datasetService{c: connector}
		err := svc.Create(context.Background(), &interfaces.Resource{
			ID:               "resource-1",
			SchemaDefinition: []*interfaces.Property{{Name: "title"}},
		})
		So(err, ShouldBeNil)
	})
}

func TestDatasetService_Create_Error(t *testing.T) {
	Convey("Test DatasetService Create wraps connector error", t, func() {
		connector := &fakeIndexConnector{createFunc: func(ctx context.Context, name string, schemaDefinition []*interfaces.Property) error {
			return errors.New("create failed")
		}}
		svc := &datasetService{c: connector}
		err := svc.Create(context.Background(), &interfaces.Resource{ID: "resource-1"})
		assertDatasetHTTPError(err, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_CreateFailed)
	})
}

func TestDatasetService_Update_UsesSourceIdentifierAndID(t *testing.T) {
	Convey("Test DatasetService Update uses source identifier and id", t, func() {
		connector := &fakeIndexConnector{updateFunc: func(ctx context.Context, name string, schemaDefinition []*interfaces.Property) error {
			So(name, ShouldEqual, "source-1-resource-1")
			return nil
		}}
		svc := &datasetService{c: connector}
		err := svc.Update(context.Background(), &interfaces.Resource{ID: "resource-1", SourceIdentifier: "source-1"})
		So(err, ShouldBeNil)
	})
}

func TestDatasetService_Delete_SkipsMissingDataset(t *testing.T) {
	Convey("Test DatasetService Delete skips missing dataset", t, func() {
		connector := &fakeIndexConnector{checkExistFunc: func(ctx context.Context, name string) (bool, error) {
			So(name, ShouldEqual, "resource-1")
			return false, nil
		}}
		svc := &datasetService{c: connector}
		err := svc.Delete(context.Background(), "resource-1")
		So(err, ShouldBeNil)
	})
}

func TestDatasetService_Delete_DeletesExistingDataset(t *testing.T) {
	Convey("Test DatasetService Delete deletes existing dataset", t, func() {
		connector := &fakeIndexConnector{
			checkExistFunc: func(ctx context.Context, name string) (bool, error) {
				return true, nil
			},
			deleteFunc: func(ctx context.Context, name string) error {
				So(name, ShouldEqual, "resource-1")
				return nil
			},
		}
		svc := &datasetService{c: connector}
		err := svc.Delete(context.Background(), "resource-1")
		So(err, ShouldBeNil)
	})
}

func TestDatasetService_CheckExist_Error(t *testing.T) {
	Convey("Test DatasetService CheckExist wraps connector error", t, func() {
		connector := &fakeIndexConnector{checkExistFunc: func(ctx context.Context, name string) (bool, error) {
			return false, errors.New("check failed")
		}}
		svc := &datasetService{c: connector}
		exists, err := svc.CheckExist(context.Background(), "resource-1")
		So(exists, ShouldBeFalse)
		assertDatasetHTTPError(err, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError)
	})
}

func TestDatasetService_ListDocuments_Success(t *testing.T) {
	Convey("Test DatasetService ListDocuments success", t, func() {
		connector := &fakeIndexConnector{executeQueryFunc: func(ctx context.Context, indexName string, resource *interfaces.Resource, params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error) {
			So(indexName, ShouldEqual, "index-1")
			return &interfaces.QueryResult{
				Rows:  []map[string]any{{"title": "hello"}},
				Total: 1,
			}, nil
		}}
		svc := &datasetService{c: connector}
		rows, total, err := svc.ListDocuments(context.Background(), "index-1", &interfaces.Resource{}, &interfaces.ResourceDataQueryParams{})
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 1)
		So(rows, ShouldResemble, []map[string]any{{"title": "hello"}})
	})
}

func TestDatasetService_CreateDocuments_Success(t *testing.T) {
	Convey("Test DatasetService CreateDocuments success", t, func() {
		connector := &fakeIndexConnector{createDocumentsFunc: func(ctx context.Context, name string, documents []map[string]any) ([]string, error) {
			So(name, ShouldEqual, "resource-1")
			So(documents, ShouldHaveLength, 1)
			return []string{"doc-1"}, nil
		}}
		svc := &datasetService{c: connector}
		docIDs, err := svc.CreateDocuments(context.Background(), "resource-1", []map[string]any{{"title": "hello"}})
		So(err, ShouldBeNil)
		So(docIDs, ShouldResemble, []string{"doc-1"})
	})
}

func TestDatasetService_GetDocument_Success(t *testing.T) {
	Convey("Test DatasetService GetDocument success", t, func() {
		connector := &fakeIndexConnector{getDocumentFunc: func(ctx context.Context, name string, docID string) (map[string]any, error) {
			So(name, ShouldEqual, "resource-1")
			So(docID, ShouldEqual, "doc-1")
			return map[string]any{"title": "hello"}, nil
		}}
		svc := &datasetService{c: connector}
		document, err := svc.GetDocument(context.Background(), "resource-1", "doc-1")
		So(err, ShouldBeNil)
		So(document, ShouldResemble, map[string]any{"title": "hello"})
	})
}

func TestDatasetService_DeleteDocument_Success(t *testing.T) {
	Convey("Test DatasetService DeleteDocument success", t, func() {
		connector := &fakeIndexConnector{deleteDocumentFunc: func(ctx context.Context, name string, docID string) error {
			So(name, ShouldEqual, "resource-1")
			So(docID, ShouldEqual, "doc-1")
			return nil
		}}
		svc := &datasetService{c: connector}
		err := svc.DeleteDocument(context.Background(), "resource-1", "doc-1")
		So(err, ShouldBeNil)
	})
}

func TestDatasetService_UpsertDocuments_ReturnsConnectorErrorAsIs(t *testing.T) {
	Convey("Test DatasetService UpsertDocuments returns connector error as is", t, func() {
		expectedErr := errors.New("upsert failed")
		connector := &fakeIndexConnector{upsertDocumentsFunc: func(ctx context.Context, name string, updateRequests []map[string]any) ([]string, error) {
			return []string{"doc-1"}, expectedErr
		}}
		svc := &datasetService{c: connector}
		docIDs, err := svc.UpsertDocuments(context.Background(), "resource-1", []map[string]any{{"id": "doc-1"}})
		So(err, ShouldEqual, expectedErr)
		So(docIDs, ShouldResemble, []string{"doc-1"})
	})
}

func TestDatasetService_DeleteDocuments_Success(t *testing.T) {
	Convey("Test DatasetService DeleteDocuments success", t, func() {
		connector := &fakeIndexConnector{deleteDocumentsFunc: func(ctx context.Context, name string, docIDs string) error {
			So(name, ShouldEqual, "resource-1")
			So(docIDs, ShouldEqual, "doc-1,doc-2")
			return nil
		}}
		svc := &datasetService{c: connector}
		err := svc.DeleteDocuments(context.Background(), "resource-1", "doc-1,doc-2")
		So(err, ShouldBeNil)
	})
}

func assertDatasetHTTPError(err error, status int, errorCode string) {
	So(err, ShouldNotBeNil)
	httpErr, ok := err.(*rest.HTTPError)
	So(ok, ShouldBeTrue)
	So(httpErr.HTTPCode, ShouldEqual, status)
	So(httpErr.BaseError.ErrorCode, ShouldEqual, errorCode)
}

type fakeIndexConnector struct {
	createFunc                 func(context.Context, string, []*interfaces.Property) error
	updateFunc                 func(context.Context, string, []*interfaces.Property) error
	deleteFunc                 func(context.Context, string) error
	checkExistFunc             func(context.Context, string) (bool, error)
	executeQueryFunc           func(context.Context, string, *interfaces.Resource, *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error)
	createDocumentsFunc        func(context.Context, string, []map[string]any) ([]string, error)
	getDocumentFunc            func(context.Context, string, string) (map[string]any, error)
	deleteDocumentFunc         func(context.Context, string, string) error
	upsertDocumentsFunc        func(context.Context, string, []map[string]any) ([]string, error)
	deleteDocumentsFunc        func(context.Context, string, string) error
	deleteDocumentsByQueryFunc func(context.Context, string, *interfaces.ResourceDataQueryParams, []*interfaces.Property) error
}

func (f *fakeIndexConnector) GetType() string { return "index" }
func (f *fakeIndexConnector) GetName() string { return "Index" }
func (f *fakeIndexConnector) GetMode() string { return interfaces.ConnectorModeLocal }
func (f *fakeIndexConnector) GetCategory() string {
	return interfaces.ConnectorCategoryIndex
}
func (f *fakeIndexConnector) GetEnabled() bool        { return true }
func (f *fakeIndexConnector) SetEnabled(enabled bool) {}
func (f *fakeIndexConnector) GetSensitiveFields() []string {
	return nil
}
func (f *fakeIndexConnector) GetFieldConfig() map[string]interfaces.ConnectorFieldConfig {
	return nil
}
func (f *fakeIndexConnector) New(cfg interfaces.ConnectorConfig) (connectors.Connector, error) {
	return f, nil
}
func (f *fakeIndexConnector) Connect(ctx context.Context) error        { return nil }
func (f *fakeIndexConnector) Ping(ctx context.Context) error           { return nil }
func (f *fakeIndexConnector) Close(ctx context.Context) error          { return nil }
func (f *fakeIndexConnector) TestConnection(ctx context.Context) error { return nil }
func (f *fakeIndexConnector) GetMetadata(ctx context.Context) (map[string]any, error) {
	return nil, nil
}
func (f *fakeIndexConnector) MapType(nativeType string) string { return nativeType }
func (f *fakeIndexConnector) ListIndexes(ctx context.Context) ([]*interfaces.IndexMeta, error) {
	return nil, nil
}
func (f *fakeIndexConnector) GetIndexMeta(ctx context.Context, index *interfaces.IndexMeta) error {
	return nil
}
func (f *fakeIndexConnector) ExecuteQuery(ctx context.Context, indexName string, resource *interfaces.Resource, params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error) {
	return f.executeQueryFunc(ctx, indexName, resource, params)
}
func (f *fakeIndexConnector) ExecuteQueryWithDsl(ctx context.Context, resourceName string, dsl string) (*interfaces.QueryResult, error) {
	return nil, nil
}
func (f *fakeIndexConnector) ExecuteRawQuery(ctx context.Context, index string, query map[string]any) (*interfaces.RawQueryResponse, error) {
	return nil, nil
}
func (f *fakeIndexConnector) Create(ctx context.Context, name string, schemaDefinition []*interfaces.Property) error {
	return f.createFunc(ctx, name, schemaDefinition)
}
func (f *fakeIndexConnector) Update(ctx context.Context, name string, schemaDefinition []*interfaces.Property) error {
	return f.updateFunc(ctx, name, schemaDefinition)
}
func (f *fakeIndexConnector) Delete(ctx context.Context, name string) error {
	return f.deleteFunc(ctx, name)
}
func (f *fakeIndexConnector) CheckExist(ctx context.Context, name string) (bool, error) {
	return f.checkExistFunc(ctx, name)
}
func (f *fakeIndexConnector) CreateDocuments(ctx context.Context, name string, documents []map[string]any) ([]string, error) {
	return f.createDocumentsFunc(ctx, name, documents)
}
func (f *fakeIndexConnector) GetDocument(ctx context.Context, name string, docID string) (map[string]any, error) {
	return f.getDocumentFunc(ctx, name, docID)
}
func (f *fakeIndexConnector) DeleteDocument(ctx context.Context, name string, docID string) error {
	return f.deleteDocumentFunc(ctx, name, docID)
}
func (f *fakeIndexConnector) UpsertDocuments(ctx context.Context, name string, updateRequests []map[string]any) ([]string, error) {
	return f.upsertDocumentsFunc(ctx, name, updateRequests)
}
func (f *fakeIndexConnector) DeleteDocuments(ctx context.Context, name string, docIDs string) error {
	return f.deleteDocumentsFunc(ctx, name, docIDs)
}
func (f *fakeIndexConnector) DeleteDocumentsByQuery(ctx context.Context, name string, params *interfaces.ResourceDataQueryParams, schemaDefinition []*interfaces.Property) error {
	return f.deleteDocumentsByQueryFunc(ctx, name, params, schemaDefinition)
}
