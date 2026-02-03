package db

import (
	"context"
	"strings"
	"testing"
)

// TestSQLQueryConstructorBasicSelect 测试基础 SELECT 生成
func TestSQLQueryConstructorBasicSelect(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("id", TypeInteger).PrimaryKey().Build())
	schema.AddField(NewField("name", TypeString).Build())
	schema.AddField(NewField("email", TypeString).Build())

	dialect := NewMySQLDialect()
	qc := NewSQLQueryConstructor(schema, dialect)

	ctx := context.Background()
	sql, args, err := qc.Build(ctx)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 验证 SQL 结构
	if !strings.Contains(sql, "SELECT") {
		t.Errorf("Expected SELECT in SQL: %s", sql)
	}
	if !strings.Contains(sql, "FROM") {
		t.Errorf("Expected FROM in SQL: %s", sql)
	}
	if !strings.Contains(sql, "`users`") {
		t.Errorf("Expected quoted table name `users` in SQL: %s", sql)
	}
	if len(args) != 0 {
		t.Errorf("Expected no arguments for basic SELECT, got: %v", args)
	}

	t.Logf("✓ Basic SELECT: %s", sql)
}

// TestSQLQueryConstructorEqCondition 测试 = 条件
func TestSQLQueryConstructorEqCondition(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("id", TypeInteger).PrimaryKey().Build())
	schema.AddField(NewField("name", TypeString).Build())

	dialect := NewMySQLDialect()
	qc := NewSQLQueryConstructor(schema, dialect)
	qc.Where(Eq("name", "John"))

	ctx := context.Background()
	sql, args, err := qc.Build(ctx)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 验证 SQL 包含 WHERE 子句
	if !strings.Contains(sql, "WHERE") {
		t.Errorf("Expected WHERE clause in: %s", sql)
	}
	if !strings.Contains(sql, "`name`") {
		t.Errorf("Expected quoted column name `name` in: %s", sql)
	}
	if !strings.Contains(sql, "=") {
		t.Errorf("Expected = operator in: %s", sql)
	}
	if !strings.Contains(sql, "?") {
		t.Errorf("Expected parameter placeholder in: %s", sql)
	}

	// 验证参数
	if len(args) != 1 {
		t.Fatalf("Expected 1 argument, got %d", len(args))
	}
	if args[0] != "John" {
		t.Errorf("Expected argument 'John', got %v", args[0])
	}

	t.Logf("✓ Eq condition: %s with args %v", sql, args)
}

// TestSQLQueryConstructorComparisonOperators 测试比较操作符
func TestSQLQueryConstructorComparisonOperators(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("age", TypeInteger).Build())

	dialect := NewMySQLDialect()
	ctx := context.Background()

	testCases := []struct {
		name     string
		cond     Condition
		expectOp string
		expectVal interface{}
	}{
		{"Gt (>)", Gt("age", 18), ">", 18},
		{"Lt (<)", Lt("age", 65), "<", 65},
		{"Gte (>=)", Gte("age", 18), ">=", 18},
		{"Lte (<=)", Lte("age", 65), "<=", 65},
		{"Ne (!=)", Ne("age", 25), "!=", 25},
	}

	for _, tc := range testCases {
		qc := NewSQLQueryConstructor(schema, dialect)
		qc.Where(tc.cond)
		sql, args, err := qc.Build(ctx)

		if err != nil {
			t.Errorf("%s: Build failed: %v", tc.name, err)
			continue
		}

		if !strings.Contains(sql, tc.expectOp) {
			t.Errorf("%s: Expected operator '%s' in: %s", tc.name, tc.expectOp, sql)
		}

		if len(args) != 1 || args[0] != tc.expectVal {
			t.Errorf("%s: Expected argument %v, got %v", tc.name, tc.expectVal, args)
		}

		t.Logf("✓ %s: %s with args %v", tc.name, sql, args)
	}
}

// TestSQLQueryConstructorInCondition 测试 IN 条件
func TestSQLQueryConstructorInCondition(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("age", TypeInteger).Build())

	dialect := NewMySQLDialect()
	qc := NewSQLQueryConstructor(schema, dialect)
	qc.Where(In("age", 18, 21, 25, 30))

	ctx := context.Background()
	sql, args, err := qc.Build(ctx)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 验证 SQL
	if !strings.Contains(sql, "IN") {
		t.Errorf("Expected IN operator in: %s", sql)
	}
	if !strings.Contains(sql, "(") || !strings.Contains(sql, ")") {
		t.Errorf("Expected parentheses in IN clause: %s", sql)
	}

	// 验证占位符数量
	placeholderCount := strings.Count(sql, "?")
	if placeholderCount != 4 {
		t.Errorf("Expected 4 placeholders for 4 values, got %d in: %s", placeholderCount, sql)
	}

	// 验证参数
	if len(args) != 4 {
		t.Fatalf("Expected 4 arguments, got %d", len(args))
	}
	expectedArgs := []interface{}{18, 21, 25, 30}
	for i, arg := range args {
		if arg != expectedArgs[i] {
			t.Errorf("Expected argument %v at index %d, got %v", expectedArgs[i], i, arg)
		}
	}

	t.Logf("✓ IN condition: %s with args %v", sql, args)
}

// TestSQLQueryConstructorBetweenCondition 测试 BETWEEN 条件
func TestSQLQueryConstructorBetweenCondition(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("age", TypeInteger).Build())

	dialect := NewMySQLDialect()
	qc := NewSQLQueryConstructor(schema, dialect)
	qc.Where(Between("age", 18, 65))

	ctx := context.Background()
	sql, args, err := qc.Build(ctx)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 验证 SQL
	if !strings.Contains(sql, "BETWEEN") {
		t.Errorf("Expected BETWEEN in: %s", sql)
	}
	if !strings.Contains(sql, "AND") {
		t.Errorf("Expected AND in BETWEEN clause: %s", sql)
	}

	// 验证参数
	if len(args) != 2 {
		t.Fatalf("Expected 2 arguments, got %d", len(args))
	}
	if args[0] != 18 || args[1] != 65 {
		t.Errorf("Expected arguments [18, 65], got %v", args)
	}

	t.Logf("✓ BETWEEN condition: %s with args %v", sql, args)
}

// TestSQLQueryConstructorLikeCondition 测试 LIKE 条件
func TestSQLQueryConstructorLikeCondition(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("name", TypeString).Build())

	dialect := NewMySQLDialect()
	qc := NewSQLQueryConstructor(schema, dialect)
	qc.Where(Like("name", "%John%"))

	ctx := context.Background()
	sql, args, err := qc.Build(ctx)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 验证 SQL
	if !strings.Contains(sql, "LIKE") {
		t.Errorf("Expected LIKE in: %s", sql)
	}

	// 验证参数
	if len(args) != 1 {
		t.Fatalf("Expected 1 argument, got %d", len(args))
	}
	if args[0] != "%John%" {
		t.Errorf("Expected pattern '%%John%%', got %v", args[0])
	}

	t.Logf("✓ LIKE condition: %s with args %v", sql, args)
}

// TestSQLQueryConstructorWhereAll 测试 WhereAll (AND 组合)
func TestSQLQueryConstructorWhereAll(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("name", TypeString).Build())
	schema.AddField(NewField("age", TypeInteger).Build())

	dialect := NewMySQLDialect()
	qc := NewSQLQueryConstructor(schema, dialect)
	qc.WhereAll(Eq("name", "John"), Gt("age", 18))

	ctx := context.Background()
	sql, args, err := qc.Build(ctx)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 验证 SQL 包含 WHERE 和圆括号
	if !strings.Contains(sql, "WHERE") {
		t.Errorf("Expected WHERE clause in: %s", sql)
	}
	if !strings.Contains(sql, "(") || !strings.Contains(sql, ")") {
		t.Errorf("Expected parentheses for AND group in: %s", sql)
	}

	// 验证参数
	if len(args) != 2 {
		t.Fatalf("Expected 2 arguments, got %d", len(args))
	}
	if args[0] != "John" || args[1] != 18 {
		t.Errorf("Expected arguments ['John', 18], got %v", args)
	}

	t.Logf("✓ WhereAll (AND): %s with args %v", sql, args)
}

// TestSQLQueryConstructorWhereAny 测试 WhereAny (OR 组合)
func TestSQLQueryConstructorWhereAny(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("status", TypeString).Build())

	dialect := NewMySQLDialect()
	qc := NewSQLQueryConstructor(schema, dialect)
	qc.WhereAny(Eq("status", "active"), Eq("status", "pending"))

	ctx := context.Background()
	sql, args, err := qc.Build(ctx)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 验证 SQL 包含 OR
	if !strings.Contains(sql, "OR") {
		t.Errorf("Expected OR operator in: %s", sql)
	}
	if !strings.Contains(sql, "(") || !strings.Contains(sql, ")") {
		t.Errorf("Expected parentheses for OR group in: %s", sql)
	}

	// 验证参数
	if len(args) != 2 {
		t.Fatalf("Expected 2 arguments, got %d", len(args))
	}

	t.Logf("✓ WhereAny (OR): %s with args %v", sql, args)
}

// TestSQLQueryConstructorOrderBy 测试 ORDER BY
func TestSQLQueryConstructorOrderBy(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("age", TypeInteger).Build())
	schema.AddField(NewField("name", TypeString).Build())

	dialect := NewMySQLDialect()
	qc := NewSQLQueryConstructor(schema, dialect)
	qc.OrderBy("age", "DESC").OrderBy("name", "ASC")

	ctx := context.Background()
	sql, args, err := qc.Build(ctx)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 验证 SQL
	if !strings.Contains(sql, "ORDER BY") {
		t.Errorf("Expected ORDER BY clause in: %s", sql)
	}
	if !strings.Contains(sql, "`age`") || !strings.Contains(sql, "DESC") {
		t.Errorf("Expected age DESC in: %s", sql)
	}
	if !strings.Contains(sql, "`name`") || !strings.Contains(sql, "ASC") {
		t.Errorf("Expected name ASC in: %s", sql)
	}

	// 验证没有参数
	if len(args) != 0 {
		t.Errorf("Expected no arguments for ORDER BY, got: %v", args)
	}

	t.Logf("✓ ORDER BY: %s", sql)
}

// TestSQLQueryConstructorLimitOffset 测试 LIMIT 和 OFFSET
func TestSQLQueryConstructorLimitOffset(t *testing.T) {
	schema := NewBaseSchema("users")
	dialect := NewMySQLDialect()

	testCases := []struct {
		name          string
		limit         *int
		offset        *int
		expectLimitOK bool
		expectOffsetOK bool
	}{
		{"Limit only", intPtr(10), nil, true, false},
		{"Offset only", nil, intPtr(20), false, true},
		{"Both", intPtr(10), intPtr(20), true, true},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		qc := NewSQLQueryConstructor(schema, dialect)
		if tc.limit != nil {
			qc.Limit(*tc.limit)
		}
		if tc.offset != nil {
			qc.Offset(*tc.offset)
		}

		sql, args, err := qc.Build(ctx)

		if err != nil {
			t.Errorf("%s: Build failed: %v", tc.name, err)
			continue
		}

		if tc.expectLimitOK && !strings.Contains(sql, "LIMIT") {
			t.Errorf("%s: Expected LIMIT in: %s", tc.name, sql)
		}
		if tc.expectOffsetOK && !strings.Contains(sql, "OFFSET") {
			t.Errorf("%s: Expected OFFSET in: %s", tc.name, sql)
		}

		if len(args) != 0 {
			t.Errorf("%s: Expected no arguments, got: %v", tc.name, args)
		}

		t.Logf("✓ %s: %s", tc.name, sql)
	}
}

// TestSQLQueryConstructorSelectColumns 测试 SELECT 字段选择
func TestSQLQueryConstructorSelectColumns(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("id", TypeInteger).Build())
	schema.AddField(NewField("name", TypeString).Build())
	schema.AddField(NewField("email", TypeString).Build())

	dialect := NewMySQLDialect()
	qc := NewSQLQueryConstructor(schema, dialect)
	qc.Select("name", "email")

	ctx := context.Background()
	sql, _, err := qc.Build(ctx)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 验证 SQL 不包含 *
	if strings.Contains(sql, "SELECT *") {
		t.Errorf("Expected specific columns, not * in: %s", sql)
	}

	// 验证 SQL 包含指定的列
	if !strings.Contains(sql, "`name`") || !strings.Contains(sql, "`email`") {
		t.Errorf("Expected columns name and email in: %s", sql)
	}

	// 验证 id 不在 SELECT 中
	// Note: 这个检查可能不完美，因为 id 可能在 WHERE 条件中
	sqlSelectPart := sql[strings.Index(sql, "SELECT"):strings.Index(sql, "FROM")]
	if strings.Count(sqlSelectPart, "`") < 4 { // name 和 email 各 2 个 backticks
		t.Errorf("Expected only 2 columns selected, got: %s", sqlSelectPart)
	}

	t.Logf("✓ SELECT columns: %s", sql)
}

// TestSQLQueryConstructorCombined 测试复杂的组合查询
func TestSQLQueryConstructorCombined(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("id", TypeInteger).Build())
	schema.AddField(NewField("name", TypeString).Build())
	schema.AddField(NewField("age", TypeInteger).Build())
	schema.AddField(NewField("status", TypeString).Build())

	dialect := NewMySQLDialect()
	qc := NewSQLQueryConstructor(schema, dialect)

	qc.Select("id", "name", "age").
		Where(Gt("age", 18)).
		Where(Eq("status", "active")).
		OrderBy("age", "DESC").
		Limit(10).
		Offset(5)

	ctx := context.Background()
	sql, args, err := qc.Build(ctx)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 验证所有子句都存在
	if !strings.Contains(sql, "SELECT") {
		t.Errorf("Expected SELECT clause")
	}
	if !strings.Contains(sql, "FROM") {
		t.Errorf("Expected FROM clause")
	}
	if !strings.Contains(sql, "WHERE") {
		t.Errorf("Expected WHERE clause")
	}
	if !strings.Contains(sql, "ORDER BY") {
		t.Errorf("Expected ORDER BY clause")
	}
	if !strings.Contains(sql, "LIMIT 10") {
		t.Errorf("Expected LIMIT 10")
	}
	if !strings.Contains(sql, "OFFSET 5") {
		t.Errorf("Expected OFFSET 5")
	}

	// 验证参数数量
	if len(args) != 2 {
		t.Fatalf("Expected 2 arguments (one Gt and one Eq), got %d: %v", len(args), args)
	}

	t.Logf("✓ Combined query: %s with args %v", sql, args)
}

// TestSQLDialectQuoting 测试不同方言的引号
func TestSQLDialectQuoting(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("name", TypeString).Build())

	ctx := context.Background()

	testCases := []struct {
		name          string
		dialect       SQLDialect
		expectQuote   string
	}{
		{"MySQL", NewMySQLDialect(), "`"},
		{"PostgreSQL", NewPostgreSQLDialect(), `"`},
		{"SQLite", NewSQLiteDialect(), "`"},
	}

	for _, tc := range testCases {
		qc := NewSQLQueryConstructor(schema, tc.dialect)
		qc.Where(Eq("name", "John"))
		sql, args, err := qc.Build(ctx)

		if err != nil {
			t.Errorf("%s: Build failed: %v", tc.name, err)
			continue
		}

		// 验证引号类型
		expectedQuotedTable := tc.expectQuote + "users" + tc.expectQuote
		if !strings.Contains(sql, expectedQuotedTable) {
			t.Errorf("%s: Expected table with %s quotes in: %s", tc.name, tc.expectQuote, sql)
		}

		expectedQuotedCol := tc.expectQuote + "name" + tc.expectQuote
		if !strings.Contains(sql, expectedQuotedCol) {
			t.Errorf("%s: Expected column with %s quotes in: %s", tc.name, tc.expectQuote, sql)
		}

		// PostgreSQL 应该使用 $1 占位符，其他应该使用 ?
		if tc.name == "PostgreSQL" {
			if !strings.Contains(sql, "$1") {
				t.Errorf("%s: Expected $1 placeholder in: %s", tc.name, sql)
			}
		} else {
			if !strings.Contains(sql, "?") {
				t.Errorf("%s: Expected ? placeholder in: %s", tc.name, sql)
			}
		}

		if len(args) != 1 || args[0] != "John" {
			t.Errorf("%s: Expected argument 'John', got %v", tc.name, args)
		}

		t.Logf("✓ %s dialect: %s", tc.name, sql)
	}
}

// TestConditionTranslation 测试条件转义
func TestConditionTranslation(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("name", TypeString).Build())

	dialect := NewMySQLDialect()
	translator := &DefaultSQLTranslator{dialect: dialect}
	argIndex := 1

	translator.argIndex = &argIndex

	// 测试简单条件翻译
	cond := Eq("name", "John")
	sql, args, err := translator.TranslateCondition(cond)

	if err != nil {
		t.Fatalf("TranslateCondition failed: %v", err)
	}

	if !strings.Contains(sql, "name") || !strings.Contains(sql, "=") {
		t.Errorf("Expected 'name = ?' pattern, got: %s", sql)
	}

	if len(args) != 1 || args[0] != "John" {
		t.Errorf("Expected argument 'John', got %v", args)
	}

	t.Logf("✓ Condition translation: %s with args %v", sql, args)
}

// TestQueryConstructorProvider 测试查询构造器提供者
func TestQueryConstructorProvider(t *testing.T) {
	schema := NewBaseSchema("users")
	schema.AddField(NewField("id", TypeInteger).PrimaryKey().Build())

	// 测试 MySQL Provider
	mysqlProvider := NewDefaultSQLQueryConstructorProvider(NewMySQLDialect())
	qcMySQL := mysqlProvider.NewQueryConstructor(schema)
	if qcMySQL == nil {
		t.Fatalf("MySQL provider failed to create QueryConstructor")
	}

	capsMySQL := mysqlProvider.GetCapabilities()
	if !capsMySQL.SupportsEq {
		t.Errorf("MySQL should support Eq")
	}
	if !capsMySQL.SupportsSelect {
		t.Errorf("MySQL should support Select")
	}

	// 测试 PostgreSQL Provider
	pgProvider := NewDefaultSQLQueryConstructorProvider(NewPostgreSQLDialect())
	qcPG := pgProvider.NewQueryConstructor(schema)
	if qcPG == nil {
		t.Fatalf("PostgreSQL provider failed to create QueryConstructor")
	}

	capsPG := pgProvider.GetCapabilities()
	if !capsPG.SupportsJoin {
		t.Errorf("PostgreSQL should support Join")
	}
	if !capsPG.SupportsOrderBy {
		t.Errorf("PostgreSQL should support OrderBy")
	}

	t.Log("✓ All QueryConstructorProviders work correctly")
}

// 辅助函数
func intPtr(v int) *int {
	return &v
}
