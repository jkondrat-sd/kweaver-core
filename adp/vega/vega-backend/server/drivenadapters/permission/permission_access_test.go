package permission

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"vega-backend/common"
	"vega-backend/interfaces"
)

func newTestPermissionAccess(appSetting *common.AppSetting, httpClient rest.HTTPClient) *permissionAccess {
	return &permissionAccess{
		appSetting:    appSetting,
		permissionUrl: appSetting.PermissionUrl,
		httpClient:    httpClient,
	}
}

func newPermissionCheck() interfaces.PermissionCheck {
	return interfaces.PermissionCheck{
		Accessor:   interfaces.PermissionAccessor{Type: interfaces.ACCESSOR_TYPE_USER, ID: "user-1"},
		Resource:   interfaces.PermissionResource{Type: interfaces.AUTH_RESOURCE_TYPE_RESOURCE, ID: "resource-1"},
		Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
	}
}

func newPermissionPolicy() interfaces.PermissionPolicy {
	return interfaces.PermissionPolicy{
		Accessor: interfaces.PermissionAccessor{Type: interfaces.ACCESSOR_TYPE_USER, ID: "user-1"},
		Resource: interfaces.PermissionResource{Type: interfaces.AUTH_RESOURCE_TYPE_RESOURCE, ID: "resource-1"},
		Operations: interfaces.PermissionPolicyOps{
			Allow: []interfaces.PermissionOperation{{Operation: interfaces.OPERATION_TYPE_VIEW_DETAIL}},
		},
	}
}

func Test_permissionAccess_CheckPermission(t *testing.T) {
	Convey("Test CheckPermission", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		pa := newTestPermissionAccess(&common.AppSetting{PermissionUrl: "http://permission"}, mockHTTPClient)

		Convey("Should return permission result", func() {
			respData, _ := sonic.Marshal(interfaces.PermissionCheckResult{Result: true})
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), "http://permission/operation-check", gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, url string, headers map[string]string, body any) (int, []byte, error) {
					So(headers[interfaces.CONTENT_TYPE_NAME], ShouldEqual, interfaces.CONTENT_TYPE_JSON)
					check := body.(interfaces.PermissionCheck)
					So(check.Method, ShouldEqual, http.MethodGet)
					return http.StatusOK, respData, nil
				})

			ok, err := pa.CheckPermission(ctx, newPermissionCheck())
			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)
		})

		Convey("Should return false when result body is nil", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, nil, nil)

			ok, err := pa.CheckPermission(ctx, newPermissionCheck())
			So(err, ShouldBeNil)
			So(ok, ShouldBeFalse)
		})

		Convey("Should return error when HTTP request fails", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, errors.New("network error"))

			ok, err := pa.CheckPermission(ctx, newPermissionCheck())
			So(err, ShouldNotBeNil)
			So(ok, ShouldBeFalse)
		})

		Convey("Should return HTTPError when status is not OK", func() {
			respData, _ := sonic.Marshal(PermissionError{Code: "Forbidden", Message: "denied"})
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusForbidden, respData, nil)

			ok, err := pa.CheckPermission(ctx, newPermissionCheck())
			So(ok, ShouldBeFalse)
			So(err, ShouldHaveSameTypeAs, &rest.HTTPError{})
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Should return error when response cannot be unmarshaled", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("invalid json"), nil)

			ok, err := pa.CheckPermission(ctx, newPermissionCheck())
			So(err, ShouldNotBeNil)
			So(ok, ShouldBeFalse)
		})
	})
}

func Test_permissionAccess_CreateResources(t *testing.T) {
	Convey("Test CreateResources", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		pa := newTestPermissionAccess(&common.AppSetting{PermissionUrl: "http://permission"}, mockHTTPClient)

		Convey("Should create resources", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), "http://permission/policy", gomock.Any(), gomock.Any()).
				Return(http.StatusNoContent, nil, nil)

			err := pa.CreateResources(ctx, []interfaces.PermissionPolicy{newPermissionPolicy()})
			So(err, ShouldBeNil)
		})

		Convey("Should return HTTPError when status is not NoContent", func() {
			respData, _ := sonic.Marshal(PermissionError{Code: "Invalid", Description: "bad request"})
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, respData, nil)

			err := pa.CreateResources(ctx, []interfaces.PermissionPolicy{newPermissionPolicy()})
			So(err, ShouldHaveSameTypeAs, &rest.HTTPError{})
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func Test_permissionAccess_DeleteResources(t *testing.T) {
	Convey("Test DeleteResources", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		pa := newTestPermissionAccess(&common.AppSetting{PermissionUrl: "http://permission"}, mockHTTPClient)

		Convey("Should delete resources with method override", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), "http://permission/policy-delete", gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, url string, headers map[string]string, body any) (int, []byte, error) {
					requestBody := body.(map[string]any)
					So(requestBody["method"], ShouldEqual, http.MethodDelete)
					So(requestBody["resources"], ShouldResemble, []interfaces.PermissionResource{{Type: interfaces.AUTH_RESOURCE_TYPE_RESOURCE, ID: "resource-1"}})
					return http.StatusNoContent, nil, nil
				})

			err := pa.DeleteResources(ctx, []interfaces.PermissionResource{{Type: interfaces.AUTH_RESOURCE_TYPE_RESOURCE, ID: "resource-1"}})
			So(err, ShouldBeNil)
		})

		Convey("Should return error when HTTP request fails", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, errors.New("network error"))

			err := pa.DeleteResources(ctx, []interfaces.PermissionResource{{Type: interfaces.AUTH_RESOURCE_TYPE_RESOURCE, ID: "resource-1"}})
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_permissionAccess_FilterResources(t *testing.T) {
	Convey("Test FilterResources", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		pa := newTestPermissionAccess(&common.AppSetting{PermissionUrl: "http://permission"}, mockHTTPClient)
		filter := interfaces.PermissionResourcesFilter{
			Accessor:   interfaces.PermissionAccessor{Type: interfaces.ACCESSOR_TYPE_USER, ID: "user-1"},
			Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
		}

		Convey("Should return allowed operations by resource", func() {
			respData, _ := sonic.Marshal([]map[string]any{
				{"id": "resource-1", "allow_operation": []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}},
			})
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), "http://permission/resource-filter", gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, url string, headers map[string]string, body any) (int, []byte, error) {
					requestBody := body.(interfaces.PermissionResourcesFilter)
					So(requestBody.Method, ShouldEqual, http.MethodGet)
					return http.StatusOK, respData, nil
				})

			ops, err := pa.FilterResources(ctx, filter)
			So(err, ShouldBeNil)
			So(ops, ShouldResemble, map[string]interfaces.PermissionResourceOps{
				"resource-1": {ResourceID: "resource-1", Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}},
			})
		})

		Convey("Should return empty map when body is nil", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, nil, nil)

			ops, err := pa.FilterResources(ctx, filter)
			So(err, ShouldBeNil)
			So(ops, ShouldResemble, map[string]interfaces.PermissionResourceOps{})
		})

		Convey("Should return error when response cannot be unmarshaled", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("invalid json"), nil)

			ops, err := pa.FilterResources(ctx, filter)
			So(err, ShouldNotBeNil)
			So(ops, ShouldResemble, map[string]interfaces.PermissionResourceOps{})
		})
	})
}
