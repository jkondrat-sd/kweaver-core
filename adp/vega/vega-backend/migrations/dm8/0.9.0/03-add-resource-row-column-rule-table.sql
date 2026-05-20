-- Add resource row column rule table
-- Version: 0.9.0
-- Description: Create t_resource_row_column_rule table for storing resource row/column permission rules

CREATE TABLE IF NOT EXISTS "t_resource_row_column_rule" (
  "f_rule_id" VARCHAR(64 CHAR) NOT NULL,
  "f_rule_name" VARCHAR(255 CHAR) NOT NULL,
  "f_resource_id" VARCHAR(64 CHAR) NOT NULL,
  "f_tags" VARCHAR(1024 CHAR) DEFAULT '',
  "f_comment" VARCHAR(1024 CHAR) DEFAULT '',
  "f_fields" TEXT DEFAULT NULL,
  "f_row_filters" TEXT DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL,
  "f_update_time" BIGINT NOT NULL,
  "f_creator" VARCHAR(64 CHAR) NOT NULL,
  "f_creator_type" VARCHAR(32 CHAR) NOT NULL,
  "f_updater" VARCHAR(64 CHAR) NOT NULL,
  "f_updater_type" VARCHAR(32 CHAR) NOT NULL,
  PRIMARY KEY ("f_rule_id")
);

CREATE INDEX "idx_resource_id" ON "t_resource_row_column_rule" ("f_resource_id");
CREATE INDEX "idx_rule_name" ON "t_resource_row_column_rule" ("f_rule_name");
CREATE INDEX "idx_update_time" ON "t_resource_row_column_rule" ("f_update_time");