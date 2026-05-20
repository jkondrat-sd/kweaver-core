// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

// ResourceRowColumnRule 资源行列规则结构体
type ResourceRowColumnRule struct {
	RuleID     string         `json:"id"`
	RuleName   string         `json:"name"`
	ResourceID string         `json:"resource_id"`
	ResourceName string       `json:"resource_name,omitempty"`
	Tags       []string       `json:"tags"`
	Comment    string         `json:"comment"`
	CreateTime int64          `json:"create_time"`
	UpdateTime int64          `json:"update_time"`
	Creator    AccountInfo    `json:"creator"`
	Updater    AccountInfo    `json:"updater"`
	Fields     []string       `json:"fields"`
	RowFilters *FilterCondCfg `json:"row_filters"`

	// 操作权限
	Operations []string `json:"operations,omitempty"`
}

type ListRowColumnRuleQueryParams struct {
	Name           string
	NamePattern    string
	ResourceID     string
	Tag            string
	IsInnerRequest bool
	PaginationQueryParams
}

//go:generate mockgen -source ../interfaces/resource_row_column_rule.go -destination ../interfaces/mock/mock_resource_row_column_rule.go
type ResourceRowColumnRuleAccess interface {
	CreateResourceRowColumnRules(ctx context.Context, rules []*ResourceRowColumnRule) error
	UpdateResourceRowColumnRule(ctx context.Context, rule *ResourceRowColumnRule) error
	GetResourceRowColumnRules(ctx context.Context, ruleIDs []string) ([]*ResourceRowColumnRule, error)
	ListResourceRowColumnRules(ctx context.Context, query *ListRowColumnRuleQueryParams) ([]*ResourceRowColumnRule, error)
	DeleteResourceRowColumnRules(ctx context.Context, tx *sql.Tx, ruleIDs []string) error

	CheckResourceRowColumnRuleExistByID(ctx context.Context, ruleID string) (string, bool, error)
	CheckResourceRowColumnRuleExistByName(ctx context.Context, ruleName, resourceID string) (string, bool, error)
	GetSimpleRulesByResourceIDs(ctx context.Context, tx *sql.Tx, resourceIDs []string) ([]*ResourceRowColumnRule, error)
	GetSimpleRulesByRuleIDs(ctx context.Context, tx *sql.Tx, ruleIDs []string) ([]*ResourceRowColumnRule, error)
}

//go:generate mockgen -source ../interfaces/resource_row_column_rule.go -destination ../interfaces/mock/mock_resource_row_column_rule_service.go
type ResourceRowColumnRuleService interface {
	CreateResourceRowColumnRules(ctx context.Context, rules []*ResourceRowColumnRule) ([]string, error)
	UpdateResourceRowColumnRule(ctx context.Context, rule *ResourceRowColumnRule) error
	GetResourceRowColumnRules(ctx context.Context, ruleIDs []string) ([]*ResourceRowColumnRule, error)
	ListResourceRowColumnRules(ctx context.Context, query *ListRowColumnRuleQueryParams) ([]*ResourceRowColumnRule, int, error)
	DeleteResourceRowColumnRules(ctx context.Context, ruleIDs []string) error

	// 校验资源行列规则是否存在
	CheckResourceRowColumnRuleExistByID(ctx context.Context, ruleID string) (string, error)

	// 获取行列规则的资源实例列表
	ListResourceRowColumnRuleSrcs(ctx context.Context, params *ListRowColumnRuleQueryParams) ([]PermissionResource, int, error)
	DeleteRowColumnRulesByResourceIDs(ctx context.Context, tx *sql.Tx, resourceIDs []string) error
}


