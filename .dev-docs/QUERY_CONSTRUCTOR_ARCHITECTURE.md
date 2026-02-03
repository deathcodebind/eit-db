# Query Constructor 三层架构 - v0.4.1

## 概述

EIT-DB v0.4.1 引入了灵活的三层查询构造器架构，为支持多数据库和多种查询语言（SQL、Cypher、聚合管道等）奠定了基础。

## 架构设计

```
┌─────────────────────────────────────────────────────────┐
│ 顶层：用户 API Layer                                     │
│ ✓ QueryConstructor 接口                                 │
│ ✓ Condition builders (Eq, Ne, Gt, Lt, In, Between...)  │
│ ✓ 链式 API (Where, Select, OrderBy, Limit, Offset)     │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 中层：Adapter 转义 Layer                                │
│ ✓ QueryConstructorProvider 接口                         │
│ ✓ QueryBuilderCapabilities 声明                         │
│ ✓ 每个 Adapter 声明支持的功能                            │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 底层：DB 执行 Layer                                      │
│ ✓ SQLQueryConstructor 实现                              │
│ ✓ SQLDialect 接口（MySQL/PostgreSQL/SQLite）           │
│ ✓ ConditionTranslator 条件转义                          │
└─────────────────────────────────────────────────────────┘
```

## 顶层 API - QueryConstructor 接口

```go
type QueryConstructor interface {
    // 条件查询
    Where(condition Condition) QueryConstructor
    WhereAll(conditions ...Condition) QueryConstructor  // AND 组合
    WhereAny(conditions ...Condition) QueryConstructor  // OR 组合
    
    // 字段选择
    Select(fields ...string) QueryConstructor
    
    // 排序
    OrderBy(field string, direction string) QueryConstructor // "ASC" | "DESC"
    
    // 分页
    Limit(count int) QueryConstructor
    Offset(count int) QueryConstructor
    
    // 构建查询
    Build(ctx context.Context) (string, []interface{}, error)
    
    // 获取底层构造器（用于适配器特定优化）
    GetNativeBuilder() interface{}
}
```

### 条件 Builders

顶层提供了便利的条件构造函数：

```go
// 比较操作
Eq(field, value)           // 等于
Ne(field, value)           // 不等于
Gt(field, value)           // 大于
Lt(field, value)           // 小于
Gte(field, value)          // 大于等于
Lte(field, value)          // 小于等于

// 集合操作
In(field, values...)       // IN 条件
Between(field, min, max)   // BETWEEN 条件
Like(field, pattern)       // 模糊匹配

// 逻辑操作
And(conditions...)         // AND 组合
Or(conditions...)          // OR 组合
Not(condition)             // 非操作
```

### Condition 接口

```go
type Condition interface {
    // 获取条件类型 ("simple", "composite", "not")
    Type() string
    
    // 转义条件
    Translate(translator ConditionTranslator) (string, []interface{}, error)
}
```

## 中层 - Adapter 转义层

### QueryConstructorProvider 接口

每个 Adapter 必须实现此接口：

```go
type QueryConstructorProvider interface {
    // 创建新的查询构造器
    NewQueryConstructor(schema Schema) QueryConstructor
    
    // 获取此 Adapter 的查询能力声明
    GetCapabilities() *QueryBuilderCapabilities
}
```

### QueryBuilderCapabilities 能力声明

```go
type QueryBuilderCapabilities struct {
    // 支持的条件操作
    SupportsEq       bool
    SupportsNe       bool
    SupportsGt       bool
    SupportsLt       bool
    SupportsGte      bool
    SupportsLte      bool
    SupportsIn       bool
    SupportsBetween  bool
    SupportsLike     bool
    SupportsAnd      bool
    SupportsOr       bool
    SupportsNot      bool
    
    // 支持的查询特性
    SupportsSelect   bool        // 字段选择
    SupportsOrderBy  bool        // 排序
    SupportsLimit    bool        // LIMIT
    SupportsOffset   bool        // OFFSET
    SupportsJoin     bool        // JOIN（关系查询）
    SupportsSubquery bool        // 子查询
    
    // 优化特性
    SupportsQueryPlan bool       // 查询计划分析
    SupportsIndex     bool       // 索引提示
    
    // 原生查询支持
    SupportsNativeQuery bool     // 是否支持原生查询
    NativeQueryLang     string   // 原生查询语言名称
    
    // 其他标记
    Description string           // 简要描述
}
```

### Adapter 实现示例

```go
// MySQL Adapter 实现
func (a *MySQLAdapter) GetQueryBuilderProvider() QueryConstructorProvider {
    return NewDefaultSQLQueryConstructorProvider(NewMySQLDialect())
}

// PostgreSQL Adapter 实现
func (a *PostgreSQLAdapter) GetQueryBuilderProvider() QueryConstructorProvider {
    return NewDefaultSQLQueryConstructorProvider(NewPostgreSQLDialect())
}

// SQLite Adapter 实现
func (a *SQLiteAdapter) GetQueryBuilderProvider() QueryConstructorProvider {
    return NewDefaultSQLQueryConstructorProvider(NewSQLiteDialect())
}
```

## 底层 - DB 执行层

### SQLQueryConstructor 实现

```go
type SQLQueryConstructor struct {
    schema       Schema
    dialect      SQLDialect
    selectedCols []string
    conditions   []Condition
    orderBys     []OrderBy
    limitVal     *int
    offsetVal    *int
}
```

### SQLDialect 接口

不同的数据库通过实现 SQLDialect 来提供特定的 SQL 生成逻辑：

```go
type SQLDialect interface {
    // 获取方言名称
    Name() string
    
    // 转义标识符（表名、列名）
    QuoteIdentifier(name string) string
    
    // 转义字符串值
    QuoteValue(value interface{}) string
    
    // 返回参数化占位符
    GetPlaceholder(index int) string
    
    // 生成 LIMIT/OFFSET 子句
    GenerateLimitOffset(limit *int, offset *int) string
    
    // 转换条件为 SQL
    TranslateCondition(condition Condition, argIndex *int) (string, []interface{}, error)
}
```

### 内置方言实现

#### MySQL 方言

```go
// 标识符使用反引号
dialect := NewMySQLDialect()
// 生成：SELECT `id`, `name` FROM `users` WHERE `age` > ?
```

#### PostgreSQL 方言

```go
// 标识符使用双引号，参数使用 $1, $2
dialect := NewPostgreSQLDialect()
// 生成：SELECT "id", "name" FROM "users" WHERE "age" > $1
```

#### SQLite 方言

```go
// 标识符使用反引号，参数使用 ?
dialect := NewSQLiteDialect()
// 生成：SELECT `id`, `name` FROM `users` WHERE `age` > ?
```

### ConditionTranslator 接口

```go
type ConditionTranslator interface {
    TranslateCondition(condition Condition) (string, []interface{}, error)
    TranslateComposite(operator string, conditions []Condition) (string, []interface{}, error)
}
```

## 使用示例

### 基本查询

```go
package main

import (
    "context"
    "log"
    "github.com/eit-cms/eit-db"
)

func main() {
    // 1. 创建 Repository
    config := &db.Config{
        Adapter: "mysql",
        Host:    "localhost",
        Port:    3306,
        Username: "root",
        Password: "password",
        Database: "myapp",
    }
    repo, err := db.NewRepository(config)
    if err != nil {
        panic(err)
    }
    defer repo.Close()
    
    // 2. 创建 Schema
    userSchema := db.NewBaseSchema("users")
    userSchema.AddField(db.NewField("id", db.TypeInteger).PrimaryKey().Build())
    userSchema.AddField(db.NewField("name", db.TypeString).Build())
    userSchema.AddField(db.NewField("age", db.TypeInteger).Build())
    
    // 3. 获取 QueryConstructorProvider
    adapter := repo.GetAdapter()
    provider := adapter.GetQueryBuilderProvider()
    
    // 4. 创建查询
    ctx := context.Background()
    qc := provider.NewQueryConstructor(userSchema)
    
    qc.Where(Gt("age", 18)).
        OrderBy("age", "DESC").
        Limit(10)
    
    sql, args, err := qc.Build(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // 5. 执行查询
    rows, err := repo.Query(ctx, sql, args...)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
}
```

### 复杂条件

```go
// 多条件 AND
qc.WhereAll(
    Eq("status", "active"),
    Gt("age", 18),
    Lt("age", 65),
)

// 多条件 OR
qc.WhereAny(
    Eq("role", "admin"),
    Eq("role", "moderator"),
)

// 混合条件
qc.Where(
    And(
        Eq("status", "active"),
        Or(
            Eq("role", "admin"),
            Eq("role", "moderator"),
        ),
    ),
)

// IN 条件
qc.Where(In("status", "active", "pending", "approved"))

// BETWEEN 条件
qc.Where(Between("age", 18, 65))

// LIKE 条件
qc.Where(Like("name", "%John%"))
```

### 能力检查

```go
// 检查 Adapter 的能力
caps := provider.GetCapabilities()

if caps.SupportsJoin {
    // 执行 JOIN 查询
}

if caps.SupportsNativeQuery {
    // 使用原生查询语言
    nativeBuilder := qc.GetNativeBuilder()
    // 对于 Neo4j，可能是 CypherQueryBuilder
    // 对于 MongoDB，可能是 AggregationPipelineBuilder
}

log.Printf("Query Builder: %s", caps.Description)
log.Printf("Supports: Eq=%v, In=%v, Join=%v, NativeQuery=%v",
    caps.SupportsEq, caps.SupportsIn, caps.SupportsJoin, caps.SupportsNativeQuery)
```

## 扩展 - 自定义方言

如果需要支持新的数据库，可以实现自定义 SQLDialect：

```go
type CustomDialect struct {
    DefaultSQLDialect
}

func (d *CustomDialect) QuoteIdentifier(name string) string {
    return "[" + name + "]"  // SQL Server 风格
}

func (d *CustomDialect) GetPlaceholder(index int) string {
    return fmt.Sprintf("@p%d", index)  // SQL Server 风格
}

// 在 Adapter 中使用
func (a *CustomAdapter) GetQueryBuilderProvider() QueryConstructorProvider {
    return NewDefaultSQLQueryConstructorProvider(&CustomDialect{})
}
```

## 扩展 - 自定义 Condition

如果需要自定义条件（例如 FULL TEXT SEARCH），可以实现 Condition 接口：

```go
type FullTextCondition struct {
    Field string
    Query string
}

func (c *FullTextCondition) Type() string {
    return "fulltext"
}

func (c *FullTextCondition) Translate(translator ConditionTranslator) (string, []interface{}, error) {
    return translator.TranslateCondition(c)
}

// 在 Translator 中处理
func (t *DefaultSQLTranslator) translateFullTextCondition(cond *FullTextCondition) (string, []interface{}, error) {
    sql := "MATCH(" + t.dialect.QuoteIdentifier(cond.Field) + ") AGAINST(? IN BOOLEAN MODE)"
    return sql, []interface{}{cond.Query}, nil
}
```

## 设计原则

### 1. 分离关注点
- **顶层**：用户操作，不关心数据库细节
- **中层**：能力声明和路由，决定如何转义
- **底层**：具体实现，生成数据库特定的代码

### 2. 可插拔性
- 新数据库只需实现 SQLDialect 接口
- 无需修改顶层 API
- 用户代码保持不变

### 3. 性能考虑
- 延迟 SQL 生成（只在 Build() 时生成）
- 支持查询计划分析和索引提示
- 可在底层实现优化

### 4. 原生查询支持
- 不强制将所有查询转换为通用形式
- 支持数据库原生语言（Cypher、聚合管道等）
- 通过 GetNativeBuilder() 暴露底层实现

## 版本演进

- **v0.4.1** - 三层架构基础、SQL 实现
- **v0.4.2** - 关系查询支持（JOIN、Includes）
- **v0.5.0** - Neo4j Cypher 支持（验证架构）
- **v0.5.1+** - 其他 Adapter（MongoDB、SQL Server 等）
- **v1.0.0** - 完整生产就绪

## 与 v0.4.0 Migration 的结合

QueryConstructor 与 v0.4.0 的 Migration 工具无缝协作：

```go
// Migration 中使用查询构造器
func (m *SchemaMigration) Down(ctx context.Context, repo *Repository) error {
    // 查询要清理的数据
    provider := repo.GetAdapter().GetQueryBuilderProvider()
    qc := provider.NewQueryConstructor(someSchema)
    qc.Where(Eq("status", "archived"))
    
    sql, args, _ := qc.Build(ctx)
    repo.Exec(ctx, "DELETE FROM ... WHERE "+sql, args...)
}
```

## 与 v0.3.1 Changeset 的结合

QueryConstructor 也可以与 Changeset 结合用于复杂的数据操作：

```go
// 查询需要更新的记录
provider := repo.GetAdapter().GetQueryBuilderProvider()
qc := provider.NewQueryConstructor(userSchema)
qc.Where(Eq("status", "pending")).Select("id", "data")
sql, args, _ := qc.Build(ctx)
rows, _ := repo.Query(ctx, sql, args...)

// 使用 Changeset 进行验证和更新
for rows.Next() {
    var id int64
    var data string
    rows.Scan(&id, &data)
    
    cs := db.NewChangeset(userSchema)
    cs.PutChange("status", "active")
    cs.Validate()
    
    // 使用 QueryBuilder（Changeset 操作）
    qb := db.NewQueryBuilder(userSchema, repo)
    qb.UpdateByID(id, cs)
}
```

## 测试

v0.4.1 包含完整的测试覆盖：

```bash
# 运行所有查询构造器测试
go test -v -run TestSQLQueryConstructor

# 运行条件构造器测试
go test -v -run TestConditionBuilders

# 运行 SQL 方言测试
go test -v -run TestSQLDialects

# 运行查询提供者测试
go test -v -run TestQueryConstructorProvider
```

## 性能考虑

### 查询构造和缓存

由于 QueryConstructor 的链式 API，每次调用都会修改内部状态。如果需要重复使用相同的查询结构，建议提前构建：

```go
// ✗ 不高效 - 每次循环都重新构建
for i := 0; i < 1000; i++ {
    qc := provider.NewQueryConstructor(schema)
    qc.Where(Eq("status", "active")).Limit(10)
    sql, args, _ := qc.Build(ctx)
    repo.Query(ctx, sql, args...)
}

// ✓ 高效 - 构建一次，重复使用
baseQC := provider.NewQueryConstructor(schema)
baseQC.Where(Eq("status", "active")).Limit(10)
sql, args, _ := baseQC.Build(ctx)

for i := 0; i < 1000; i++ {
    repo.Query(ctx, sql, args...)
}
```

## 常见问题

### Q: 能否使用原生 SQL？
A: 可以。通过 Repository.Query() 或 Repository.Exec() 直接执行原生 SQL。QueryConstructor 只是便利工具，不是强制要求。

### Q: 如何处理复杂的 JOIN？
A: v0.4.1 还不支持 JOIN，但提供了基础架构。v0.4.2 会添加 JOIN 支持。

### Q: 能否添加自定义条件？
A: 可以。实现 Condition 接口并在自定义 ConditionTranslator 中处理。

### Q: 如何支持新数据库？
A: 实现 SQLDialect 接口或直接实现 QueryConstructor 接口。

## 总结

Query Constructor 三层架构为 EIT-DB 提供了：

1. ✅ 清晰的架构分离
2. ✅ 灵活的扩展点
3. ✅ 类型安全的 API
4. ✅ 多数据库支持基础
5. ✅ 原生查询语言支持

这为后续的功能（关系查询、其他 Adapter、高级优化等）奠定了坚实的基础。
