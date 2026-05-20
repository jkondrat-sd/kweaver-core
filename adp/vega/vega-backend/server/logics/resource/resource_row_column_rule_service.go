// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package resource provides Resource management business logic.
package resource

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics"
	"vega-backend/logics/permission"
)

var (
	rcrServiceOnce sync.Once
	rcrService     interfaces.ResourceRowColumnRuleService
)

type resourceRowColumnRuleService struct {
	appSetting *common.AppSetting
	ps         interfaces.PermissionService
	rs         interfaces.ResourceService
	rcra       interfaces.ResourceRowColumnRuleAccess
}

func NewResourceRowColumnRuleService(appSetting *common.AppSetting) interfaces.ResourceRowColumnRuleService {
	rcrServiceOnce.Do(func() {
		rcrService = &resourceRowColumnRuleService{
			appSetting: appSetting,
			ps:         permission.NewPermissionService(appSetting),
			rs:         NewResourceService(appSetting),
			rcra:       logics.RCRA,
		}
	})

	return rcrService
}

// 创建数据视图行列规则
func (rcrs *resourceRowColumnRuleService) CreateResourceRowColumnRules(ctx context.Context, resourceRowColumnRules []*interfaces.ResourceRowColumnRule) ([]string, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "logic layer: Create resource row column rules")
	defer span.End()

	resourceIDs := make([]string, 0, len(resourceRowColumnRules))
	for _, rule := range resourceRowColumnRules {
		resourceIDs = append(resourceIDs, rule.ResourceID)
	}

	// 判断userid对于当前ResourceID是否有行列规则管理的权限（策略决策）
	matchResouces, err := rcrs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE, resourceIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL, interfaces.OPERATION_TYPE_RULE_MANAGE}, false, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return nil, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range resourceIDs {
		if _, exist := matchResouces[mID]; !exist {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for row column rule's manage operation")
		}
	}

	currentTime := time.Now().UnixMilli()
	ruleIDs := make([]string, 0, len(resourceRowColumnRules))
	for _, rule := range resourceRowColumnRules {
		// 校验创建参数
		err = rcrs.validateCreateUpdateParams(ctx, rule)
		if err != nil {
			return nil, err
		}

		// 校验规则名称是否存在
		_, exist, httpErr := rcrs.CheckResourceRowColumnRuleExistByName(ctx, rule.RuleName, rule.ResourceID)
		if httpErr != nil {
			return nil, httpErr
		}
		if exist {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
				verrors.VegaBackend_ResourceRowColumnRule_ExistByName).
				WithErrorDetails(fmt.Sprintf("rule name %s already exist", rule.RuleName))
		}

		// 如果规则ID为空，则生成一个
		if rule.RuleID == "" {
			rule.RuleID = xid.New().String()
		}

		accountInfo := interfaces.AccountInfo{}
		if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
			accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
		}

		rule.Creator = accountInfo
		rule.Updater = accountInfo
		rule.CreateTime = currentTime
		rule.UpdateTime = currentTime

		ruleIDs = append(ruleIDs, rule.RuleID)

	}

	err = rcrs.rcra.CreateResourceRowColumnRules(ctx, resourceRowColumnRules)
	if err != nil {
		logger.Errorf("Create resource row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "create resource row column rules failed")

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return ruleIDs, nil
}

// 删除数据视图行列权限
func (rcrs *resourceRowColumnRuleService) DeleteResourceRowColumnRules(ctx context.Context, ruleIDs []string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "logic layer: Delete resource row column rules")
	defer span.End()

	// 获取视图ID
	rules, err := rcrs.rcra.GetSimpleRulesByRuleIDs(ctx, nil, ruleIDs)
	if err != nil {
		return err
	}
	resourceIDs := make([]string, 0, len(rules))
	for _, rule := range rules {
		resourceIDs = append(resourceIDs, rule.ResourceID)
	}

	// 判断userid对于当前ResourceID是否有行列规则管理的权限（策略决策）
	matchResouces, err := rcrs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE, resourceIDs,
		[]string{interfaces.OPERATION_TYPE_RULE_MANAGE}, false, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range resourceIDs {
		if _, exist := matchResouces[mID]; !exist {
			return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for row column rule's delete operation")
		}
	}

	err = rcrs.rcra.DeleteResourceRowColumnRules(ctx, nil, ruleIDs)
	if err != nil {
		logger.Errorf("Delete resource row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "delete resource row column rules failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	// 清除策略
	err = rcrs.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE_ROW_COLUMN_RULE, ruleIDs)
	if err != nil {
		logger.Errorf("Delete resource row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "delete resource row column rules failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除某个数据视图下的行列权限，内部使用，不校验权限
func (rcrs *resourceRowColumnRuleService) DeleteRowColumnRulesByResourceIDs(ctx context.Context, tx *sql.Tx, resourceIDs []string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "logic layer: Delete resource row column rules")
	defer span.End()

	// 获取视图下所有的行列规则ID
	rules, err := rcrs.rcra.GetSimpleRulesByResourceIDs(ctx, tx, resourceIDs)
	if err != nil {
		logger.Errorf("List resource row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "list resource row column rules failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	ruleIDs := make([]string, 0, len(rules))
	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.RuleID)
	}

	err = rcrs.rcra.DeleteResourceRowColumnRules(ctx, tx, ruleIDs)
	if err != nil {
		logger.Errorf("Delete resource row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "delete resource row column rules failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	// 清除策略
	err = rcrs.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE_ROW_COLUMN_RULE, ruleIDs)
	if err != nil {
		logger.Errorf("Delete resource row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "delete resource row column rules failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 修改数据视图行列权限
func (rcrs *resourceRowColumnRuleService) UpdateResourceRowColumnRule(ctx context.Context, rule *interfaces.ResourceRowColumnRule) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "logic layer: Update resource row column rule")
	defer span.End()

	span.SetAttributes(
		attr.Key("data_view_row_column_rule_id").String(rule.RuleID),
		attr.Key("data_view_row_column_rule_name").String(rule.RuleName),
	)

	// 判断userid是否有修改数据视图行列权限的权限（策略决策）
	err := rcrs.ps.CheckPermission(ctx,
		interfaces.PermissionResource{
			Type: interfaces.RESOURCE_TYPE_RESOURCE,
			ID:   rule.ResourceID,
		},
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL, interfaces.OPERATION_TYPE_RULE_MANAGE},
	)
	if err != nil {
		span.SetStatus(codes.Error, "Check permission failed")
		return err
	}

	oldRules, err := rcrs.GetResourceRowColumnRules(ctx, []string{rule.RuleID})
	if err != nil {
		span.SetStatus(codes.Error, "Get resource row column rules failed")
		return err
	}
	if len(oldRules) == 0 {
		errDetails := fmt.Sprintf("resource row column rule '%s' not exists", rule.RuleID)
		logger.Errorf(errDetails)
		span.SetStatus(codes.Error, errDetails)
		return rest.NewHTTPError(ctx, http.StatusNotFound, rest.PublicError_NotFound).
			WithErrorDetails(errDetails)
	}

	oldRule := oldRules[0]

	oldRuleName := oldRule.RuleName
	oldResourceID := oldRule.ResourceID
	newRuleName := rule.RuleName
	newResourceID := rule.ResourceID

	// 校验更新参数
	err = rcrs.validateCreateUpdateParams(ctx, rule)
	if err != nil {
		return err
	}

	// 视图id不允许变更
	if newResourceID != oldResourceID {
		errDetails := fmt.Sprintf("resource row column rule view id '%s' not allow change, old view id: '%s'", rule.ResourceID, oldResourceID)
		logger.Errorf(errDetails)
		span.SetStatus(codes.Error, errDetails)
		return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails(errDetails)
	}

	// 校验行列规则名称在当前视图下是否已存�?	// if newResourceID != oldResourceID || newRuleName != oldRuleName {
	if newRuleName != oldRuleName {
		_, exist, httpErr := rcrs.CheckResourceRowColumnRuleExistByName(ctx, newRuleName, newResourceID)
		if httpErr != nil {
			span.SetStatus(codes.Error, "Check resource exist by name failed")
			return httpErr
		}

		if exist {
			errDetails := fmt.Sprintf("resource row column rule '%s' already exists in view '%s'", rule.RuleName, rule.ResourceID)
			logger.Errorf(errDetails)
			span.SetStatus(codes.Error, errDetails)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails(errDetails)
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	rule.Updater = accountInfo
	rule.UpdateTime = time.Now().UnixMilli()
	err = rcrs.rcra.UpdateResourceRowColumnRule(ctx, rule)
	if err != nil {
		logger.Errorf("update resource row column rule error, %v", err)
		span.SetStatus(codes.Error, "update resource row column rule failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询数据视图行列权限列表
func (rcrs *resourceRowColumnRuleService) ListResourceRowColumnRules(ctx context.Context,
	params *interfaces.ListRowColumnRuleQueryParams) ([]*interfaces.ResourceRowColumnRule, int, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "logic layer: List resource row column rules")
	defer span.End()

	rules, err := rcrs.rcra.ListResourceRowColumnRules(ctx, params)
	if err != nil {
		logger.Errorf("ListresourceRowColumnRules error: %s", err.Error())

		span.SetStatus(codes.Error, "List resource row column rules failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	if len(rules) == 0 {
		return rules, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// ruleIDs := make([]string, 0)
	resourceIDs := make([]string, 0)
	for _, v := range rules {
		// ruleIDs = append(ruleIDs, v.RuleID)
		resourceIDs = append(resourceIDs, v.ResourceID)
	}

	// matchResouces, err := rcrs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE,
	// 	resourceIDs, []string{interfaces.OPERATION_TYPE_RULE_MANAGE}, true, interfaces.COMMON_OPERATIONS)
	// if err != nil {
	// 	return nil, 0, err
	// }

	// 获取资源操作
	viewOpsMap, err := rcrs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE,
		resourceIDs, []string{interfaces.OPERATION_TYPE_RULE_MANAGE, interfaces.OPERATION_TYPE_RULE_AUTHORIZE},
		false, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return nil, 0, err
	}

	// 决策是否有rule_manage或rule_authorize权限，有其中一个就ok
	checkPermission := func(slice []string, val1, val2 string) bool {
		found1, found2 := false, false
		for _, v := range slice {
			if v == val1 {
				found1 = true
			}
			if v == val2 {
				found2 = true
			}
			// 如果两个都提前找到了，可以立即退出循环
			if found1 && found2 {
				return true
			}
		}

		if !found1 && !found2 {
			return false
		} else {
			return true
		}
	}

	ResourceIDMap := make(map[string]interfaces.PermissionResourceOps)
	for _, ops := range viewOpsMap {
		if checkPermission(ops.Operations, interfaces.OPERATION_TYPE_RULE_MANAGE, interfaces.OPERATION_TYPE_RULE_AUTHORIZE) {
			ResourceIDMap[ops.ResourceID] = ops
		}
	}

	// 根据视图id获取视图名称
	resources, err := rcrs.rs.GetByIDs(ctx, resourceIDs)
	if err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}
	resourceMap := make(map[string]*interfaces.Resource, len(resources))
	for _, r := range resources {
		resourceMap[r.ID] = r
	}

	// 如果包含不存在的，不报错，只打印错误日志
	for _, resourceID := range resourceIDs {
		if _, ok := resourceMap[resourceID]; !ok {
			logger.Errorf("Data view '%s' does not exist!", resourceID)
		}
	}

	// 遍历对象
	results := make([]*interfaces.ResourceRowColumnRule, 0)
	if params.IsInnerRequest {
		results = rules
	} else {
		for _, rl := range rules {
			if _, exist := ResourceIDMap[rl.ResourceID]; exist {
				// 这里返回的是视图的资源类�?				// v.Operations = resrc.Operations
				rl.ResourceName = resourceMap[rl.ResourceID].Name

				results = append(results, rl)
			}
		}
	}

	// limit = -1,则返回所有
	if params.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if params.Offset < 0 || params.Offset >= len(results) {
		return []*interfaces.ResourceRowColumnRule{}, 0, nil
	}
	// 计算结束位置
	end := params.Offset + params.Limit
	if end > len(results) {
		end = len(results)
	}

	span.SetStatus(codes.Ok, "")
	return results[params.Offset:end], len(results), nil
}

// 获取行列规则的资源实例列表
func (rcrs *resourceRowColumnRuleService) ListResourceRowColumnRuleSrcs(ctx context.Context,
	params *interfaces.ListRowColumnRuleQueryParams) ([]interfaces.PermissionResource, int, error) {
	listCtx, listSpan := oteltrace.StartNamedInternalSpan(ctx, "logic layer: List resource row column rule resources")
	listSpan.End()

	rules, err := rcrs.rcra.ListResourceRowColumnRules(listCtx, params)
	if err != nil {
		logger.Errorf("ListresourceRowColumnRules error: %s", err.Error())
		listSpan.SetStatus(codes.Error, "List resource row column rules error")
		listSpan.End()
		return []interfaces.PermissionResource{}, 0, rest.NewHTTPError(listCtx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}
	if len(rules) == 0 {
		return []interfaces.PermissionResource{}, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	ruleIDs := make([]string, 0)
	resourceIDs := make([]string, 0)
	for _, v := range rules {
		ruleIDs = append(ruleIDs, v.RuleID)
		resourceIDs = append(resourceIDs, v.ResourceID)
	}
	// 校验权限管理的操作权限
	matchResouces, err := rcrs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE, ruleIDs,
		[]string{interfaces.OPERATION_TYPE_RULE_MANAGE}, false, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return []interfaces.PermissionResource{}, 0, err
	}

	// 所有有权限的模型数据
	idMap := make(map[string]bool)
	for _, resourceOps := range matchResouces {
		idMap[resourceOps.ResourceID] = true
	}

	// 根据视图id获取视图名称
	resources, err := rcrs.rs.GetByIDs(ctx, resourceIDs)
	if err != nil {
		return []interfaces.PermissionResource{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}
	resourceMap := make(map[string]*interfaces.Resource, len(resources))
	for _, r := range resources {
		resourceMap[r.ID] = r
	}

	// 如果包含不存在的，不报错，只打印错误日志
	for _, rid := range resourceIDs {
		if _, ok := resourceMap[rid]; !ok {
			logger.Errorf("Data view '%s' does not exist!", rid)
		}
	}

	// 遍历对象
	results := make([]interfaces.PermissionResource, 0)
	for _, rule := range rules {
		if idMap[rule.RuleID] {
			if r, ok := resourceMap[rule.ResourceID]; ok {
				rule.ResourceName = r.Name
			}

			results = append(results, interfaces.PermissionResource{
				ID:   rule.RuleID,
				Type: interfaces.RESOURCE_TYPE_RESOURCE,
				Name: rule.ResourceName,
			})
		}
	}

	// 分页
	// 检查起始位置是否越界
	if params.Offset < 0 || params.Offset >= len(results) {
		return []interfaces.PermissionResource{}, 0, nil
	}
	// 计算结束位置
	end := params.Offset + params.Limit
	if end > len(results) {
		end = len(results)
	}

	listSpan.SetStatus(codes.Ok, "")
	return results[params.Offset:end], len(results), nil
}

// 按ID获取数据视图行列权限信息
func (rcrs *resourceRowColumnRuleService) GetResourceRowColumnRules(ctx context.Context, ruleIDs []string) ([]*interfaces.ResourceRowColumnRule, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, fmt.Sprintf("logic layer: Get resource row column rules '%s' info", strings.Join(ruleIDs, ",")))
	defer span.End()

	span.SetAttributes(attr.Key("rule_ids").String(strings.Join(ruleIDs, ",")))

	rules, err := rcrs.rcra.GetResourceRowColumnRules(ctx, ruleIDs)
	if err != nil {
		logger.Errorf("Get resource row column rule by id error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get resource row column rules '%s' failed", strings.Join(ruleIDs, ",")))
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 找到不存在的视图 id，如果有视图 id 不存在，则返回错误
	if len(rules) < len(ruleIDs) {
		ruleMap := make(map[string]struct{}, len(rules))
		for _, r := range rules {
			ruleMap[r.RuleID] = struct{}{}
		}

		for _, ruleID := range ruleIDs {
			if _, ok := ruleMap[ruleID]; !ok {
				errDetails := fmt.Sprintf("The resource row column rule %s was not found", ruleID)
				logger.Errorf(errDetails)

				otellog.LogError(ctx, errDetails, nil)
				span.SetStatus(codes.Error, errDetails)
				return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound).
					WithErrorDetails(errDetails)
			}
		}
	}

	// 先获取资源序列
	matchResouces, err := rcrs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE, ruleIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return nil, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range ruleIDs {
		if _, exist := matchResouces[mID]; !exist {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for resource row column rule's view_detail operation.")
		}
	}

	for index, r := range rules {
		// 补充视图的可操作权限、数据源名称
		r.Operations = matchResouces[r.RuleID].Operations

		rules[index] = r
	}

	span.SetStatus(codes.Ok, "")
	return rules, nil
}

// 根据id检查行列规则是否存在
func (rcrs *resourceRowColumnRuleService) CheckResourceRowColumnRuleExistByID(ctx context.Context, ruleID string) (string, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, fmt.Sprintf("logic layer: Check resource row column rule '%s' existence", ruleID))
	defer span.End()

	span.SetAttributes(attr.Key("rule_id").String(ruleID))

	ruleID, exist, err := rcrs.rcra.CheckResourceRowColumnRuleExistByID(ctx, ruleID)
	if err != nil {
		logger.Errorf("Check resource row column rule existence by id error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("failed to check resource row column rule '%s' existence", ruleID))
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 校验规则是否存在
	if !exist {
		errDetails := fmt.Sprintf("The resource row column rule %s was not found", ruleID)
		logger.Errorf(errDetails)

		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, errDetails)
		return "", rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound).
			WithErrorDetails(errDetails)
	}

	span.SetStatus(codes.Ok, "")
	return ruleID, nil
}

// 根据名称检查行列规则是否存在
func (rcrs *resourceRowColumnRuleService) CheckResourceRowColumnRuleExistByName(ctx context.Context, ruleName string, ResourceID string) (string, bool, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, fmt.Sprintf("logic layer: Check resource row column rule '%s' existence", ruleName))
	defer span.End()

	span.SetAttributes(attr.Key("rule_name").String(ruleName))

	ruleID, exist, err := rcrs.rcra.CheckResourceRowColumnRuleExistByName(ctx, ruleName, ResourceID)
	if err != nil {
		logger.Errorf("Check resource row column rule existence by name error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("failed to check resource row column rule '%s' existence", ruleName))
		return ruleID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return ruleID, exist, nil
}

// 校验创建和更新的参数
func (rcrs *resourceRowColumnRuleService) validateCreateUpdateParams(ctx context.Context, rule *interfaces.ResourceRowColumnRule) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "logic layer: Validate create/update resource row column rule params")
	defer span.End()

	// TODO: 校验视图是否存在（暂未实现）

	return nil
}
