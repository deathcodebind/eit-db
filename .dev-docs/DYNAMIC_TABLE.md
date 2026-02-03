markdown# 动态建表功能文档

## 📋 概述

动态建表（Dynamic Table Creation）是 eit-db 提供的一个强大功能，用于支持需要在运行时动态创建表的场景。

### 主要使用场景

1. **SaaS 多租户系统** - 为每个客户/项目创建独立的数据表
2. **CMS 自定义字段** - 为每个自定义字段集合创建专属表
3. **电商订单分表** - 为每个店铺/仓库创建独立的订单表
4. **日志系统分表** - 为每个应用/服务创建独立的日志表
5. **时间序列数据** - 为每个时间段创建新表（如按月分表）

## 🏗️ 架构设计

### 核心接口

```go
type DynamicTableHook interface {
    // 注册动态表配置
    RegisterDynamicTable(ctx context.Context, config *DynamicTableConfig) error

    // 注销动态表配置
    UnregisterDynamicTable(ctx context.Context, configName string) error

    // 列出所有已注册的配置
    ListDynamicTableConfigs(ctx context.Context) ([]*DynamicTableConfig, error)

    // 获取特定配置
    GetDynamicTableConfig(ctx context.Context, configName string) (*DynamicTableConfig, error)

    // 手动创建动态表
    CreateDynamicTable(ctx context.Context, configName string, 
        params map[string]interface{}) (string, error)

    // 列出已创建的表
    ListCreatedDynamicTables(ctx context.Context, configName string) ([]string, error)
}
```

### 三层实现架构

#### PostgreSQL（触发器模式）

优势：
- 原生支持触发器（Trigger）和存储函数（Stored Function）
- 完全由数据库负责，性能最优
- 事务安全性最强
- 支持复杂的触发条件

实现方式：
```
配置注册 → 创建 PL/pgSQL 函数 → 创建 AFTER INSERT 触发器
          ↓
      数据插入 → 触发器自动执行 → 函数创建新表
```

#### MySQL 和 SQLite（GORM Hook 模式）

优势：
- GORM hook 在应用层捕获事件
- 与应用逻辑紧密集成
- 可灵活实现复杂的业务逻辑

实现方式：
```
配置注册 → 向 GORM 注册 AfterCreate Hook
          ↓
      数据插入 → Hook 回调触发 → 检查条件并创建表
```

## 🚀 快速开始

### 第一步：定义动态表配置

```go
import "github.com/eit-cms/eit-db"

// 为每个项目创建独立的内容表
projectContentConfig := db.NewDynamicTableConfig("project_contents").
    WithDescription("项目的内容存储表").
    WithParentTable("projects", "").  // 监听 projects 表的插入
    WithStrategy("auto").              // 自动创建
    AddField(
        db.NewDynamicTableField("id", db.TypeInteger).
            AsPrimaryKey().
            WithAutoinc(),
    ).
    AddField(
        db.NewDynamicTableField("title", db.TypeString).
            AsNotNull().
            WithIndex(),
    ).
    AddField(
        db.NewDynamicTableField("content", db.TypeString).AsNotNull(),
    ).
    AddField(
        db.NewDynamicTableField("created_at", db.TypeTime).AsNotNull(),
    )
```

### 第二步：初始化数据库连接

```go
repo, err := db.NewRepository(ctx, &db.Config{
    Adapter:   "postgres",  // 或 "mysql" / "sqlite"
    Host:      "localhost",
    Port:      5432,
    Username:  "postgres",
    Password:  "password",
    Database:  "myapp",
})
if err != nil {
    panic(err)
}
defer repo.Close()
```

### 第三步：创建 Hook 并注册配置

```go
// 获取适配器
adapter := repo.adapter.(*db.PostgreSQLAdapter)

// 创建 hook（根据数据库类型选择）
hook := db.NewPostgreSQLDynamicTableHook(adapter)

// 注册配置
if err := hook.RegisterDynamicTable(ctx, projectContentConfig); err != nil {
    panic(err)
}
```

### 第四步：自动创建表

```go
// 现在，每当向 projects 表插入新记录时，都会自动创建对应的表：
// - project_contents_1
// - project_contents_2
// ...

gormDB := repo.GetGormDB()

type Project struct {
    ID   int
    Name string
}

// 插入项目
if err := gormDB.Create(&Project{ID: 1, Name: "Project 1"}).Error; err != nil {
    panic(err)
}
// 此时 project_contents_1 表已自动创建
```

## 📖 详细使用指南

### 触发条件配置

#### 无条件触发（所有插入都触发）

```go
config := db.NewDynamicTableConfig("my_table").
    WithParentTable("parent_table", "").  // 空字符串表示无条件
    WithStrategy("auto")
```

#### 条件触发（仅特定情况触发）

**PostgreSQL 示例：**
```go
config := db.NewDynamicTableConfig("shop_orders").
    WithParentTable("shops", "status = 'active'").  // 仅活跃店铺
    WithStrategy("auto")
```

**MySQL/SQLite 示例：**
```go
// MySQL/SQLite 的 hook 会在 handleAfterCreateCallback 中检查条件
config := db.NewDynamicTableConfig("shop_orders").
    WithParentTable("shops", "status = 'active'").
    WithStrategy("auto")
```

### 创建策略

#### 自动策略（Auto Strategy）

表在父表插入记录时自动创建：

```go
config := db.NewDynamicTableConfig("my_table").
    WithStrategy("auto").
    WithParentTable("parent_table", "")
```

#### 手动策略（Manual Strategy）

需要显式调用 `CreateDynamicTable` 来创建表：

```go
config := db.NewDynamicTableConfig("my_table").
    WithStrategy("manual").  // 不需要关联父表
    WithDescription("手动创建的表")

hook.RegisterDynamicTable(ctx, config)

// 后续手动创建
tableName, err := hook.CreateDynamicTable(ctx, "my_table", map[string]interface{}{
    "id": 123,
})
// 返回 "my_table_123"
```

### 字段配置链式方法

```go
field := db.NewDynamicTableField("email", db.TypeString).
    AsNotNull().              // NOT NULL 约束
    WithIndex().              // 创建索引
    WithUnique().             // 唯一约束
    WithDefault("").          // 默认值
    WithDescription("用户邮箱")

// 主键字段
pkField := db.NewDynamicTableField("id", db.TypeInteger).
    AsPrimaryKey().           // 设为主键
    WithAutoinc()             // 自增

config.AddField(pkField)
config.AddField(field)
```

### 字段类型映射

| FieldType | PostgreSQL | MySQL | SQLite |
|-----------|-----------|-------|--------|
| TypeString | VARCHAR(255) | VARCHAR(255) | TEXT |
| TypeInteger | INTEGER | INT | INTEGER |
| TypeFloat | FLOAT | FLOAT | REAL |
| TypeBoolean | BOOLEAN | TINYINT(1) | INTEGER |
| TypeTime | TIMESTAMP | DATETIME | TEXT |
| TypeBinary | BYTEA | LONGBLOB | BLOB |
| TypeDecimal | DECIMAL(18,2) | DECIMAL(18,2) | REAL |
| TypeJSON | JSONB | JSON | TEXT |
| TypeArray | TEXT[] | TEXT | TEXT |

## 💡 实际案例

### 案例 1：SaaS 项目管理系统

**需求：** 每个项目有独立的任务表

```go
taskTableConfig := db.NewDynamicTableConfig("project_tasks").
    WithDescription("项目的任务数据").
    WithParentTable("projects", "").
    WithStrategy("auto").
    AddField(
        db.NewDynamicTableField("id", db.TypeInteger).
            AsPrimaryKey().WithAutoinc(),
    ).
    AddField(
        db.NewDynamicTableField("title", db.TypeString).
            AsNotNull().WithIndex(),
    ).
    AddField(
        db.NewDynamicTableField("status", db.TypeString).
            WithDefault("todo"),
    ).
    AddField(
        db.NewDynamicTableField("assigned_to", db.TypeInteger).
            WithIndex(),
    ).
    AddField(
        db.NewDynamicTableField("created_at", db.TypeTime).AsNotNull(),
    )

// 当插入项目时：
// projects 表插入 ID=1 → project_tasks_1 自动创建
// projects 表插入 ID=2 → project_tasks_2 自动创建
```

### 案例 2：电商店铺订单系统

**需求：** 每个店铺维护独立的订单历史表

```go
orderConfig := db.NewDynamicTableConfig("shop_orders_history").
    WithParentTable("shops", "type = 'premium'").
    WithStrategy("auto").
    AddField(
        db.NewDynamicTableField("id", db.TypeInteger).
            AsPrimaryKey().WithAutoinc(),
    ).
    AddField(
        db.NewDynamicTableField("order_id", db.TypeString).
            WithUnique(),
    ).
    AddField(
        db.NewDynamicTableField("customer_id", db.TypeInteger).
            WithIndex(),
    ).
    AddField(
        db.NewDynamicTableField("amount", db.TypeDecimal),
    ).
    AddField(
        db.NewDynamicTableField("status", db.TypeString),
    ).
    AddField(
        db.NewDynamicTableField("created_at", db.TypeTime).AsNotNull(),
    )

// 表命名规则：shop_orders_history_1, shop_orders_history_2, ...
```

### 案例 3：应用日志分表系统

**需求：** 每个应用模块有独立的日志表

```go
logConfig := db.NewDynamicTableConfig("app_logs").
    WithStrategy("manual").  // 手动创建，用于事先初始化
    AddField(
        db.NewDynamicTableField("id", db.TypeInteger).
            AsPrimaryKey().WithAutoinc(),
    ).
    AddField(
        db.NewDynamicTableField("level", db.TypeString).
            WithIndex(),
    ).
    AddField(
        db.NewDynamicTableField("message", db.TypeString),
    ).
    AddField(
        db.NewDynamicTableField("context", db.TypeJSON),
    ).
    AddField(
        db.NewDynamicTableField("created_at", db.TypeTime).AsNotNull(),
    )

// 在应用启动时：
hook.RegisterDynamicTable(ctx, logConfig)

// 为每个模块创建表
modules := []string{"auth", "api", "admin", "scheduler"}
for i, module := range modules {
    tableName, err := hook.CreateDynamicTable(ctx, "app_logs", 
        map[string]interface{}{"id": i + 1})
    // 创建：app_logs_1, app_logs_2, app_logs_3, app_logs_4
}
```

## ⚙️ 高级功能

### 列出所有动态表配置

```go
configs, err := hook.ListDynamicTableConfigs(ctx)
if err != nil {
    panic(err)
}

for _, config := range configs {
    fmt.Printf("表: %s, 描述: %s, 策略: %s\n", 
        config.TableName, config.Description, config.Strategy)
}
```

### 查询已创建的表

```go
tables, err := hook.ListCreatedDynamicTables(ctx, "project_tasks")
if err != nil {
    panic(err)
}

for _, table := range tables {
    fmt.Println(table)  // project_tasks_1, project_tasks_2, ...
}
```

### 操作动态表中的数据

```go
gormDB := repo.GetGormDB()

type TaskRecord struct {
    ID    int
    Title string
    Status string
}

// 向项目1的任务表插入数据
task := TaskRecord{Title: "Task 1", Status: "todo"}
if err := gormDB.Table("project_tasks_1").Create(&task).Error; err != nil {
    panic(err)
}

// 查询项目1的任务
var tasks []TaskRecord
if err := gormDB.Table("project_tasks_1").Find(&tasks).Error; err != nil {
    panic(err)
}
```

### 注销配置（删除表和触发器）

```go
// 仅删除配置和触发器，已创建的表保留（可选手动删除）
if err := hook.UnregisterDynamicTable(ctx, "project_tasks"); err != nil {
    panic(err)
}
```

## 🔒 最佳实践

### 1. 合理设计表名

使用有意义的前缀和统一的命名规则：
```
✅ 好: project_tasks_1, project_tasks_2
❌ 差: t1, t2, table_1
```

### 2. 避免过度分表

不是所有场景都适合动态建表：
```
✅ 适合: 租户级别的隔离、大规模数据的分表
❌ 不适合: 每条记录一个表、字段数量少的情况
```

### 3. 保持字段一致

所有同系列的动态表应有相同的字段结构：
```go
// ✅ 所有 project_tasks_* 表结构相同
// ❌ 避免 project_tasks_1 和 project_tasks_2 字段不同
```

### 4. 及时清理过期表

定期检查并删除不再需要的表：
```go
// 获取所有已创建的表
tables, _ := hook.ListCreatedDynamicTables(ctx, "project_tasks")

// 检查是否关联的项目仍存在
for _, table := range tables {
    // ... 检查逻辑 ...
    // 删除对应表：DROP TABLE IF EXISTS table_name
}
```

### 5. 考虑索引策略

在创建表时为常查询字段添加索引：
```go
field.AddField(
    db.NewDynamicTableField("user_id", db.TypeInteger).
        WithIndex(),  // 频繁查询的字段加索引
)
```

## 🐛 常见问题

**Q: 为什么 PostgreSQL 使用触发器而其他数据库使用 Hook？**

A: PostgreSQL 的触发器和存储过程功能强大且高效，完全由数据库负责管理。MySQL 和 SQLite 的此类功能较弱，使用应用层 Hook 更灵活可控。

**Q: 如何确保表创建的事务安全性？**

A: PostgreSQL 的触发器在同一事务中执行，天然事务安全。MySQL/SQLite 的 Hook 也在事务中执行，但需要确保数据库支持 DDL 在事务中（MySQL 8.0+ 支持原子 DDL）。

**Q: 已创建的表如何处理数据迁移？**

A: 可使用 `ListCreatedDynamicTables` 获取所有表名，然后批量执行迁移语句。

**Q: 动态表支持外键约束吗？**

A: 支持，在定义字段时可添加索引。但跨越不同动态表的外键约束需要谨慎处理。

## 📝 完整示例代码

参见 [dynamic_table_examples.go](dynamic_table_examples.go)
