-- Add resource row column rule table
-- Version: 0.9.0
-- Description: Create t_resource_row_column_rule table for storing resource row/column permission rules

CREATE TABLE IF NOT EXISTS `t_resource_row_column_rule` (
  `f_rule_id` varchar(64) NOT NULL COMMENT 'Rule ID',
  `f_rule_name` varchar(255) NOT NULL COMMENT 'Rule name',
  `f_resource_id` varchar(64) NOT NULL COMMENT 'Resource ID',
  `f_tags` varchar(1024) DEFAULT '' COMMENT 'Tags, separated by commas',
  `f_comment` varchar(1024) DEFAULT '' COMMENT 'Comment',
  `f_fields` text DEFAULT NULL COMMENT 'Fields (JSON array)',
  `f_row_filters` text DEFAULT NULL COMMENT 'Row filters (JSON)',
  `f_create_time` bigint NOT NULL COMMENT 'Create time (milliseconds timestamp)',
  `f_update_time` bigint NOT NULL COMMENT 'Update time (milliseconds timestamp)',
  `f_creator` varchar(64) NOT NULL COMMENT 'Creator ID',
  `f_creator_type` varchar(32) NOT NULL COMMENT 'Creator type (user/app/role)',
  `f_updater` varchar(64) NOT NULL COMMENT 'Updater ID',
  `f_updater_type` varchar(32) NOT NULL COMMENT 'Updater type (user/app/role)',
  PRIMARY KEY (`f_rule_id`),
  KEY `idx_resource_id` (`f_resource_id`),
  KEY `idx_rule_name` (`f_rule_name`),
  KEY `idx_update_time` (`f_update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Resource row column rule table';
