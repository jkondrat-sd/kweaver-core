package user_mgmt

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

func newTestUserMgmtAccess(appSetting *common.AppSetting, httpClient rest.HTTPClient) *userMgmtAccess {
	return &userMgmtAccess{
		appSetting:  appSetting,
		httpClient:  httpClient,
		userMgmtUrl: appSetting.UserMgmtUrl,
	}
}

func Test_userMgmtAccess_GetAccountNames(t *testing.T) {
	Convey("Test GetAccountNames", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{UserMgmtUrl: "http://test-user-mgmt"}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		uma := newTestUserMgmtAccess(appSetting, mockHTTPClient)

		Convey("Should return nil when accounts are empty", func() {
			err := uma.GetAccountNames(ctx, nil)
			So(err, ShouldBeNil)
		})

		Convey("Should fill account names and keep default name for missing accounts", func() {
			respData, _ := sonic.Marshal(map[string]any{
				"user_names": []map[string]string{{"id": "u1", "name": "Alice"}},
				"app_names":  []map[string]string{{"id": "a1", "name": "Robot"}},
			})

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), "http://test-user-mgmt/api/user-management/v2/names", gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, url string, headers map[string]string, body any) (int, []byte, error) {
					So(headers["Content-Type"], ShouldEqual, "application/json")
					requestBody := body.(map[string]any)
					So(requestBody["method"], ShouldEqual, http.MethodGet)
					So(requestBody["strict"], ShouldEqual, false)
					So(requestBody["user_ids"], ShouldResemble, []string{"u1", "u2"})
					So(requestBody["app_ids"], ShouldResemble, []string{"a1"})
					return http.StatusOK, respData, nil
				})

			accounts := []*interfaces.AccountInfo{
				{ID: "u1", Type: interfaces.ACCESSOR_TYPE_USER},
				{ID: "u1", Type: interfaces.ACCESSOR_TYPE_USER},
				{ID: "a1", Type: interfaces.ACCESSOR_TYPE_APP},
				{ID: "u2", Type: interfaces.ACCESSOR_TYPE_USER},
				{ID: "x1", Type: "unknown"},
			}

			err := uma.GetAccountNames(ctx, accounts)
			So(err, ShouldBeNil)
			So(accounts[0].Name, ShouldEqual, "Alice")
			So(accounts[1].Name, ShouldEqual, "Alice")
			So(accounts[2].Name, ShouldEqual, "Robot")
			So(accounts[3].Name, ShouldEqual, "-")
			So(accounts[4].Name, ShouldEqual, "")
		})

		Convey("Should return error when HTTP request fails", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, errors.New("network error"))

			err := uma.GetAccountNames(ctx, []*interfaces.AccountInfo{{ID: "u1", Type: interfaces.ACCESSOR_TYPE_USER}})
			So(err, ShouldNotBeNil)
		})

		Convey("Should return error when status is not OK", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusInternalServerError, []byte("internal error"), nil)

			err := uma.GetAccountNames(ctx, []*interfaces.AccountInfo{{ID: "u1", Type: interfaces.ACCESSOR_TYPE_USER}})
			So(err, ShouldNotBeNil)
		})

		Convey("Should return error when response cannot be unmarshaled", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("invalid json"), nil)

			err := uma.GetAccountNames(ctx, []*interfaces.AccountInfo{{ID: "u1", Type: interfaces.ACCESSOR_TYPE_USER}})
			So(err, ShouldNotBeNil)
		})
	})
}
