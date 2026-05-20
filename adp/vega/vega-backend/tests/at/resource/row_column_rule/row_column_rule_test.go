// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package row_column_rule

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend-tests/at/setup"
	"vega-backend-tests/testutil"
)



// ========== 辅助函数 ==========

// generateUniqueName 生成唯一名称
func generateUniqueName(prefix string) string {
	suffix := rand.Intn(10000)
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().Unix(), suffix)
}

// cleanupRowColumnRules 清理现有的行列规则
func cleanupRowColumnRules(client *testutil.HTTPClient, t *testing.T) {
	resp := client.GET("/api/vega-backend/v1/resource-row-column-rules?offset=0&limit=100")
	if resp.StatusCode != http.StatusOK {
		t.Logf("获取行列规则列表失败，状态码: %d", resp.StatusCode)
		return
	}

	if resp.Body == nil {
		return
	}

	rules, ok := resp.Body["rules"].([]any)
	if !ok || len(rules) == 0 {
		return
	}

	t.Logf("清理 %d 个现有行列规则...", len(rules))
	var ruleIDs []string
	for _, rule := range rules {
		if ruleMap, ok := rule.(map[string]any); ok {
			if id, ok := ruleMap["id"].(string); ok {
				ruleIDs = append(ruleIDs, id)
			}
		}
	}

	if len(ruleIDs) > 0 {
		ids := ""
		for i, id := range ruleIDs {
			if i > 0 {
				ids += ","
			}
			ids += id
		}
		deleteResp := client.DELETE("/api/vega-backend/v1/resource-row-column-rules/" + ids)
		if deleteResp.StatusCode != http.StatusOK {
			t.Logf("批量删除行列规则失败，状态码: %d", deleteResp.StatusCode)
		} else {
			t.Log("行列规则清理完成")
		}
	}
}

// cleanupExistingResources 清理现有资源
func cleanupExistingResources(client *testutil.HTTPClient, t *testing.T) {
	resp := client.GET("/api/vega-backend/v1/resources?offset=0&limit=100")
	if resp.StatusCode == http.StatusOK {
		if entries, ok := resp.Body["entries"].([]any); ok {
			for _, entry := range entries {
				if entryMap, ok := entry.(map[string]any); ok {
					if id, ok := entryMap["id"].(string); ok {
						deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + id)
						if deleteResp.StatusCode != http.StatusOK && deleteResp.StatusCode != http.StatusNoContent {
							t.Logf("清理资源失败 %s: %d", id, deleteResp.StatusCode)
						}
					}
				}
			}
		}
	}
}

// cleanupExistingCatalogs 清理现有catalog
func cleanupExistingCatalogs(client *testutil.HTTPClient, t *testing.T) {
	resp := client.GET("/api/vega-backend/v1/catalogs?offset=0&limit=100")
	if resp.StatusCode == http.StatusOK {
		if entries, ok := resp.Body["entries"].([]any); ok {
			for _, entry := range entries {
				if entryMap, ok := entry.(map[string]any); ok {
					if id, ok := entryMap["id"].(string); ok {
						deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + id)
						if deleteResp.StatusCode != http.StatusOK && deleteResp.StatusCode != http.StatusNoContent {
							t.Logf("清理catalog失败 %s: %d", id, deleteResp.StatusCode)
						}
					}
				}
			}
		}
	}
}



// 行列规则名称常量
const (
	rulPrefix     = "rr-test"
	ruleNameValid = "valid-rule-name"
)

// ========== 测试函数 ==========

// TestResourceRowColumnRule 资源行列规则 CRUD AT 测试
// 测试编号前缀: RR1xx (Resource Row Column Rule)
func TestResourceRowColumnRule(t *testing.T) {
	Convey("资源行列规则 AT 测试 - 初始化", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 清理环境
		cleanupRowColumnRules(client, t)
		cleanupExistingResources(client, t)
		cleanupExistingCatalogs(client, t)

		// ========== 准备测试数据：创建 Catalog + Resource ==========
		var catalogID string
		var resourceID string

		Convey("创建测试 Catalog 和 Resource", func() {
			// 1. 创建 catalog
			catalogPayload := map[string]any{
				"name":           generateUniqueName("test-mysql-catalog"),
				"description":    "测试mysql catalog",
				"tags":           []string{"test", "mysql", "catalog"},
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":     "localhost",
					"port":     3330,
					"username": "username",
					"password": "password",
					"database": "test",
				},
			}
			catalogResp := client.POST("/api/vega-backend/v1/catalogs", catalogPayload)
			So(catalogResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(catalogResp.Body["id"], ShouldNotBeEmpty)
			catalogID = catalogResp.Body["id"].(string)

			// 2. 直接创建 resource（不触发 discovery），使用手动定义的 schema_definition
			resourcePayload := map[string]any{
				"catalog_id":        catalogID,
				"name":              generateUniqueName("test-resource-rr"),
				"tags":              []string{"test", "resource"},
				"description":       "测试资源行列规则",
				"category":          "table",
				"status":            "active",
				"database":          "test",
				"source_identifier": "test.resource_rr",
				"schema_definition": []map[string]any{
					{"name": "id", "type": "keyword", "display_name": "ID", "original_name": "id", "description": "唯一标识符"},
					{"name": "name", "type": "keyword", "display_name": "名称", "original_name": "name", "description": "用户名称"},
					{"name": "department", "type": "keyword", "display_name": "部门", "original_name": "department", "description": "部门"},
					{"name": "salary", "type": "long", "display_name": "薪资", "original_name": "salary", "description": "薪资"},
				},
			}
			resourceResp := client.POST("/api/vega-backend/v1/resources", resourcePayload)
			So(resourceResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resourceResp.Body["id"], ShouldNotBeEmpty)
			resourceID = resourceResp.Body["id"].(string)

			t.Logf("测试准备完成: catalogID=%s, resourceID=%s", catalogID, resourceID)
		})

		if catalogID == "" || resourceID == "" {
			return
		}

		// ========== RR101-RR110: 正向测试 - 创建行列规则 ==========

		Convey("RR101: 创建行列规则 - 基本场景(必填字段)", func() {
			ruleName := generateUniqueName(rulPrefix)
			payload := []any{
				map[string]any{
					"name":        ruleName,
					"resource_id": resourceID,
				},
			}
			resp := client.POST("/api/vega-backend/v1/resource-row-column-rules", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(resp.Body, ShouldNotBeEmpty)
			ruleIDs, ok := resp.Body["rule_ids"].([]any)
			So(ok, ShouldBeTrue)
			So(len(ruleIDs), ShouldEqual, 1)
			So(ruleIDs[0].(string), ShouldNotBeEmpty)
		})

		Convey("RR102: 创建行列规则 - 完整字段(含 fields 和 row_filters)", func() {
			ruleName := generateUniqueName(rulPrefix)
			payload := []any{
				map[string]any{
					"name":        ruleName,
					"resource_id": resourceID,
					"fields":      []string{"col1", "col2", "col3"},
					"tags":        []string{"test", "at"},
					"comment":     "测试行列规则完整字段",
					"row_filters": map[string]any{
						"condition": map[string]any{
							"field":    "status",
							"operator": "eq",
							"value":    "active",
						},
					},
				},
			}
			resp := client.POST("/api/vega-backend/v1/resource-row-column-rules", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			ruleIDs, ok := resp.Body["rule_ids"].([]any)
			So(ok, ShouldBeTrue)
			So(len(ruleIDs), ShouldEqual, 1)
			So(ruleIDs[0].(string), ShouldNotBeEmpty)
		})

		Convey("RR103: 批量创建行列规则 - 多条规则", func() {
			ruleName1 := generateUniqueName(rulPrefix)
			ruleName2 := generateUniqueName(rulPrefix)
			payload := []any{
				map[string]any{
					"name":        ruleName1,
					"resource_id": resourceID,
				},
				map[string]any{
					"name":        ruleName2,
					"resource_id": resourceID,
				},
			}
			resp := client.POST("/api/vega-backend/v1/resource-row-column-rules", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			ruleIDs, ok := resp.Body["rule_ids"].([]any)
			So(ok, ShouldBeTrue)
			So(len(ruleIDs), ShouldEqual, 2)
			So(ruleIDs[0].(string), ShouldNotBeEmpty)
			So(ruleIDs[1].(string), ShouldNotBeEmpty)
			So(ruleIDs[0].(string), ShouldNotEqual, ruleIDs[1].(string))
		})

		// 先创建一条规则供后续测试使用
		var createdRuleID string
		Convey("创建测试用的行列规则", func() {
			ruleName := generateUniqueName(rulPrefix)
			payload := []any{
				map[string]any{
					"name":        ruleName,
					"resource_id": resourceID,
					"fields":      []string{"col_a", "col_b"},
					"tags":        []string{"test"},
					"comment":     "测试规则",
				},
			}
			resp := client.POST("/api/vega-backend/v1/resource-row-column-rules", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			ruleIDs, ok := resp.Body["rule_ids"].([]any)
			So(ok, ShouldBeTrue)
			So(len(ruleIDs), ShouldEqual, 1)
			createdRuleID = ruleIDs[0].(string)
			So(createdRuleID, ShouldNotBeEmpty)
		})

		if createdRuleID == "" {
			return
		}

		// ========== RR111-RR120: 获取行列规则 ==========

		Convey("RR111: 获取行列规则 - 通过 ID 查询", func() {
			resp := client.GET("/api/vega-backend/v1/resource-row-column-rules/" + createdRuleID)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			// 获取响应中返回的规则列表
			if rules, ok := resp.Body["rules"].([]any); ok && len(rules) > 0 {
				if rule, ok := rules[0].(map[string]any); ok {
					So(rule["id"], ShouldEqual, createdRuleID)
					So(rule["name"], ShouldNotBeEmpty)
					So(rule["resource_id"], ShouldEqual, resourceID)
				}
			}
		})

		Convey("RR112: 获取行列规则 - 不存在的 ID", func() {
			resp := client.GET("/api/vega-backend/v1/resource-row-column-rules/non-existent-id")
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("RR113: 获取行列规则 - 空 ID", func() {
			resp := client.GET("/api/vega-backend/v1/resource-row-column-rules/")
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		// ========== RR121-RR130: 列表查询 ==========

		Convey("RR121: 列表查询行列规则 - 默认分页", func() {
			resp := client.GET("/api/vega-backend/v1/resource-row-column-rules")
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(resp.Body, ShouldNotBeEmpty)
			_, hasTotal := resp.Body["total"]
			So(hasTotal, ShouldBeTrue)
			rules, hasRules := resp.Body["rules"].([]any)
			So(hasRules, ShouldBeTrue)
			So(len(rules), ShouldBeGreaterThan, 0)
		})

		Convey("RR122: 列表查询行列规则 - 按 resource_id 过滤", func() {
			resp := client.GET("/api/vega-backend/v1/resource-row-column-rules?resource_id=" + resourceID)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			rules, ok := resp.Body["rules"].([]any)
			So(ok, ShouldBeTrue)
			for _, rule := range rules {
				if ruleMap, ok := rule.(map[string]any); ok {
					So(ruleMap["resource_id"], ShouldEqual, resourceID)
				}
			}
		})

		Convey("RR123: 列表查询行列规则 - 分页参数", func() {
			resp := client.GET("/api/vega-backend/v1/resource-row-column-rules?offset=0&limit=5")
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			rules, ok := resp.Body["rules"].([]any)
			So(ok, ShouldBeTrue)
			So(len(rules), ShouldBeLessThanOrEqualTo, 5)
		})

		// ========== RR131-RR140: 更新行列规则 ==========

		Convey("RR131: 更新行列规则 - 更新名称和标签", func() {
			newName := generateUniqueName(rulPrefix + "-updated")
			payload := map[string]any{
				"name": newName,
				"tags": []string{"updated", "test"},
			}
			resp := client.PUT("/api/vega-backend/v1/resource-row-column-rules/"+createdRuleID, payload)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("RR132: 更新行列规则 - 更新 fields", func() {
			payload := map[string]any{
				"fields": []string{"new_col1", "new_col2", "new_col3"},
			}
			resp := client.PUT("/api/vega-backend/v1/resource-row-column-rules/"+createdRuleID, payload)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("RR133: 更新行列规则 - 更新 row_filters", func() {
			payload := map[string]any{
				"row_filters": map[string]any{
					"condition": map[string]any{
						"field":    "department",
						"operator": "eq",
						"value":    "engineering",
					},
				},
			}
			resp := client.PUT("/api/vega-backend/v1/resource-row-column-rules/"+createdRuleID, payload)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("RR134: 更新行列规则 - 不存在的规则 ID", func() {
			payload := map[string]any{
				"name": "should-not-exist",
			}
			resp := client.PUT("/api/vega-backend/v1/resource-row-column-rules/non-existent-id", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		// ========== RR141-RR150: 失败场景 - 创建校验 ==========

		Convey("RR141: 创建行列规则 - 缺少 name", func() {
			payload := []any{
				map[string]any{
					"resource_id": resourceID,
				},
			}
			resp := client.POST("/api/vega-backend/v1/resource-row-column-rules", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("RR142: 创建行列规则 - 缺少 resource_id", func() {
			payload := []any{
				map[string]any{
					"name": generateUniqueName(rulPrefix),
				},
			}
			resp := client.POST("/api/vega-backend/v1/resource-row-column-rules", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("RR143: 创建行列规则 - name 为空字符串", func() {
			payload := []any{
				map[string]any{
					"name":        "",
					"resource_id": resourceID,
				},
			}
			resp := client.POST("/api/vega-backend/v1/resource-row-column-rules", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("RR144: 创建行列规则 - 重复的 rule_id", func() {
			dupID := generateUniqueName("dup-id")
			payload := []any{
				map[string]any{
					"name":        generateUniqueName(rulPrefix),
					"resource_id": resourceID,
					"id":          dupID,
				},
				map[string]any{
					"name":        generateUniqueName(rulPrefix),
					"resource_id": resourceID,
					"id":          dupID,
				},
			}
			resp := client.POST("/api/vega-backend/v1/resource-row-column-rules", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("RR145: 创建行列规则 - 空请求体", func() {
			payload := []any{}
			resp := client.POST("/api/vega-backend/v1/resource-row-column-rules", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			ruleIDs, ok := resp.Body["rule_ids"].([]any)
			So(ok, ShouldBeTrue)
			So(len(ruleIDs), ShouldEqual, 0)
		})

		// ========== RR151-RR160: 删除行列规则 ==========

		Convey("RR151: 删除行列规则 - 单条删除", func() {
			// 创建一条临时规则用于删除
			ruleName := generateUniqueName(rulPrefix + "-del")
			createPayload := []any{
				map[string]any{
					"name":        ruleName,
					"resource_id": resourceID,
				},
			}
			createResp := client.POST("/api/vega-backend/v1/resource-row-column-rules", createPayload)
			So(createResp.StatusCode, ShouldEqual, http.StatusOK)
			ruleIDs, ok := createResp.Body["rule_ids"].([]any)
			So(ok, ShouldBeTrue)
			So(len(ruleIDs), ShouldEqual, 1)
			delRuleID := ruleIDs[0].(string)

			// 删除
			deleteResp := client.DELETE("/api/vega-backend/v1/resource-row-column-rules/" + delRuleID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusOK)

			// 验证已删除
			getResp := client.GET("/api/vega-backend/v1/resource-row-column-rules/" + delRuleID)
			So(getResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("RR152: 删除行列规则 - 批量删除(逗号分隔)", func() {
			// 创建两条临时规则
			ruleName1 := generateUniqueName(rulPrefix + "-batch1")
			ruleName2 := generateUniqueName(rulPrefix + "-batch2")
			createPayload := []any{
				map[string]any{
					"name":        ruleName1,
					"resource_id": resourceID,
				},
				map[string]any{
					"name":        ruleName2,
					"resource_id": resourceID,
				},
			}
			createResp := client.POST("/api/vega-backend/v1/resource-row-column-rules", createPayload)
			So(createResp.StatusCode, ShouldEqual, http.StatusOK)
			ruleIDs, ok := createResp.Body["rule_ids"].([]any)
			So(ok, ShouldBeTrue)
			So(len(ruleIDs), ShouldEqual, 2)

			// 批量删除
			ids := ruleIDs[0].(string) + "," + ruleIDs[1].(string)
			deleteResp := client.DELETE("/api/vega-backend/v1/resource-row-column-rules/" + ids)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusOK)

			// 验证已删除
			getResp := client.GET("/api/vega-backend/v1/resource-row-column-rules/" + ids)
			So(getResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		// ========== 清理测试数据 ==========

		Convey("清理测试数据", func() {
			cleanupRowColumnRules(client, t)
			cleanupExistingResources(client, t)
			cleanupExistingCatalogs(client, t)
			t.Log("测试清理完成")
		})
	})
}
