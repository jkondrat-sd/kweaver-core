// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package catalog

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"

	"vega-backend/interfaces"
	vmock "vega-backend/interfaces/mock"

	. "github.com/smartystreets/goconvey/convey"
)

// mockCipher 实现 kwcrypto.Cipher 接口用于测试
// 注：kwcrypto.Cipher 是外部库接口，无 mockgen 生成的版本，手写 mock 是合理的
type mockCipher struct {
	decryptFunc func(ciphertext string) (string, error)
}

func (m *mockCipher) Encrypt(plaintext string) (string, error) {
	return "encrypted_" + plaintext, nil
}

func (m *mockCipher) Decrypt(ciphertext string) (string, error) {
	return m.decryptFunc(ciphertext)
}

func (m *mockCipher) Signature(data string) (string, error) {
	return "", nil
}

func TestCatalogService_validateAndDecryptSensitiveFields(t *testing.T) {
	Convey("Test catalogService.validateAndDecryptSensitiveFields", t, func() {
		Convey("without cipher returns plaintext as is", func() {
			cs := &catalogService{cipher: nil}
			config := map[string]any{"password": "secret123", "host": "localhost"}

			decrypted, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
			So(err, ShouldBeNil)
			So(decrypted["password"], ShouldEqual, "secret123")
			So(config["password"], ShouldEqual, "secret123")
		})

		Convey("with cipher decrypts and rewrites config with prefix", func() {
			cs := &catalogService{
				cipher: &mockCipher{
					decryptFunc: func(ciphertext string) (string, error) {
						return "decrypted_" + ciphertext, nil
					},
				},
			}
			config := map[string]any{"password": "rsa_ciphertext", "host": "localhost"}

			decrypted, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
			So(err, ShouldBeNil)
			So(decrypted["password"], ShouldEqual, "decrypted_rsa_ciphertext")
			So(config["password"], ShouldEqual, EncryptedPrefix+"rsa_ciphertext")
			So(decrypted["host"], ShouldEqual, "localhost")
		})

		Convey("returns error when cipher decrypt fails", func() {
			cs := &catalogService{
				cipher: &mockCipher{
					decryptFunc: func(ciphertext string) (string, error) {
						return "", fmt.Errorf("invalid ciphertext")
					},
				},
			}
			config := map[string]any{"password": "bad_data"}

			_, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "password")
		})

		Convey("skips empty value without invoking cipher", func() {
			cs := &catalogService{
				cipher: &mockCipher{
					decryptFunc: func(ciphertext string) (string, error) {
						So("decrypt called", ShouldBeBlank)
						return "", nil
					},
				},
			}
			config := map[string]any{"password": ""}

			_, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
			So(err, ShouldBeNil)
		})

		Convey("skips non-string value without invoking cipher", func() {
			cs := &catalogService{
				cipher: &mockCipher{
					decryptFunc: func(ciphertext string) (string, error) {
						So("decrypt called", ShouldBeBlank)
						return "", nil
					},
				},
			}
			config := map[string]any{"password": 12345}

			_, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
			So(err, ShouldBeNil)
		})
	})
}

func TestCatalogService_decryptSensitiveFields(t *testing.T) {
	Convey("Test catalogService.decryptSensitiveFields", t, func() {
		Convey("without cipher returns stored value as is", func() {
			cs := &catalogService{cipher: nil}
			config := map[string]any{"password": "ENC:ciphertext"}

			decrypted, err := cs.decryptSensitiveFields([]string{"password"}, config)
			So(err, ShouldBeNil)
			So(decrypted["password"], ShouldEqual, "ENC:ciphertext")
		})

		Convey("with cipher strips prefix and decrypts", func() {
			cs := &catalogService{
				cipher: &mockCipher{
					decryptFunc: func(ciphertext string) (string, error) {
						return "plaintext_" + ciphertext, nil
					},
				},
			}
			config := map[string]any{"password": "ENC:rsa_data"}

			decrypted, err := cs.decryptSensitiveFields([]string{"password"}, config)
			So(err, ShouldBeNil)
			So(decrypted["password"], ShouldEqual, "plaintext_rsa_data")
		})

		Convey("rejects value missing encryption prefix", func() {
			cs := &catalogService{
				cipher: &mockCipher{
					decryptFunc: func(ciphertext string) (string, error) {
						return "", nil
					},
				},
			}
			config := map[string]any{"password": "no_prefix_value"}

			_, err := cs.decryptSensitiveFields([]string{"password"}, config)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not encrypted")
		})

		Convey("returns error when cipher decrypt fails", func() {
			cs := &catalogService{
				cipher: &mockCipher{
					decryptFunc: func(ciphertext string) (string, error) {
						return "", fmt.Errorf("corrupted data")
					},
				},
			}
			config := map[string]any{"password": "ENC:bad_data"}

			_, err := cs.decryptSensitiveFields([]string{"password"}, config)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "password")
		})
	})
}

func TestCatalogService_CheckExistByID(t *testing.T) {
	Convey("Test catalogService.CheckExistByID", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := vmock.NewMockCatalogAccess(ctrl)
		cs := &catalogService{ca: mockCA}

		Convey("returns true when catalog found", func() {
			mockCA.EXPECT().GetByID(gomock.Any(), "test-id").
				Return(&interfaces.Catalog{ID: "test-id"}, nil)

			exists, err := cs.CheckExistByID(context.Background(), "test-id")
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
		})

		Convey("returns false when catalog not found", func() {
			mockCA.EXPECT().GetByID(gomock.Any(), "missing-id").
				Return(nil, nil)

			exists, err := cs.CheckExistByID(context.Background(), "missing-id")
			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
		})

		Convey("propagates access error", func() {
			mockCA.EXPECT().GetByID(gomock.Any(), "test-id").
				Return(nil, fmt.Errorf("db error"))

			_, err := cs.CheckExistByID(context.Background(), "test-id")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestCatalogService_CheckExistByName(t *testing.T) {
	Convey("Test catalogService.CheckExistByName", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := vmock.NewMockCatalogAccess(ctrl)
		cs := &catalogService{ca: mockCA}

		Convey("returns true when catalog with name found", func() {
			mockCA.EXPECT().GetByName(gomock.Any(), "test").
				Return(&interfaces.Catalog{Name: "test"}, nil)

			exists, err := cs.CheckExistByName(context.Background(), "test")
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
		})
	})
}

func TestCatalogService_Create(t *testing.T) {
	Convey("Test catalogService.Create", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := vmock.NewMockCatalogAccess(ctrl)
		mockPS := vmock.NewMockPermissionService(ctrl)
		cs := &catalogService{ca: mockCA, ps: mockPS}

		Convey("defaults missing enabled to disabled with unchecked health", func() {
			mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockCA.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, catalog *interfaces.Catalog) error {
					So(catalog.Enabled, ShouldBeFalse)
					So(catalog.HealthCheckStatus, ShouldEqual, interfaces.CatalogHealthStatusUnchecked)
					return nil
				},
			)
			mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			_, err := cs.Create(context.Background(), &interfaces.CatalogRequest{Name: "catalog"})
			So(err, ShouldBeNil)
		})

		Convey("keeps enabled true with unchecked health", func() {
			mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockCA.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, catalog *interfaces.Catalog) error {
					So(catalog.Enabled, ShouldBeTrue)
					So(catalog.HealthCheckStatus, ShouldEqual, interfaces.CatalogHealthStatusUnchecked)
					return nil
				},
			)
			mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			_, err := cs.Create(context.Background(), &interfaces.CatalogRequest{
				Name:    "catalog",
				Enabled: true,
			})
			So(err, ShouldBeNil)
		})
	})
}

func TestCatalogService_TestConnection(t *testing.T) {
	Convey("Test catalogService.TestConnection", t, func() {
		cs := &catalogService{}

		Convey("rejects nil catalog", func() {
			_, err := cs.TestConnection(context.Background(), nil)
			So(err, ShouldNotBeNil)
		})

		Convey("returns existing health status", func() {
			catalog := &interfaces.Catalog{
				CatalogHealthCheckStatus: interfaces.CatalogHealthCheckStatus{
					HealthCheckStatus: interfaces.CatalogHealthStatusHealthy,
					LastCheckTime:     1234567890,
				},
			}
			result, err := cs.TestConnection(context.Background(), catalog)
			So(err, ShouldBeNil)
			So(result.HealthCheckStatus, ShouldEqual, interfaces.CatalogHealthStatusHealthy)
		})
	})
}

func TestCatalogService_SetEnabled(t *testing.T) {
	Convey("Test catalogService.SetEnabled", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := vmock.NewMockCatalogAccess(ctrl)
		mockPS := vmock.NewMockPermissionService(ctrl)
		cs := &catalogService{ca: mockCA, ps: mockPS}

		Convey("re-enable resets health status to unchecked", func() {
			mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockCA.EXPECT().UpdateEnabled(gomock.Any(), "catalog-1", true, gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, enabled bool, status interfaces.CatalogHealthCheckStatus, _ int64, _ interfaces.AccountInfo) error {
					So(enabled, ShouldBeTrue)
					So(status.HealthCheckStatus, ShouldEqual, interfaces.CatalogHealthStatusUnchecked)
					return nil
				},
			)

			err := cs.SetEnabled(context.Background(), &interfaces.Catalog{
				ID:      "catalog-1",
				Name:    "catalog",
				Enabled: false,
				CatalogHealthCheckStatus: interfaces.CatalogHealthCheckStatus{
					HealthCheckStatus: interfaces.CatalogHealthStatusHealthy,
				},
			}, true)
			So(err, ShouldBeNil)
		})

		Convey("disable preserves existing health status", func() {
			mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockCA.EXPECT().UpdateEnabled(gomock.Any(), "catalog-1", false, gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, enabled bool, status interfaces.CatalogHealthCheckStatus, _ int64, _ interfaces.AccountInfo) error {
					So(enabled, ShouldBeFalse)
					So(status.HealthCheckStatus, ShouldEqual, interfaces.CatalogHealthStatusHealthy)
					return nil
				},
			)

			err := cs.SetEnabled(context.Background(), &interfaces.Catalog{
				ID:      "catalog-1",
				Name:    "catalog",
				Enabled: true,
				CatalogHealthCheckStatus: interfaces.CatalogHealthCheckStatus{
					HealthCheckStatus: interfaces.CatalogHealthStatusHealthy,
				},
			}, false)
			So(err, ShouldBeNil)
		})
	})
}

func TestCatalogService_List(t *testing.T) {
	Convey("Test catalogService.List", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := vmock.NewMockCatalogAccess(ctrl)
		mockPS := vmock.NewMockPermissionService(ctrl)
		mockUMS := vmock.NewMockUserMgmtService(ctrl)
		cs := &catalogService{ca: mockCA, ps: mockPS, ums: mockUMS}

		Convey("returns all when limit is -1", func() {
			ids := []string{"c1", "c2", "c3"}
			catalogs := []*interfaces.Catalog{{ID: "c1"}, {ID: "c2"}, {ID: "c3"}}
			mockCA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(ids, nil)
			mockPS.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true, gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"c1": {ResourceID: "c1"}, "c2": {ResourceID: "c2"}, "c3": {ResourceID: "c3"},
				}, nil)
			mockCA.EXPECT().GetByIDs(gomock.Any(), gomock.Any()).Return(catalogs, nil)
			mockCA.EXPECT().AttachListExtensions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			result, total, err := cs.List(context.Background(), interfaces.CatalogsQueryParams{
				PaginationQueryParams: interfaces.PaginationQueryParams{Limit: -1},
			})
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 3)
			So(result, ShouldHaveLength, 3)
		})

		Convey("applies offset and limit pagination", func() {
			ids := []string{"c1", "c2", "c3", "c4", "c5"}
			mockCA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(ids, nil)
			mockPS.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true, gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"c1": {ResourceID: "c1"}, "c2": {ResourceID: "c2"}, "c3": {ResourceID: "c3"}, "c4": {ResourceID: "c4"}, "c5": {ResourceID: "c5"},
				}, nil)
			catalogs := []*interfaces.Catalog{{ID: "c2"}, {ID: "c3"}}
			mockCA.EXPECT().GetByIDs(gomock.Any(), []string{"c2", "c3"}).Return(catalogs, nil)
			mockCA.EXPECT().AttachListExtensions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			result, total, err := cs.List(context.Background(), interfaces.CatalogsQueryParams{
				PaginationQueryParams: interfaces.PaginationQueryParams{Offset: 1, Limit: 2},
			})
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 5)
			So(result, ShouldHaveLength, 2)
			So(result[0].ID, ShouldEqual, "c2")
		})

		Convey("returns empty when offset beyond total", func() {
			ids := []string{"c1", "c2"}
			mockCA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(ids, nil)
			mockPS.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true, gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"c1": {ResourceID: "c1"}, "c2": {ResourceID: "c2"},
				}, nil)

			result, total, err := cs.List(context.Background(), interfaces.CatalogsQueryParams{
				PaginationQueryParams: interfaces.PaginationQueryParams{Offset: 10, Limit: 5},
			})
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
			So(result, ShouldBeEmpty)
		})

		Convey("filters out catalogs without permission", func() {
			ids := []string{"c1", "c2", "c3"}
			catalogs := []*interfaces.Catalog{{ID: "c1"}, {ID: "c3"}}
			mockCA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(ids, nil)
			// 权限只返回 c1 和 c3，c2 被过滤
			mockPS.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true, gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"c1": {ResourceID: "c1"}, "c3": {ResourceID: "c3"},
				}, nil)
			mockCA.EXPECT().GetByIDs(gomock.Any(), []string{"c1", "c3"}).Return(catalogs, nil)
			mockCA.EXPECT().AttachListExtensions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			result, total, err := cs.List(context.Background(), interfaces.CatalogsQueryParams{
				PaginationQueryParams: interfaces.PaginationQueryParams{Limit: -1},
			})
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
			So(result, ShouldHaveLength, 2)
		})

		Convey("propagates db error from ListIDs", func() {
			mockCA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("db error"))

			_, _, err := cs.List(context.Background(), interfaces.CatalogsQueryParams{})
			So(err, ShouldNotBeNil)
		})
	})
}

func TestCatalogService_DeleteByIDs(t *testing.T) {
	Convey("Test catalogService.DeleteByIDs", t, func() {
		cs := &catalogService{}

		Convey("returns nil on empty input", func() {
			err := cs.DeleteByIDs(context.Background(), []string{})
			So(err, ShouldBeNil)
		})
	})
}
