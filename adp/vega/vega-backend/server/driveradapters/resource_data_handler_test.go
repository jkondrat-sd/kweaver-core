package driveradapters

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"vega-backend/common"
	"vega-backend/interfaces"
	vmock "vega-backend/interfaces/mock"
)

func newResourceDataTestEngine(handler *restHandler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	handler.RegisterPublic(engine)
	return engine
}

func newDatasetResource() *interfaces.Resource {
	return &interfaces.Resource{
		ID:       "resource-1",
		Name:     "orders",
		Category: interfaces.ResourceCategoryDataset,
		Status:   interfaces.ResourceStatusActive,
	}
}

func Test_ResourceDataRestHandler_PostResourceDataByIn(t *testing.T) {
	Convey("Test ResourceDataHandler PostResourceDataByIn\n", t, func() {
		restore := setGinMode()
		defer restore()

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		rs := vmock.NewMockResourceService(mockCtrl)
		ds := vmock.NewMockDatasetService(mockCtrl)
		rds := vmock.NewMockResourceDataService(mockCtrl)
		handler := MockNewRestHandler(&common.AppSetting{}, nil, nil, rs, nil, ds, nil, nil, nil, rds, nil)
		engine := newResourceDataTestEngine(handler)
		url := "/api/vega-backend/in/v1/resources/resource-1/data"

		Convey("Invalid override method", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{}`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPatch)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.InvalidParameter.OverrideMethod")
		})

		Convey("Query resource data success", func() {
			rs.EXPECT().GetByID(gomock.Any(), "resource-1").Return(newDatasetResource(), nil)
			rds.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, resource *interfaces.Resource, params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
					So(resource.ID, ShouldEqual, "resource-1")
					So(params.NeedTotal, ShouldBeTrue)
					return []map[string]any{{"id": "doc-1"}}, int64(1), nil
				})

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"need_total":true}`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldContainSubstring, "doc-1")
			So(w.Body.String(), ShouldContainSubstring, "total_count")
		})

		Convey("Query resource data returns not found when resource is missing", func() {
			rs.EXPECT().GetByID(gomock.Any(), "resource-1").Return(nil, nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{}`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.Resource.NotFound")
		})

		Convey("Create documents success", func() {
			rs.EXPECT().GetByID(gomock.Any(), "resource-1").Return(newDatasetResource(), nil)
			ds.EXPECT().CreateDocuments(gomock.Any(), "resource-1", gomock.Any()).
				DoAndReturn(func(_ context.Context, id string, documents []map[string]any) ([]string, error) {
					So(len(documents), ShouldEqual, 1)
					So(documents[0]["name"], ShouldEqual, "order-1")
					return []string{"doc-1"}, nil
				})

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`[{"name":"order-1"}]`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
			So(w.Body.String(), ShouldContainSubstring, "doc-1")
		})

		Convey("Create documents rejects non-dataset resource", func() {
			rs.EXPECT().GetByID(gomock.Any(), "resource-1").Return(&interfaces.Resource{
				ID:       "resource-1",
				Category: interfaces.ResourceCategoryTable,
				Status:   interfaces.ResourceStatusActive,
			}, nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`[{"name":"order-1"}]`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPost)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "operation requires resource category=dataset")
		})

		Convey("Delete by query requires filter", func() {
			rs.EXPECT().GetByID(gomock.Any(), "resource-1").Return(newDatasetResource(), nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{}`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodDelete)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "filter is required for delete-by-query")
		})
	})
}

func Test_ResourceDataRestHandler_PutResourceDataByIn(t *testing.T) {
	Convey("Test ResourceDataHandler PutResourceDataByIn\n", t, func() {
		restore := setGinMode()
		defer restore()

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		rs := vmock.NewMockResourceService(mockCtrl)
		ds := vmock.NewMockDatasetService(mockCtrl)
		handler := MockNewRestHandler(&common.AppSetting{}, nil, nil, rs, nil, ds, nil, nil, nil, nil, nil)
		engine := newResourceDataTestEngine(handler)
		url := "/api/vega-backend/in/v1/resources/resource-1/data"

		Convey("Reject empty documents", func() {
			rs.EXPECT().GetByID(gomock.Any(), "resource-1").Return(newDatasetResource(), nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBufferString(`[]`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "documents array cannot be empty")
		})

		Convey("Reject documents without id", func() {
			rs.EXPECT().GetByID(gomock.Any(), "resource-1").Return(newDatasetResource(), nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBufferString(`[{"name":"order-1"}]`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "invalid_indexes")
		})

		Convey("Upsert documents success", func() {
			rs.EXPECT().GetByID(gomock.Any(), "resource-1").Return(newDatasetResource(), nil)
			ds.EXPECT().UpsertDocuments(gomock.Any(), "resource-1", gomock.Any()).
				Return([]string{"doc-1"}, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBufferString(`[{"id":"doc-1","name":"order-1"}]`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldContainSubstring, "doc-1")
		})
	})
}

func Test_ResourceDataRestHandler_DocumentByIn(t *testing.T) {
	Convey("Test ResourceDataHandler document APIs\n", t, func() {
		restore := setGinMode()
		defer restore()

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		rs := vmock.NewMockResourceService(mockCtrl)
		ds := vmock.NewMockDatasetService(mockCtrl)
		handler := MockNewRestHandler(&common.AppSetting{}, nil, nil, rs, nil, ds, nil, nil, nil, nil, nil)
		engine := newResourceDataTestEngine(handler)

		Convey("Get document success", func() {
			rs.EXPECT().GetByID(gomock.Any(), "resource-1").Return(newDatasetResource(), nil)
			ds.EXPECT().GetDocument(gomock.Any(), "resource-1", "doc-1").Return(map[string]any{"id": "doc-1"}, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/vega-backend/in/v1/resources/resource-1/data/doc-1", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldContainSubstring, "doc-1")
		})

		Convey("Put document overrides body id with path id", func() {
			rs.EXPECT().GetByID(gomock.Any(), "resource-1").Return(newDatasetResource(), nil)
			ds.EXPECT().UpsertDocuments(gomock.Any(), "resource-1", gomock.Any()).
				DoAndReturn(func(_ context.Context, id string, documents []map[string]any) ([]string, error) {
					So(documents, ShouldHaveLength, 1)
					So(documents[0]["id"], ShouldEqual, "doc-1")
					return []string{"doc-1"}, nil
				})

			req := httptest.NewRequest(http.MethodPut, "/api/vega-backend/in/v1/resources/resource-1/data/doc-1", bytes.NewBufferString(`{"id":"wrong","name":"order-1"}`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldContainSubstring, "doc-1")
		})

		Convey("Delete documents success", func() {
			rs.EXPECT().GetByID(gomock.Any(), "resource-1").Return(newDatasetResource(), nil)
			ds.EXPECT().DeleteDocuments(gomock.Any(), "resource-1", "doc-1,doc-2").Return(nil)

			req := httptest.NewRequest(http.MethodDelete, "/api/vega-backend/in/v1/resources/resource-1/data/doc-1,doc-2", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}
