// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/common"
	"vega-backend/interfaces"
)

func TestNoopAuthService_VerifyToken(t *testing.T) {
	Convey("Test NoopAuthService.VerifyToken", t, func() {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		c.Request = req
		svc := NewNoopAuthService(nil)

		Convey("builds visitor from request headers", func() {
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "account-1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			req.Header.Set("X-Request-MAC", "00:11:22:33")
			req.Header.Set("User-Agent", "vega-test")

			visitor, err := svc.VerifyToken(context.Background(), c)
			So(err, ShouldBeNil)
			So(visitor.ID, ShouldEqual, "account-1")
			So(string(visitor.Type), ShouldEqual, "user")
			So(visitor.Mac, ShouldEqual, "00:11:22:33")
			So(visitor.UserAgent, ShouldEqual, "vega-test")
		})
	})
}

func TestNewAuthService(t *testing.T) {
	Convey("Test NewAuthService", t, func() {
		Convey("returns noop service when auth is disabled", func() {
			t.Setenv("AUTH_ENABLED", "false")
			resetAuthServiceSingleton()

			svc := NewAuthService(&common.AppSetting{})
			_, ok := svc.(*NoopAuthService)
			So(ok, ShouldBeTrue)
		})
	})
}

func resetAuthServiceSingleton() {
	authServiceOnce = sync.Once{}
	authService = nil
}
