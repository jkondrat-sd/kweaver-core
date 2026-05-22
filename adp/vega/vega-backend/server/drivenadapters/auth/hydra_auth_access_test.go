package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/common"
)

type fakeHydra struct {
	visitor hydra.Visitor
	err     error
}

func (h fakeHydra) Introspect(ctx context.Context, token string) (hydra.TokenIntrospectInfo, error) {
	return hydra.TokenIntrospectInfo{}, nil
}

func (h fakeHydra) VerifyToken(ctx context.Context, c *gin.Context) (hydra.Visitor, error) {
	return h.visitor, h.err
}

func TestHydraAuthAccess_VerifyToken(t *testing.T) {
	Convey("Test hydraAuthAccess.VerifyToken", t, func() {
		ctx := context.Background()
		ginCtx := &gin.Context{}

		Convey("returns visitor when hydra verifies token", func() {
			expectedVisitor := hydra.Visitor{ID: "user-1", ClientID: "client-1"}
			access := &hydraAuthAccess{
				appSetting: &common.AppSetting{},
				hydra:      fakeHydra{visitor: expectedVisitor},
			}

			visitor, err := access.VerifyToken(ctx, ginCtx)
			So(err, ShouldBeNil)
			So(visitor, ShouldResemble, expectedVisitor)
		})

		Convey("returns error when hydra rejects token", func() {
			expectedErr := errors.New("invalid token")
			access := &hydraAuthAccess{
				appSetting: &common.AppSetting{},
				hydra:      fakeHydra{err: expectedErr},
			}

			visitor, err := access.VerifyToken(ctx, ginCtx)
			So(err, ShouldEqual, expectedErr)
			So(visitor, ShouldResemble, hydra.Visitor{})
		})
	})
}
