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
	mock_interfaces "vega-backend/interfaces/mock"

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

// ===== validateAndDecryptSensitiveFields =====

func TestValidateAndDecrypt_NoCipher(t *testing.T) {
	Convey("Test validateAndDecryptSensitiveFields without cipher", t, func() {
		cs := &catalogService{cipher: nil}
		config := map[string]any{"password": "secret123", "host": "localhost"}

		decrypted, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
		So(err, ShouldBeNil)
		So(decrypted["password"], ShouldEqual, "secret123")
		So(config["password"], ShouldEqual, "secret123")
	})
}

func TestValidateAndDecrypt_WithCipher_Success(t *testing.T) {
	Convey("Test validateAndDecryptSensitiveFields with cipher success", t, func() {
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
}

func TestValidateAndDecrypt_WithCipher_DecryptFails(t *testing.T) {
	Convey("Test validateAndDecryptSensitiveFields returns decrypt error", t, func() {
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
}

func TestValidateAndDecrypt_EmptyValue(t *testing.T) {
	Convey("Test validateAndDecryptSensitiveFields skips empty value", t, func() {
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
}

func TestValidateAndDecrypt_NonStringValue(t *testing.T) {
	Convey("Test validateAndDecryptSensitiveFields skips non-string value", t, func() {
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
}

// ===== decryptSensitiveFields =====

func TestDecrypt_NoCipher(t *testing.T) {
	Convey("Test decryptSensitiveFields without cipher", t, func() {
		cs := &catalogService{cipher: nil}
		config := map[string]any{"password": "ENC:ciphertext"}

		decrypted, err := cs.decryptSensitiveFields([]string{"password"}, config)
		So(err, ShouldBeNil)
		So(decrypted["password"], ShouldEqual, "ENC:ciphertext")
	})
}

func TestDecrypt_WithCipher_Success(t *testing.T) {
	Convey("Test decryptSensitiveFields with cipher success", t, func() {
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
}

func TestDecrypt_MissingEncPrefix(t *testing.T) {
	Convey("Test decryptSensitiveFields rejects missing prefix", t, func() {
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
}

func TestDecrypt_DecryptFails(t *testing.T) {
	Convey("Test decryptSensitiveFields returns decrypt error", t, func() {
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
}

// ===== CheckExistByID（使用 mockgen 生成的 mock） =====

func TestCheckExistByID_Found(t *testing.T) {
	Convey("Test CheckExistByID found", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockCA.EXPECT().GetByID(gomock.Any(), "test-id").
			Return(&interfaces.Catalog{ID: "test-id"}, nil)

		cs := &catalogService{ca: mockCA}
		exists, err := cs.CheckExistByID(context.Background(), "test-id")
		So(err, ShouldBeNil)
		So(exists, ShouldBeTrue)
	})
}

func TestCheckExistByID_NotFound(t *testing.T) {
	Convey("Test CheckExistByID not found", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockCA.EXPECT().GetByID(gomock.Any(), "missing-id").
			Return(nil, nil)

		cs := &catalogService{ca: mockCA}
		exists, err := cs.CheckExistByID(context.Background(), "missing-id")
		So(err, ShouldBeNil)
		So(exists, ShouldBeFalse)
	})
}

func TestCheckExistByID_Error(t *testing.T) {
	Convey("Test CheckExistByID error", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockCA.EXPECT().GetByID(gomock.Any(), "test-id").
			Return(nil, fmt.Errorf("db error"))

		cs := &catalogService{ca: mockCA}
		_, err := cs.CheckExistByID(context.Background(), "test-id")
		So(err, ShouldNotBeNil)
	})
}

// ===== CheckExistByName =====

func TestCheckExistByName_Found(t *testing.T) {
	Convey("Test CheckExistByName found", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockCA.EXPECT().GetByName(gomock.Any(), "test").
			Return(&interfaces.Catalog{Name: "test"}, nil)

		cs := &catalogService{ca: mockCA}
		exists, err := cs.CheckExistByName(context.Background(), "test")
		So(err, ShouldBeNil)
		So(exists, ShouldBeTrue)
	})
}

// ===== Create =====

func TestCreate_MissingEnabledDefaultsToDisabledAndUnchecked(t *testing.T) {
	Convey("Test Create defaults missing enabled to disabled and unchecked", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockPS := mock_interfaces.NewMockPermissionService(ctrl)

		mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockCA.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, catalog *interfaces.Catalog) error {
				So(catalog.Enabled, ShouldBeFalse)
				So(catalog.HealthCheckStatus, ShouldEqual, interfaces.CatalogHealthStatusUnchecked)
				return nil
			},
		)
		mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		cs := &catalogService{ca: mockCA, ps: mockPS}
		_, err := cs.Create(context.Background(), &interfaces.CatalogRequest{
			Name: "catalog",
		})
		So(err, ShouldBeNil)
	})
}

func TestCreate_EnabledTrue(t *testing.T) {
	Convey("Test Create keeps enabled true and unchecked", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockPS := mock_interfaces.NewMockPermissionService(ctrl)

		mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockCA.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, catalog *interfaces.Catalog) error {
				So(catalog.Enabled, ShouldBeTrue)
				So(catalog.HealthCheckStatus, ShouldEqual, interfaces.CatalogHealthStatusUnchecked)
				return nil
			},
		)
		mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		cs := &catalogService{ca: mockCA, ps: mockPS}
		_, err := cs.Create(context.Background(), &interfaces.CatalogRequest{
			Name:    "catalog",
			Enabled: true,
		})
		So(err, ShouldBeNil)
	})
}

// ===== TestConnection =====

func TestTestConnection_NilCatalog(t *testing.T) {
	Convey("Test TestConnection rejects nil catalog", t, func() {
		cs := &catalogService{}
		_, err := cs.TestConnection(context.Background(), nil)
		So(err, ShouldNotBeNil)
	})
}

func TestTestConnection_Valid(t *testing.T) {
	Convey("Test TestConnection returns catalog health status", t, func() {
		cs := &catalogService{}
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
}

func TestSetEnabled_ReenableSetsHealthStatusUnchecked(t *testing.T) {
	Convey("Test SetEnabled re-enable sets health status unchecked", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockPS := mock_interfaces.NewMockPermissionService(ctrl)

		mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockCA.EXPECT().UpdateEnabled(gomock.Any(), "catalog-1", true, gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, _ string, enabled bool, status interfaces.CatalogHealthCheckStatus, _ int64, _ interfaces.AccountInfo) error {
				So(enabled, ShouldBeTrue)
				So(status.HealthCheckStatus, ShouldEqual, interfaces.CatalogHealthStatusUnchecked)
				return nil
			},
		)

		cs := &catalogService{ca: mockCA, ps: mockPS}
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
}

func TestSetEnabled_DisablePreservesHealthStatus(t *testing.T) {
	Convey("Test SetEnabled disable preserves health status", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockPS := mock_interfaces.NewMockPermissionService(ctrl)

		mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockCA.EXPECT().UpdateEnabled(gomock.Any(), "catalog-1", false, gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, _ string, enabled bool, status interfaces.CatalogHealthCheckStatus, _ int64, _ interfaces.AccountInfo) error {
				So(enabled, ShouldBeFalse)
				So(status.HealthCheckStatus, ShouldEqual, interfaces.CatalogHealthStatusHealthy)
				return nil
			},
		)

		cs := &catalogService{ca: mockCA, ps: mockPS}
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
}

// ===== List 分页逻辑 =====

func TestList_ReturnAll(t *testing.T) {
	Convey("Test List return all", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockPS := mock_interfaces.NewMockPermissionService(ctrl)
		mockUMS := mock_interfaces.NewMockUserMgmtService(ctrl)

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

		cs := &catalogService{ca: mockCA, ps: mockPS, ums: mockUMS}
		result, total, err := cs.List(context.Background(), interfaces.CatalogsQueryParams{
			PaginationQueryParams: interfaces.PaginationQueryParams{Limit: -1},
		})
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 3)
		So(result, ShouldHaveLength, 3)
	})
}

func TestList_Pagination(t *testing.T) {
	Convey("Test List pagination", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockPS := mock_interfaces.NewMockPermissionService(ctrl)
		mockUMS := mock_interfaces.NewMockUserMgmtService(ctrl)

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

		cs := &catalogService{ca: mockCA, ps: mockPS, ums: mockUMS}
		result, total, err := cs.List(context.Background(), interfaces.CatalogsQueryParams{
			PaginationQueryParams: interfaces.PaginationQueryParams{Offset: 1, Limit: 2},
		})
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 5)
		So(result, ShouldHaveLength, 2)
		So(result[0].ID, ShouldEqual, "c2")
	})
}

func TestList_OffsetBeyondTotal(t *testing.T) {
	Convey("Test List offset beyond total", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockPS := mock_interfaces.NewMockPermissionService(ctrl)

		ids := []string{"c1", "c2"}
		mockCA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(ids, nil)
		mockPS.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true, gomock.Any()).
			Return(map[string]interfaces.PermissionResourceOps{
				"c1": {ResourceID: "c1"}, "c2": {ResourceID: "c2"},
			}, nil)

		cs := &catalogService{ca: mockCA, ps: mockPS}
		result, total, err := cs.List(context.Background(), interfaces.CatalogsQueryParams{
			PaginationQueryParams: interfaces.PaginationQueryParams{Offset: 10, Limit: 5},
		})
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 2)
		So(result, ShouldBeEmpty)
	})
}

func TestList_PermissionFiltersOut(t *testing.T) {
	Convey("Test List permission filters out catalogs", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockPS := mock_interfaces.NewMockPermissionService(ctrl)
		mockUMS := mock_interfaces.NewMockUserMgmtService(ctrl)

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

		cs := &catalogService{ca: mockCA, ps: mockPS, ums: mockUMS}
		result, total, err := cs.List(context.Background(), interfaces.CatalogsQueryParams{
			PaginationQueryParams: interfaces.PaginationQueryParams{Limit: -1},
		})
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 2)
		So(result, ShouldHaveLength, 2)
	})
}

func TestList_DBError(t *testing.T) {
	Convey("Test List db error", t, func() {
		ctrl := gomock.NewController(t)
		mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
		mockCA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("db error"))

		cs := &catalogService{ca: mockCA}
		_, _, err := cs.List(context.Background(), interfaces.CatalogsQueryParams{})
		So(err, ShouldNotBeNil)
	})
}

// ===== DeleteByIDs empty =====

func TestDeleteByIDs_Empty(t *testing.T) {
	Convey("Test DeleteByIDs empty", t, func() {
		cs := &catalogService{}
		err := cs.DeleteByIDs(context.Background(), []string{})
		So(err, ShouldBeNil)
	})
}
