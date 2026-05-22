// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateSQLSyntax(t *testing.T) {
	Convey("Test validateSQLSyntax", t, func() {
		Convey("Valid SQL with node variable", func() {
			assertSQLValid("SELECT * FROM .node1")
		})

		Convey("Valid SQL with multiple node variables", func() {
			assertSQLValid("SELECT .node1.id, .node2.name FROM .node1 JOIN .node2 ON .node1.id = .node2.user_id")
		})

		Convey("Valid SQL with node variable in subquery", func() {
			assertSQLValid("SELECT * FROM (SELECT id, name FROM .node1) AS subq")
		})

		Convey("Valid SQL with node variable and WHERE", func() {
			assertSQLValid("SELECT * FROM .node1 WHERE .node1.age > 18")
		})

		Convey("Valid SQL with node variable and GROUP BY", func() {
			assertSQLValid("SELECT .node1.department, COUNT(*) FROM .node1 GROUP BY .node1.department")
		})

		Convey("Valid SQL with node variable and ORDER BY", func() {
			assertSQLValid("SELECT * FROM .node1 ORDER BY .node1.created_at DESC LIMIT 10")
		})

		Convey("Valid simple SELECT", func() {
			assertSQLValid("SELECT * FROM users")
		})

		Convey("Valid SELECT with columns", func() {
			assertSQLValid("SELECT id, name, email FROM users WHERE age > 18")
		})

		Convey("Valid SELECT with JOIN", func() {
			assertSQLValid("SELECT u.id, o.total FROM users u JOIN orders o ON u.id = o.user_id")
		})

		Convey("Valid SQL with GROUP BY", func() {
			assertSQLValid("SELECT department, COUNT(*) FROM employees GROUP BY department")
		})

		Convey("Valid SQL with ORDER BY", func() {
			assertSQLValid("SELECT * FROM users ORDER BY created_at DESC LIMIT 10")
		})

		Convey("Valid SQL with subquery", func() {
			assertSQLValid("SELECT * FROM (SELECT id, name FROM users) AS subq")
		})

		Convey("Valid SQL with DISTINCT", func() {
			assertSQLValid("SELECT DISTINCT name FROM users")
		})

		Convey("Valid SQL with WITH clause", func() {
			assertSQLValid("WITH temp AS (SELECT * FROM users) SELECT * FROM temp")
		})

		Convey("Empty SQL", func() {
			assertSQLValid("")
		})

		Convey("Invalid SQL with double FROM", func() {
			assertSQLInvalid("SELECT * FROM FROM users", "Duplicate FROM")
		})

		Convey("Invalid SQL with double SELECT", func() {
			assertSQLInvalid("SELECT SELECT * FROM users", "Duplicate SELECT")
		})

		Convey("Invalid SQL missing table after FROM", func() {
			assertSQLInvalid("SELECT * FROM", "FROM clause must specify a table")
		})

		Convey("Invalid SQL with unclosed parenthesis", func() {
			assertSQLInvalid("SELECT * FROM users WHERE (id = 1", "Unbalanced parentheses")
		})

		Convey("Invalid SQL with extra closing parenthesis", func() {
			assertSQLInvalid("SELECT * FROM users WHERE id = 1)", "Unbalanced parentheses")
		})

		Convey("Invalid SQL missing SELECT keyword", func() {
			assertSQLInvalid("* FROM users", "must start with SELECT")
		})

		Convey("Invalid SQL with WHERE without condition", func() {
			assertSQLInvalid("SELECT * FROM users WHERE", "WHERE clause must have a condition")
		})

		Convey("Invalid SQL with GROUP BY without column", func() {
			assertSQLInvalid("SELECT * FROM users GROUP BY", "GROUP BY must have at least one column")
		})

		Convey("Invalid SQL with ORDER BY without column", func() {
			assertSQLInvalid("SELECT * FROM users ORDER BY", "ORDER BY must have at least one column")
		})

		Convey("Invalid SQL with dot notation without table", func() {
			assertSQLInvalid("SELECT * FROM FROM .node1", "Duplicate FROM")
		})

		Convey("Invalid SQL with SELECT without FROM", func() {
			assertSQLInvalid("SELECT id, name", "must contain a FROM clause")
		})
	})
}

func assertSQLValid(sql string) {
	err := validateSQLSyntax(context.Background(), sql)
	So(err, ShouldBeNil)
}

func assertSQLInvalid(sql string, errorContains string) {
	err := validateSQLSyntax(context.Background(), sql)
	So(err, ShouldNotBeNil)
	So(err.Error(), ShouldContainSubstring, errorContains)
}
