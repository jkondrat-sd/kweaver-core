package driveradapters

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/common"
	"vega-backend/interfaces"
)

func Test_QueryRestHandler_RawQueryByInValidation(t *testing.T) {
	Convey("Test QueryHandler RawQueryByIn validation\n", t, func() {
		restore := setGinMode()
		defer restore()

		engine := gin.New()
		engine.Use(gin.Recovery())

		handler := MockNewRestHandler(&common.AppSetting{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		handler.RegisterPublic(engine)

		url := "/api/vega-backend/in/v1/resources/query"

		Convey("Invalid request body", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.InvalidParameter.RequestBody")
		})

		Convey("Missing resource type", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"query":"select 1"}`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.InvalidParameter.ResourceType")
			So(w.Body.String(), ShouldContainSubstring, "resource_type is required")
		})

		Convey("Unsupported resource type", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"resource_type":"oracle","query":"select 1"}`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.InvalidParameter.ResourceType")
			So(w.Body.String(), ShouldContainSubstring, "got: oracle")
		})

		Convey("Invalid stream size", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"resource_type":"mysql","query":"select 1","stream_size":99}`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.InvalidParameter.StreamSize")
			So(w.Body.String(), ShouldContainSubstring, "stream_size must be between 100 and 10000")
		})

		Convey("Invalid query timeout", func() {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"resource_type":"mysql","query":"select 1","query_timeout":3601}`))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.Query.InvalidParameter.QueryTimeout")
			So(w.Body.String(), ShouldContainSubstring, "query_timeout must be between 1 and 3600")
		})
	})
}
