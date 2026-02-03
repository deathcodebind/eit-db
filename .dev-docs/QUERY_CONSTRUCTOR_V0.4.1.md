# Query Constructor - v0.4.1 查询构造器架构

## 概述

v0.4.1 引入了**三层分离架构**的查询构造器设计，为 eit-db 的跨数据库支持奠定基础。

```
┌─────────────────────────────────────────┐
│ 上层：用户操作 Layer (User-Facing API)  │
│ QueryConstructor 接口                   │
│ 流式 API: Where/Select/OrderBy/Limit    │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│ 中层：Adapter 转义 Layer (Translation)  │
│ QueryConstructorProvider 接口            │
│ - 声明支持的操作 (Capabilities)         │
│ - 提供方言特定的实现                     │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│ 底层：DB 特定执行 Layer (Backend)       │
│ SQLQueryConstructor + SQLDialect        │
│ - 生成方言特定的 SQL                     │
│ - 参数化查询防止 SQL 注入                │
└─────────────────────────────────────────┘
```

## 核心接口

### 顶层 API - QueryConstructor 接口

用户与之交互的接口，完全与数据库无关：

```go
type QueryConstructor interface {
	// 条件查询
	Where(condition Condition) QueryConstructor
	WhereAll(conditions ...Condition) QueryConstructor  // AND
	WhereAny(conditions ...Condition) QueryConstructor  // OR
	
	// 字段选择
	Select(fields ...string) QueryConstructor
	
	// 排序
	OrderBy(field string, direction string) QueryConstructor
	
	// 分页
	Limit(count int) QueryConstructor
	Offset(count int) QueryConstructor
	
	// 构建查询
	Build(ctx context.Context) (string, []interface{}, error)
	
	// 获取底层构造器
	GetNativeBuilder() interface{}
}
```

### 中层 - QueryConstructorProvider 接口

每个 Adapter 实现此接口，提供方言特定的构造器：

```go
type QueryConstructorProvider interface {
	// 创建新的查询构造器
	NewQueryConstructor(schema Schema) QueryConstructor
	
	// 获取此 Adapter 的查询能力声明
	GetCapabilities() *QueryBuilderCapabilities
}

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
	SupportsSelect   bool
	SupportsOrderBy  bool
	SupportsLimit    bool
	SupportsOffset   bool
	SupportsJoin     bool
	SupportsSubquery bool
	
	// 优化特性
	SupportsQueryPlan bool
	SupportsIndex     bool
	
	// 原生查询支持（如 Cypher）
	SupportsNativeQuery bool
	NativeQueryLang     string
}
```

### 底层 - SQLDialect 接口

定义 SQL 方言的实现细节：

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
	
	// 转换条件为 SQL（可选优化）
	TranslateCondition(condition Condition, argIndex *int) (string, []interface{}, error)
}
```

## 使用示例

### 基础用法

```go
// 创建 Provider（通常由 Adapter 提供）
provider := NewDefaultSQLQueryConstructorProvider(NewMySQLDialect())

// 获取 Schema
schema := buildUserSchema()

// 创建查询构造器
qc := provider.NewQueryConstructor(schema)

// 构建查询
qc.Where(Eq("status", "active")).
   Where(Gt("age", 18)).
   OrderBy("created_at", "DESC").
   Limit(10).
   Offset(5)

// 生成 SQL
sql, args, err := qc.Build(context.Background())
// sql: SELECT * FROM `users` WHERE `status` = ? AND `age` > ? ORDER BY `created_at` DESC LIMIT 10 OFFSET 5
// args: ["active", 18]
```

### 复杂条件

```go
// AND 条件组
qc.WhereAll(
    Eq("status", "active"),
    Gt("age", 18),
    Lt("age", 65),
)

// OR 条件组
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

// NOT 条件
qc.Where(Not(Eq("deleted", true)))
```

### 字段选择

```go
qc.Select("id", "name", "email")
// 生成：SELECT `id`, `name`, `email` FROM `users`

// 如果不指定，默认选择所有字段
qc2.Build(ctx)
// 生成：SELECT * FROM `users`
```

## 条件构造器

### 支持的操作符

| 函数 | SQL 操作符 | 示例 |
|------|-----------|------|
| `Eq` | `=` | `Eq("status", "active")` |
| `Ne` | `!=` | `Ne("status", "inactive")` |
| `Gt` | `>` | `Gt("age", 18)` |
| `Lt` | `<` | `Lt("age", 65)` |
| `Gte` | `>=` | `Gte("score", 80)` |
| `Lte` | `<=` | `Lte("score", 100)` |
| `In` | `IN` | `In("role", "admin", "moderator")` |
| `Between` | `BETWEEN` | `Between("age", 18, 65)` |
| `Like` | `LIKE` | `Like("name", "%John%")` |

### 逻辑操作符

```go
And(cond1, cond2, ...)    // 所有条件都要满足
Or(cond1, cond2, ...)     // 任意条件满足即可
Not(condition)             // 条件的否定
```

## SQL 方言支持

### MySQL

- **标识符引用**：`` ` `` (反引号)
- **参数占位符**：`?`
- **示例**：`` SELECT `id`, `name` FROM `users` WHERE `age` > ? ``

### PostgreSQL

- **标识符引用**：`"` (双引号)
- **参数占位符**：`$1`, `$2`, ...
- **示例**：`SELECT "id", "name" FROM "users" WHERE "age" > $1`
- **额外特性**：支持 JSON 操作、数组等

### SQLite

- **标识符引用**：`` ` `` (反引号)
- **参数占位符**：`?`
- **示例**：`` SELECT `id`, `name` FROM `users` WHERE `age` > ? ``

## 测试覆盖

v0.4.1 包含 20+ 单元测试，验证：

### ✅ 条件测试
- `TestSQLQueryConstructorBasicSelect` - 基础 SELECT
- `TestSQLQueryConstructorEqCondition` - 等于条件
- `TestSQLQueryConstructorComparisonOperators` - 所有比较操作符
- `TestSQLQueryConstructorInCondition` - IN 条件与参数验证
- `TestSQLQueryConstructorBetweenCondition` - BETWEEN 条件
- `TestSQLQueryConstructorLikeCondition` - LIKE 条件

### ✅ 逻辑操作测试
- `TestSQLQueryConstructorWhereAll` - AND 组合
- `TestSQLQueryConstructorWhereAny` - OR 组合

### ✅ 查询特性测试
- `TestSQLQueryConstructorOrderBy` - ORDER BY 多字段排序
- `TestSQLQueryConstructorLimitOffset` - LIMIT/OFFSET 组合
- `TestSQLQueryConstructorSelectColumns` - 字段选择

### ✅ 方言测试
- `TestSQLDialectQuoting` - 验证每个方言的引号和占位符
- 验证 MySQL 使用 backticks + ?
- 验证 PostgreSQL 使用 double quotes + $1
- 验证 SQLite 使用 backticks + ?

### ✅ 综合测试
- `TestSQLQueryConstructorCombined` - 复杂查询组合
- `TestQueryConstructorProvider` - Provider 功能验证

## 实现细节

### 参数化查询

所有条件值自动参数化，防止 SQL 注入：

```go
// 输入
qc.Where(Eq("email", "user@example.com"))

// 生成的 SQL 和参数
sql: "SELECT * FROM `users` WHERE `email` = ?"
args: ["user@example.com"]

// 参数由数据库驱动安全处理
```

### 参数索引管理

不同方言的参数占位符自动转换：

```go
// MySQL/SQLite
"WHERE age > ? AND status = ?"

// PostgreSQL
"WHERE age > $1 AND status = $2"

// 参数列表始终统一：[18, "active"]
```

### 条件翻译器

`ConditionTranslator` 接口处理条件到 SQL 的转换：

```go
type ConditionTranslator interface {
	TranslateCondition(condition Condition) (string, []interface{}, error)
	TranslateComposite(operator string, conditions []Condition) (string, []interface{}, error)
}
```

## v0.4.2 路线图

- ✅ 关系查询支持（JOIN、INCLUDE、PRELOAD）
- ✅ 关系查询的自动 JOIN 生成
- ✅ Relationship 架构整合

## v0.5.0 路线图

- 🔄 Neo4j Adapter 参考实现
- 🔄 CypherQueryBuilder（Cypher 查询语言）
- 🔄 验证三层架构在非 SQL 数据库中的有效性

## 常见问题

### Q: 为什么需要三层架构？

**A**: 三层分离带来：
1. **清晰的关注点分离** - 用户 API、转义逻辑、方言实现完全独立
2. **易于扩展** - 新数据库只需实现底层构造器，上层 API 不变
3. **验证机制** - Neo4j 等非 SQL 数据库可验证架构的通用性
4. **优化灵活性** - 每个 Adapter 可提供方言特定的优化

### Q: 如何添加新的 SQL 方言？

**A**: 实现 `SQLDialect` 接口：

```go
type CustomDialect struct{}

func (d *CustomDialect) Name() string { return "custom" }
func (d *CustomDialect) QuoteIdentifier(name string) string { return "\"" + name + "\"" }
func (d *CustomDialect) QuoteValue(value interface{}) string { ... }
func (d *CustomDialect) GetPlaceholder(index int) string { return fmt.Sprintf("$%d", index) }
// ... 实现其他方法
```

### Q: 如何在 Adapter 中使用 QueryConstructor？

**A**: 在 Adapter 中实现 `GetQueryBuilderProvider()` 方法：

```go
func (a *MyAdapter) GetQueryBuilderProvider() QueryConstructorProvider {
	return NewDefaultSQLQueryConstructorProvider(NewMyCustomDialect())
}
```

## 性能考虑

- **编译时优化**：没有运行时反射，所有操作都是直接的方法调用
- **参数化查询**：使用数据库驱动的参数化查询，获得数据库端的缓存优化
- **零复制**：构造器使用指针接收者，避免不必要的结构体复制
- **流式 API**：每个链式调用都直接修改构造器状态，无中间对象创建
