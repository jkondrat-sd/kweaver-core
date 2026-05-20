// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package resource provides Resource row column rule data access operations.
package resource

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	"vega-backend/interfaces"
)

const (
	RESOURCE_ROW_COLUMN_RULE_TABLE_NAME = "t_resource_row_column_rule"
)

var (
	rcrAccessOnce sync.Once
	rcrAccess     interfaces.ResourceRowColumnRuleAccess
)

type resourceRowColumnRuleAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

// NewResourceRowColumnRuleAccess creates a new ResourceRowColumnRuleAccess.
func NewResourceRowColumnRuleAccess(appSetting *common.AppSetting) interfaces.ResourceRowColumnRuleAccess {
	rcrAccessOnce.Do(func() {
		rcrAccess = &resourceRowColumnRuleAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return rcrAccess
}

// CreateResourceRowColumnRules creates resource row column rules.
func (rcra *resourceRowColumnRuleAccess) CreateResourceRowColumnRules(ctx context.Context, rules []*interfaces.ResourceRowColumnRule) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "driven layer: Insert resource row column rules into DB")
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	builder := sq.Insert(RESOURCE_ROW_COLUMN_RULE_TABLE_NAME).
		Columns(
			"f_rule_id",
			"f_rule_name",
			"f_resource_id",
			"f_tags",
			"f_comment",
			"f_fields",
			"f_row_filters",
			"f_create_time",
			"f_update_time",
			"f_creator",
			"f_creator_type",
			"f_updater",
			"f_updater_type",
		)

	for _, rule := range rules {
		tagsStr := libCommon.TagSlice2TagString(rule.Tags)

		fieldsBytes, err := sonic.Marshal(rule.Fields)
		if err != nil {
			errDetails := fmt.Sprintf("Marshal fields failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Marshal fields failed")
			return err
		}

		rowFiltersBytes, err := sonic.Marshal(rule.RowFilters)
		if err != nil {
			errDetails := fmt.Sprintf("Marshal row filters failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Marshal row filters failed")
			return err
		}

		builder = builder.Values(
			rule.RuleID,
			rule.RuleName,
			rule.ResourceID,
			tagsStr,
			rule.Comment,
			string(fieldsBytes),
			string(rowFiltersBytes),
			rule.CreateTime,
			rule.UpdateTime,
			rule.Creator.ID,
			rule.Creator.Type,
			rule.Updater.ID,
			rule.Updater.Type,
		)
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Build SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Build SQL failed")
		return err
	}

	logger.Debugf("SQL: %s, Vals: %v", sqlStr, vals)

	_, err = rcra.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		errDetails := fmt.Sprintf("Exec SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Exec SQL failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateResourceRowColumnRule updates a resource row column rule.
func (rcra *resourceRowColumnRuleAccess) UpdateResourceRowColumnRule(ctx context.Context, rule *interfaces.ResourceRowColumnRule) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "driven layer: Update resource row column rule in DB")
	defer span.End()

	tagsStr := libCommon.TagSlice2TagString(rule.Tags)

	fieldsBytes, err := sonic.Marshal(rule.Fields)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal fields failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Marshal fields failed")
		return err
	}

	rowFiltersBytes, err := sonic.Marshal(rule.RowFilters)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal row filters failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Marshal row filters failed")
		return err
	}

	sqlStr, vals, err := sq.Update(RESOURCE_ROW_COLUMN_RULE_TABLE_NAME).
		Set("f_rule_name", rule.RuleName).
		Set("f_tags", tagsStr).
		Set("f_comment", rule.Comment).
		Set("f_fields", string(fieldsBytes)).
		Set("f_row_filters", string(rowFiltersBytes)).
		Set("f_update_time", rule.UpdateTime).
		Set("f_updater", rule.Updater.ID).
		Set("f_updater_type", rule.Updater.Type).
		Where(sq.Eq{"f_rule_id": rule.RuleID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Build SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Build SQL failed")
		return err
	}

	logger.Debugf("SQL: %s, Vals: %v", sqlStr, vals)

	result, err := rcra.db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		errDetails := fmt.Sprintf("Exec SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Exec SQL failed")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errDetails := fmt.Sprintf("Get rows affected failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Get rows affected failed")
		return err
	}

	if rowsAffected == 0 {
		errDetails := fmt.Sprintf("Resource row column rule '%s' not found", rule.RuleID)
		logger.Error(errDetails)
		span.SetStatus(codes.Error, errDetails)
		return rest.NewHTTPError(ctx, http.StatusNotFound, rest.PublicError_NotFound).
			WithErrorDetails(errDetails)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetResourceRowColumnRules gets resource row column rules by IDs.
func (rcra *resourceRowColumnRuleAccess) GetResourceRowColumnRules(ctx context.Context, ruleIDs []string) ([]*interfaces.ResourceRowColumnRule, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "driven layer: Get resource row column rules from DB")
	defer span.End()

	if len(ruleIDs) == 0 {
		return []*interfaces.ResourceRowColumnRule{}, nil
	}

	sqlStr, vals, err := sq.Select(
		"f_rule_id",
		"f_rule_name",
		"f_resource_id",
		"f_tags",
		"f_comment",
		"f_fields",
		"f_row_filters",
		"f_create_time",
		"f_update_time",
		"f_creator",
		"f_creator_type",
		"f_updater",
		"f_updater_type",
	).
		From(RESOURCE_ROW_COLUMN_RULE_TABLE_NAME).
		Where(sq.Eq{"f_rule_id": ruleIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Build SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Build SQL failed")
		return nil, err
	}

	logger.Debugf("SQL: %s, Vals: %v", sqlStr, vals)

	rows, err := rcra.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		errDetails := fmt.Sprintf("Query SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Query SQL failed")
		return nil, err
	}
	defer rows.Close()

	rules := make([]*interfaces.ResourceRowColumnRule, 0)
	for rows.Next() {
		var rule interfaces.ResourceRowColumnRule
		var tagsStr string
		var fieldsStr string
		var rowFiltersStr string
		var creatorID, creatorType, updaterID, updaterType string

		err := rows.Scan(
			&rule.RuleID,
			&rule.RuleName,
			&rule.ResourceID,
			&tagsStr,
			&rule.Comment,
			&fieldsStr,
			&rowFiltersStr,
			&rule.CreateTime,
			&rule.UpdateTime,
			&creatorID,
			&creatorType,
			&updaterID,
			&updaterType,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Scan row failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, err
		}

		rule.Tags = libCommon.TagString2TagSlice(tagsStr)

		err = sonic.UnmarshalString(fieldsStr, &rule.Fields)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal fields failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Unmarshal fields failed")
			return nil, err
		}

		err = sonic.UnmarshalString(rowFiltersStr, &rule.RowFilters)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal row filters failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Unmarshal row filters failed")
			return nil, err
		}

		rule.Creator = interfaces.AccountInfo{
			ID:   creatorID,
			Type: creatorType,
		}
		rule.Updater = interfaces.AccountInfo{
			ID:   updaterID,
			Type: updaterType,
		}

		rules = append(rules, &rule)
	}

	span.SetStatus(codes.Ok, "")
	return rules, nil
}

// ListResourceRowColumnRules lists resource row column rules with query params.
func (rcra *resourceRowColumnRuleAccess) ListResourceRowColumnRules(ctx context.Context, query *interfaces.ListRowColumnRuleQueryParams) ([]*interfaces.ResourceRowColumnRule, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "driven layer: List resource row column rules from DB")
	defer span.End()

	builder := sq.Select(
		"f_rule_id",
		"f_rule_name",
		"f_resource_id",
		"f_tags",
		"f_comment",
		"f_fields",
		"f_row_filters",
		"f_create_time",
		"f_update_time",
		"f_creator",
		"f_creator_type",
		"f_updater",
		"f_updater_type",
	).
		From(RESOURCE_ROW_COLUMN_RULE_TABLE_NAME)

	// Apply filters
	if query.NamePattern != "" {
		builder = builder.Where(sq.Like{"f_rule_name": query.NamePattern})
	}
	if query.ResourceID != "" {
		builder = builder.Where(sq.Eq{"f_resource_id": query.ResourceID})
	}
	if query.Tag != "" {
		builder = builder.Where(sq.Like{"f_tags": "%" + query.Tag + "%"})
	}

	// Apply pagination
	if query.Limit > 0 {
		builder = builder.Limit(uint64(query.Limit))
	}
	if query.Offset >= 0 {
		builder = builder.Offset(uint64(query.Offset))
	}

	// Apply sorting
	if query.Sort != "" {
		sortColumn := "f_update_time" // Default sort by update_time
		switch query.Sort {
		case "name":
			sortColumn = "f_rule_name"
		case "create_time":
			sortColumn = "f_create_time"
		case "update_time":
			sortColumn = "f_update_time"
		}
		direction := "DESC"
		if query.Direction == "asc" {
			direction = "ASC"
		}
		builder = builder.OrderBy(sortColumn + " " + direction)
	}

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Build SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Build SQL failed")
		return nil, err
	}

	logger.Debugf("SQL: %s, Vals: %v", sqlStr, vals)

	rows, err := rcra.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		errDetails := fmt.Sprintf("Query SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Query SQL failed")
		return nil, err
	}
	defer rows.Close()

	rules := make([]*interfaces.ResourceRowColumnRule, 0)
	for rows.Next() {
		var rule interfaces.ResourceRowColumnRule
		var tagsStr string
		var fieldsStr string
		var rowFiltersStr string
		var creatorID, creatorType, updaterID, updaterType string

		err := rows.Scan(
			&rule.RuleID,
			&rule.RuleName,
			&rule.ResourceID,
			&tagsStr,
			&rule.Comment,
			&fieldsStr,
			&rowFiltersStr,
			&rule.CreateTime,
			&rule.UpdateTime,
			&creatorID,
			&creatorType,
			&updaterID,
			&updaterType,
		)
		if err != nil {
			errDetails := fmt.Sprintf("Scan row failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Scan row failed")
			return nil, err
		}

		rule.Tags = libCommon.TagString2TagSlice(tagsStr)

		err = sonic.UnmarshalString(fieldsStr, &rule.Fields)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal fields failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Unmarshal fields failed")
			return nil, err
		}

		err = sonic.UnmarshalString(rowFiltersStr, &rule.RowFilters)
		if err != nil {
			errDetails := fmt.Sprintf("Unmarshal row filters failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Unmarshal row filters failed")
			return nil, err
		}

		rule.Creator = interfaces.AccountInfo{
			ID:   creatorID,
			Type: creatorType,
		}
		rule.Updater = interfaces.AccountInfo{
			ID:   updaterID,
			Type: updaterType,
		}

		rules = append(rules, &rule)
	}

	span.SetStatus(codes.Ok, "")
	return rules, nil
}

// DeleteResourceRowColumnRules deletes resource row column rules by IDs.
func (rcra *resourceRowColumnRuleAccess) DeleteResourceRowColumnRules(ctx context.Context, tx *sql.Tx, ruleIDs []string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "driven layer: Delete resource row column rules from DB")
	defer span.End()

	if len(ruleIDs) == 0 {
		return nil
	}

	var sqlStr string
	var vals []interface{}

	if tx != nil {
		sqlStr, vals, _ = sq.Delete(RESOURCE_ROW_COLUMN_RULE_TABLE_NAME).
			Where(sq.Eq{"f_rule_id": ruleIDs}).
			ToSql()
		_, err := tx.Exec(sqlStr, vals...)
		if err != nil {
			errDetails := fmt.Sprintf("Exec SQL failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Exec SQL failed")
			return err
		}
	} else {
		sqlStr, vals, err := sq.Delete(RESOURCE_ROW_COLUMN_RULE_TABLE_NAME).
			Where(sq.Eq{"f_rule_id": ruleIDs}).
			ToSql()
		if err != nil {
			errDetails := fmt.Sprintf("Build SQL failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Build SQL failed")
			return err
		}

		_, err = rcra.db.ExecContext(ctx, sqlStr, vals...)
		if err != nil {
			errDetails := fmt.Sprintf("Exec SQL failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Exec SQL failed")
			return err
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// CheckResourceRowColumnRuleExistByID checks if a resource row column rule exists by ID.
func (rcra *resourceRowColumnRuleAccess) CheckResourceRowColumnRuleExistByID(ctx context.Context, ruleID string) (string, bool, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "driven layer: Check resource row column rule existence by ID")
	defer span.End()

	sqlStr, vals, err := sq.Select("f_rule_id").
		From(RESOURCE_ROW_COLUMN_RULE_TABLE_NAME).
		Where(sq.Eq{"f_rule_id": ruleID}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Build SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Build SQL failed")
		return "", false, err
	}

	var ruleIDResult string
	err = rcra.db.QueryRowContext(ctx, sqlStr, vals...).Scan(&ruleIDResult)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		errDetails := fmt.Sprintf("Query SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Query SQL failed")
		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return ruleIDResult, true, nil
}

// CheckResourceRowColumnRuleExistByName checks if a resource row column rule exists by name and resource ID.
func (rcra *resourceRowColumnRuleAccess) CheckResourceRowColumnRuleExistByName(ctx context.Context, ruleName, resourceID string) (string, bool, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "driven layer: Check resource row column rule existence by name")
	defer span.End()

	sqlStr, vals, err := sq.Select("f_rule_id").
		From(RESOURCE_ROW_COLUMN_RULE_TABLE_NAME).
		Where(sq.And{
			sq.Eq{"f_rule_name": ruleName},
			sq.Eq{"f_resource_id": resourceID},
		}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Build SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Build SQL failed")
		return "", false, err
	}

	var ruleIDResult string
	err = rcra.db.QueryRowContext(ctx, sqlStr, vals...).Scan(&ruleIDResult)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		errDetails := fmt.Sprintf("Query SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Query SQL failed")
		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return ruleIDResult, true, nil
}

// GetSimpleRulesByResourceIDs gets simple rules by resource IDs.
func (rcra *resourceRowColumnRuleAccess) GetSimpleRulesByResourceIDs(ctx context.Context, tx *sql.Tx, resourceIDs []string) ([]*interfaces.ResourceRowColumnRule, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "driven layer: Get simple rules by resource IDs")
	defer span.End()

	if len(resourceIDs) == 0 {
		return []*interfaces.ResourceRowColumnRule{}, nil
	}

	var sqlStr string
	var vals []interface{}

	selectBuilder := sq.Select(
		"f_rule_id",
		"f_rule_name",
		"f_resource_id",
		"f_fields",
		"f_row_filters",
	).
		From(RESOURCE_ROW_COLUMN_RULE_TABLE_NAME).
		Where(sq.Eq{"f_resource_id": resourceIDs})

	if tx != nil {
		sqlStr, vals, _ = selectBuilder.ToSql()
		rows, err := tx.Query(sqlStr, vals...)
		if err != nil {
			errDetails := fmt.Sprintf("Query SQL failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Query SQL failed")
			return nil, err
		}
		defer rows.Close()

		return parseSimpleRules(rows)
	}

	sqlStr, vals, err := selectBuilder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Build SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Build SQL failed")
		return nil, err
	}

	rows, err := rcra.db.QueryContext(ctx, sqlStr, vals...)
	if err != nil {
		errDetails := fmt.Sprintf("Query SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Query SQL failed")
		return nil, err
	}
	defer rows.Close()

	return parseSimpleRules(rows)
}

// GetSimpleRulesByRuleIDs gets simple rules by rule IDs.
func (rcra *resourceRowColumnRuleAccess) GetSimpleRulesByRuleIDs(ctx context.Context, tx *sql.Tx, ruleIDs []string) ([]*interfaces.ResourceRowColumnRule, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "driven layer: Get simple rules by rule IDs")
	defer span.End()

	if len(ruleIDs) == 0 {
		return []*interfaces.ResourceRowColumnRule{}, nil
	}

	selectBuilder := sq.Select(
		"f_rule_id",
		"f_rule_name",
		"f_resource_id",
		"f_fields",
		"f_row_filters",
	).
		From(RESOURCE_ROW_COLUMN_RULE_TABLE_NAME).
		Where(sq.Eq{"f_rule_id": ruleIDs})

	var sqlStr string
	var vals []interface{}
	var rows *sql.Rows
	var err error

	if tx != nil {
		sqlStr, vals, _ = selectBuilder.ToSql()
		rows, err = tx.Query(sqlStr, vals...)
	} else {
		sqlStr, vals, err = selectBuilder.ToSql()
		if err != nil {
			errDetails := fmt.Sprintf("Build SQL failed, %s", err.Error())
			logger.Error(errDetails)
			otellog.LogError(ctx, errDetails, nil)
			span.SetStatus(codes.Error, "Build SQL failed")
			return nil, err
		}
		rows, err = rcra.db.QueryContext(ctx, sqlStr, vals...)
	}

	if err != nil {
		errDetails := fmt.Sprintf("Query SQL failed, %s", err.Error())
		logger.Error(errDetails)
		otellog.LogError(ctx, errDetails, nil)
		span.SetStatus(codes.Error, "Query SQL failed")
		return nil, err
	}
	defer rows.Close()

	return parseSimpleRules(rows)
}

// parseSimpleRules parses rows into simple rules.
func parseSimpleRules(rows *sql.Rows) ([]*interfaces.ResourceRowColumnRule, error) {
	rules := make([]*interfaces.ResourceRowColumnRule, 0)
	for rows.Next() {
		var rule interfaces.ResourceRowColumnRule
		var fieldsStr string
		var rowFiltersStr string

		err := rows.Scan(
			&rule.RuleID,
			&rule.RuleName,
			&rule.ResourceID,
			&fieldsStr,
			&rowFiltersStr,
		)
		if err != nil {
			return nil, err
		}

		err = sonic.UnmarshalString(fieldsStr, &rule.Fields)
		if err != nil {
			return nil, err
		}

		err = sonic.UnmarshalString(rowFiltersStr, &rule.RowFilters)
		if err != nil {
			return nil, err
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}
