// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	"vega-backend/common/visitor"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// ========== CreateResourceRowColumnRules ==========

// CreateResourceRowColumnRulesByEx handles POST /api/vega-backend/v1/resource-row-column-rules (External)
func (r *restHandler) CreateResourceRowColumnRulesByEx(c *gin.Context) {
	logger.Debug("Handler CreateResourceRowColumnRulesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateResourceRowColumnRules(c, visitor)
}

// CreateResourceRowColumnRulesByIn handles POST /api/vega-backend/in/v1/resource-row-column-rules (Internal)
func (r *restHandler) CreateResourceRowColumnRulesByIn(c *gin.Context) {
	logger.Debug("Handler CreateResourceRowColumnRulesByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.CreateResourceRowColumnRules(c, visitor)
}

// CreateResourceRowColumnRules creates resource row column rules
func (r *restHandler) CreateResourceRowColumnRules(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler CreateResourceRowColumnRules Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	var reqBody []interfaces.ResourceRowColumnRule
	err := c.ShouldBindJSON(&reqBody)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("Binding parameter failed: " + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			generateResourceRowColumnRuleAuditObject("", ""), &httpErr.BaseError)

		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Validate request body
	nameMap := make(map[string]any)
	idMap := make(map[string]any)
	rules := make([]*interfaces.ResourceRowColumnRule, 0, len(reqBody))
	for i := 0; i < len(reqBody); i++ {
		ruleID := reqBody[i].RuleID
		ruleName := reqBody[i].RuleName
		resourceID := reqBody[i].ResourceID
		uk_rule_name := fmt.Sprintf("%s_%s", ruleName, resourceID)

		// Check duplicate rule IDs in request body
		if ruleID != "" {
			if _, ok := idMap[ruleID]; !ok {
				idMap[ruleID] = nil
			} else {
				errDetails := fmt.Sprintf("resource row column rule ID '%s' already exists in the request body", ruleID)
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails(errDetails)

				audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
					generateResourceRowColumnRuleAuditObject(ruleID, ruleName), &httpErr.BaseError)

				oteltrace.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}

		// Check duplicate rule names within the resource
		if _, ok := nameMap[uk_rule_name]; !ok {
			nameMap[uk_rule_name] = nil
		} else {
			errDetails := fmt.Sprintf("resource row column rule name '%s' already exists within the resource '%s' in the request body", ruleName, resourceID)
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails(errDetails)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				generateResourceRowColumnRuleAuditObject(ruleID, ruleName), &httpErr.BaseError)

			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		rule := &interfaces.ResourceRowColumnRule{
			RuleID:     ruleID,
			RuleName:   ruleName,
			ResourceID: resourceID,
			Tags:       reqBody[i].Tags,
			Comment:    reqBody[i].Comment,
			Fields:     reqBody[i].Fields,
			RowFilters: reqBody[i].RowFilters,
		}

		// Validate rule parameters
		err = validateResourceRowColumnRule(ctx, rule)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				generateResourceRowColumnRuleAuditObject(ruleID, ruleName), &httpErr.BaseError)

			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		rules = append(rules, rule)
	}

	// Batch create
	ruleIDs, err := r.rcrs.CreateResourceRowColumnRules(ctx, rules)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			generateResourceRowColumnRuleAuditObject("", ""), &httpErr.BaseError)

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		generateResourceRowColumnRuleAuditObject("", ""), "")

	span.SetStatus(codes.Ok, "")
	rest.ReplyOK(c, http.StatusOK, gin.H{"rule_ids": ruleIDs})
}

// ========== DeleteResourceRowColumnRules ==========

// DeleteResourceRowColumnRulesByEx handles DELETE /api/vega-backend/v1/resource-row-column-rules/:rule_ids (External)
func (r *restHandler) DeleteResourceRowColumnRulesByEx(c *gin.Context) {
	logger.Debug("Handler DeleteResourceRowColumnRulesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.DeleteResourceRowColumnRules(c, visitor)
}

// DeleteResourceRowColumnRulesByIn handles DELETE /api/vega-backend/in/v1/resource-row-column-rules/:rule_ids (Internal)
func (r *restHandler) DeleteResourceRowColumnRulesByIn(c *gin.Context) {
	logger.Debug("Handler DeleteResourceRowColumnRulesByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.DeleteResourceRowColumnRules(c, visitor)
}

// DeleteResourceRowColumnRules deletes resource row column rules
func (r *restHandler) DeleteResourceRowColumnRules(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler DeleteResourceRowColumnRules Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	ruleIDsStr := c.Param("rule_ids")
	ruleIDs := strings.Split(ruleIDsStr, ",")

	if len(ruleIDs) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("rule_ids is required")

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			generateResourceRowColumnRuleAuditObject("", ""), &httpErr.BaseError)

		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	span.SetAttributes(attr.Key("rule_ids").String(fmt.Sprintf("%v", ruleIDs)))

	// Check if rules exist
	ruleIDNameMap := make(map[string]string)
	for _, ruleID := range ruleIDs {
		ruleName, err := r.rcrs.CheckResourceRowColumnRuleExistByID(ctx, ruleID)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				generateResourceRowColumnRuleAuditObject(ruleID, ruleName), &httpErr.BaseError)

			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		ruleIDNameMap[ruleID] = ruleName
	}

	// Delete
	err := r.rcrs.DeleteResourceRowColumnRules(ctx, ruleIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			generateResourceRowColumnRuleAuditObject("", ""), &httpErr.BaseError)

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
		generateResourceRowColumnRuleAuditObject("", ""), "")

	span.SetStatus(codes.Ok, "")
	rest.ReplyOK(c, http.StatusOK, nil)
}

// ========== GetResourceRowColumnRules ==========

// GetResourceRowColumnRulesByEx handles GET /api/vega-backend/v1/resource-row-column-rules/:rule_id (External)
func (r *restHandler) GetResourceRowColumnRulesByEx(c *gin.Context) {
	logger.Debug("Handler GetResourceRowColumnRulesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetResourceRowColumnRules(c, visitor)
}

// GetResourceRowColumnRulesByIn handles GET /api/vega-backend/in/v1/resource-row-column-rules/:rule_id (Internal)
func (r *restHandler) GetResourceRowColumnRulesByIn(c *gin.Context) {
	logger.Debug("Handler GetResourceRowColumnRulesByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.GetResourceRowColumnRules(c, visitor)
}

// GetResourceRowColumnRules gets resource row column rules
func (r *restHandler) GetResourceRowColumnRules(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetResourceRowColumnRules Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	ruleID := c.Param("rule_id")
	if ruleID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("rule_id is required")

		audit.NewWarnLogWithError(audit.OPERATION, "read", audit.TransforOperator(visitor),
			generateResourceRowColumnRuleAuditObject("", ""), &httpErr.BaseError)

		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	span.SetAttributes(attr.Key("rule_id").String(ruleID))

	rule, err := r.rcrs.GetResourceRowColumnRules(ctx, []string{ruleID})
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, "read", audit.TransforOperator(visitor),
			generateResourceRowColumnRuleAuditObject(ruleID, ""), &httpErr.BaseError)

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if len(rule) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound).
			WithErrorDetails(fmt.Sprintf("Resource row column rule '%s' not found", ruleID))

		audit.NewWarnLogWithError(audit.OPERATION, "read", audit.TransforOperator(visitor),
			generateResourceRowColumnRuleAuditObject(ruleID, ""), &httpErr.BaseError)

		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, "read", audit.TransforOperator(visitor),
		generateResourceRowColumnRuleAuditObject(ruleID, rule[0].RuleName), "")

	span.SetStatus(codes.Ok, "")
	rest.ReplyOK(c, http.StatusOK, gin.H{"rule": rule[0]})
}

// ========== ListResourceRowColumnRules ==========

// ListResourceRowColumnRulesByEx handles GET /api/vega-backend/v1/resource-row-column-rules (External)
func (r *restHandler) ListResourceRowColumnRulesByEx(c *gin.Context) {
	logger.Debug("Handler ListResourceRowColumnRulesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListResourceRowColumnRules(c, visitor)
}

// ListResourceRowColumnRulesByIn handles GET /api/vega-backend/in/v1/resource-row-column-rules (Internal)
func (r *restHandler) ListResourceRowColumnRulesByIn(c *gin.Context) {
	logger.Debug("Handler ListResourceRowColumnRulesByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.ListResourceRowColumnRules(c, visitor)
}

// ListResourceRowColumnRules lists resource row column rules
func (r *restHandler) ListResourceRowColumnRules(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler ListResourceRowColumnRules Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	resourceID := c.Query("resource_id")
	name := c.Query("name")
	namePattern := c.Query("name_pattern")
	tag := c.Query("tag")
	offsetStr := common.GetQueryOrDefault(c, "offset", interfaces.DEFAULT_OFFSET)
	limitStr := common.GetQueryOrDefault(c, "limit", interfaces.DEFAULT_LIMIT)
	sort := common.GetQueryOrDefault(c, "sort", "update_time")
	direction := common.GetQueryOrDefault(c, "direction", interfaces.DESC_DIRECTION)

	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	params := &interfaces.ListRowColumnRuleQueryParams{
		ResourceID:  resourceID,
		Name:        name,
		NamePattern: namePattern,
		Tag:         tag,
		IsInnerRequest: false,
		PaginationQueryParams: interfaces.PaginationQueryParams{
			Offset:    offset,
			Limit:     limit,
			Sort:      sort,
			Direction: direction,
		},
	}

	rules, total, err := r.rcrs.ListResourceRowColumnRules(ctx, params)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, "read", audit.TransforOperator(visitor),
		generateResourceRowColumnRuleAuditObject("", ""), "")

	span.SetStatus(codes.Ok, "")
	rest.ReplyOK(c, http.StatusOK, gin.H{"total": total, "rules": rules})
}

// ========== UpdateResourceRowColumnRules ==========

// UpdateResourceRowColumnRulesByEx handles PUT /api/vega-backend/v1/resource-row-column-rules/:rule_id (External)
func (r *restHandler) UpdateResourceRowColumnRulesByEx(c *gin.Context) {
	logger.Debug("Handler UpdateResourceRowColumnRulesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateResourceRowColumnRules(c, visitor)
}

// UpdateResourceRowColumnRulesByIn handles PUT /api/vega-backend/in/v1/resource-row-column-rules/:rule_id (Internal)
func (r *restHandler) UpdateResourceRowColumnRulesByIn(c *gin.Context) {
	logger.Debug("Handler UpdateResourceRowColumnRulesByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.UpdateResourceRowColumnRules(c, visitor)
}

// UpdateResourceRowColumnRules updates resource row column rules
func (r *restHandler) UpdateResourceRowColumnRules(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler UpdateResourceRowColumnRules Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	ruleID := c.Param("rule_id")
	if ruleID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("rule_id is required")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			generateResourceRowColumnRuleAuditObject("", ""), &httpErr.BaseError)

		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	var reqBody interfaces.ResourceRowColumnRule
	err := c.ShouldBindJSON(&reqBody)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("Binding parameter failed: " + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			generateResourceRowColumnRuleAuditObject(ruleID, ""), &httpErr.BaseError)

		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	reqBody.RuleID = ruleID

	// Validate update parameters
	err = validateUpdateResourceRowColumnRule(ctx, &reqBody)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			generateResourceRowColumnRuleAuditObject(ruleID, reqBody.RuleName), &httpErr.BaseError)

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	err = r.rcrs.UpdateResourceRowColumnRule(ctx, &reqBody)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			generateResourceRowColumnRuleAuditObject(ruleID, reqBody.RuleName), &httpErr.BaseError)

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		generateResourceRowColumnRuleAuditObject(ruleID, reqBody.RuleName), "")

	span.SetStatus(codes.Ok, "")
	rest.ReplyOK(c, http.StatusOK, nil)
}

// ========== Helper Functions ==========

// generateResourceRowColumnRuleAuditObject generates audit object for resource row column rule
func generateResourceRowColumnRuleAuditObject(ruleID string, ruleName string) audit.AuditObject {
	return audit.AuditObject{
		Type: "resource_row_column_rule",
		ID:   ruleID,
		Name: ruleName,
	}
}

// validateResourceRowColumnRule validates resource row column rule creation parameters
func validateResourceRowColumnRule(ctx context.Context, rule *interfaces.ResourceRowColumnRule) error {
	// Validate rule name
	if rule.RuleName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RuleName).
			WithErrorDetails("rule_name is required")
	}

	if len(rule.RuleName) > interfaces.NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RuleName).
			WithErrorDetails(fmt.Sprintf("rule_name length must not exceed %d characters", interfaces.NAME_MAX_LENGTH))
	}

	// Validate resource ID
	if rule.ResourceID == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_ResourceID).
			WithErrorDetails("resource_id is required")
	}

	// Validate rule ID pattern if provided
	if rule.RuleID != "" {
		if matched, _ := regexp.MatchString(interfaces.RegexPattern_NonBuiltin_ID, rule.RuleID); !matched {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails(fmt.Sprintf("rule_id must match pattern: %s", interfaces.RegexPattern_NonBuiltin_ID))
		}
	}

	// Validate tags
	for _, tag := range rule.Tags {
		if len(tag) > interfaces.TAG_MAX_LENGTH {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Tags).
				WithErrorDetails(fmt.Sprintf("tag length must not exceed %d characters", interfaces.TAG_MAX_LENGTH))
		}
		if strings.ContainsAny(tag, interfaces.TAG_INVALID_CHARACTER) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Tags).
				WithErrorDetails(fmt.Sprintf("tag contains invalid characters: %s", interfaces.TAG_INVALID_CHARACTER))
		}
	}

	if len(rule.Tags) > interfaces.TAGS_MAX_NUMBER {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Tags).
			WithErrorDetails(fmt.Sprintf("number of tags must not exceed %d", interfaces.TAGS_MAX_NUMBER))
	}

	return nil
}

// validateUpdateResourceRowColumnRule validates resource row column rule update parameters
func validateUpdateResourceRowColumnRule(ctx context.Context, rule *interfaces.ResourceRowColumnRule) error {
	// Validate rule name if provided
	if rule.RuleName != "" {
		if len(rule.RuleName) > interfaces.NAME_MAX_LENGTH {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RuleName).
				WithErrorDetails(fmt.Sprintf("rule_name length must not exceed %d characters", interfaces.NAME_MAX_LENGTH))
		}
	}

	// Validate tags if provided
	if rule.Tags != nil {
		for _, tag := range rule.Tags {
			if len(tag) > interfaces.TAG_MAX_LENGTH {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Tags).
					WithErrorDetails(fmt.Sprintf("tag length must not exceed %d characters", interfaces.TAG_MAX_LENGTH))
			}
			if strings.ContainsAny(tag, interfaces.TAG_INVALID_CHARACTER) {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Tags).
					WithErrorDetails(fmt.Sprintf("tag contains invalid characters: %s", interfaces.TAG_INVALID_CHARACTER))
			}
		}

		if len(rule.Tags) > interfaces.TAGS_MAX_NUMBER {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Tags).
				WithErrorDetails(fmt.Sprintf("number of tags must not exceed %d", interfaces.TAGS_MAX_NUMBER))
		}
	}

	return nil
}
