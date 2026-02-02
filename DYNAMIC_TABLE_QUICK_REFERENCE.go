package db

/*

## 🚀 动态建表 - 快速开始指南

本文件包含动态建表功能的速查参考。

### 1. 基础概念

动态建表是指在应用运行时根据条件自动创建新表的功能：

- 场景：SaaS 多租户（每个租户一个表）、分表分库、日志分表等
- PostgreSQL：使用触发器（Trigger）+ 存储函数（Stored Function）
- MySQL/SQLite：使用 GORM Hook（AfterCreate）

### 2. 三步快速上手

#### 步骤1：定义表配置

```go
config := db.NewDynamicTableConfig("project_tasks").           // 表名前缀
    WithDescription("项目的任务表").
    WithParentTable("projects", "").                           // 监听 projects 表
    WithStrategy("auto").                                       // 自动创建
    AddField(db.NewDynamicTableField("id", db.TypeInteger).
        AsPrimaryKey().WithAutoinc()).
    AddField(db.NewDynamicTableField("title", db.TypeString).
        AsNotNull().WithIndex()).
    AddField(db.NewDynamicTableField("created_at", db.TypeTime).
        AsNotNull())
```

#### 步骤2：注册配置

```go
// 对于 PostgreSQL
pgAdapter := repo.adapter.(*db.PostgreSQLAdapter)
hook := db.NewPostgreSQLDynamicTableHook(pgAdapter)

// 对于 MySQL
mysqlAdapter := repo.adapter.(*db.MySQLAdapter)
hook := db.NewMySQLDynamicTableHook(mysqlAdapter)

// 对于 SQLite
sqliteAdapter := repo.adapter.(*db.SQLiteAdapter)
hook := db.NewSQLiteDynamicTableHook(sqliteAdapter)

// 注册配置
hook.RegisterDynamicTable(ctx, config)
```

#### 步骤3：使用

```go
// 自动创建：插入父表记录时自动创建对应的表
// project_tasks_1, project_tasks_2, ...

// 查询已创建的表
tables, _ := hook.ListCreatedDynamicTables(ctx, "project_tasks")

// 操作动态表
gormDB := repo.GetGormDB()
gormDB.Table("project_tasks_1").Create(&taskData)
```

### 3. 字段类型速查

| Go 类型 | PostgreSQL | MySQL | SQLite |
|--------|-----------|-------|--------|
| TypeString | VARCHAR(255) | VARCHAR(255) | TEXT |
| TypeInteger | INTEGER | INT | INTEGER |
| TypeFloat | FLOAT | FLOAT | REAL |
| TypeBoolean | BOOLEAN | TINYINT(1) | INTEGER |
| TypeTime | TIMESTAMP | DATETIME | TEXT |
| TypeBinary | BYTEA | LONGBLOB | BLOB |
| TypeDecimal | DECIMAL(18,2) | DECIMAL(18,2) | REAL |
| TypeJSON | JSONB | JSON | TEXT |
| TypeArray | TEXT[] | TEXT | TEXT |

### 4. 字段链式方法

```go
db.NewDynamicTableField("email", db.TypeString).
    AsNotNull().              // NOT NULL
    WithIndex().              // 添加索引
    WithUnique().             // 唯一约束
    WithDefault("").          // 默认值
    WithDescription("邮箱")
```

### 5. 常用场景模板

#### 场景A：SaaS 项目隔离（每个项目一个表）

```go
config := db.NewDynamicTableConfig("project_records").
    WithParentTable("projects", "").
    WithStrategy("auto").
    AddField(db.NewDynamicTableField("id", db.TypeInteger).
        AsPrimaryKey().WithAutoinc()).
    AddField(db.NewDynamicTableField("data", db.TypeJSON))

// 结果：project_records_1, project_records_2, ...
```

#### 场景B：条件触发（仅特定条件创建）

```go
config := db.NewDynamicTableConfig("premium_data").
    WithParentTable("users", "plan = 'premium'").  // 仅高级用户
    WithStrategy("auto").
    AddField(...)

// 仅当插入 plan='premium' 的用户时创建表
```

#### 场景C：手动创建（需要时才创建）

```go
config := db.NewDynamicTableConfig("temp_storage").
    WithStrategy("manual").  // 不自动创建
    AddField(...)

hook.RegisterDynamicTable(ctx, config)

// 需要时手动创建
tableName, _ := hook.CreateDynamicTable(ctx, "temp_storage", 
    map[string]interface{}{"id": 100})
// 结果：temp_storage_100
```

### 6. 常见操作

```go
// 列出所有配置
configs, _ := hook.ListDynamicTableConfigs(ctx)

// 获取特定配置
config, _ := hook.GetDynamicTableConfig(ctx, "project_tasks")

// 列出已创建的表
tables, _ := hook.ListCreatedDynamicTables(ctx, "project_tasks")

// 注销配置
hook.UnregisterDynamicTable(ctx, "project_tasks")
```

### 7. 错误处理

```go
if err := hook.RegisterDynamicTable(ctx, config); err != nil {
    // 配置无效或表创建失败
    log.Error("Failed to register dynamic table:", err)
}

tableName, err := hook.CreateDynamicTable(ctx, "config_name", params)
if err != nil {
    // 表可能已存在
    log.Error("Failed to create dynamic table:", err)
}
```

### 8. 选择适配器类型

```go
// 推荐使用类型断言来获取具体适配器
switch adapter := repo.adapter.(type) {
case *db.PostgreSQLAdapter:
    hook = db.NewPostgreSQLDynamicTableHook(adapter)
case *db.MySQLAdapter:
    hook = db.NewMySQLDynamicTableHook(adapter)
case *db.SQLiteAdapter:
    hook = db.NewSQLiteDynamicTableHook(adapter)
default:
    panic("Unsupported adapter type")
}
```

### 9. PostgreSQL 触发器方案优势

- ✅ 自动化：数据库层面自动执行，无需应用干预
- ✅ 一致性：原子事务，表创建和数据插入在同一事务
- ✅ 性能：触发器执行速度快，无应用开销
- ✅ 可靠性：不依赖应用状态

### 10. MySQL/SQLite Hook 方案优势

- ✅ 灵活：应用层实现，可自定义复杂逻辑
- ✅ 控制：完全在应用控制下，便于调试
- ✅ 兼容：不依赖数据库特定功能
- ✅ 监控：可添加日志、指标等

### 完整示例

```go
package main

import (
    "context"
    "log"
    db "github.com/eit-cms/eit-db"
)

func main() {
    ctx := context.Background()

    // 1. 连接数据库
    repo, err := db.NewRepository(ctx, &db.Config{
        Adapter:  "postgres",
        Host:     "localhost",
        Port:     5432,
        Username: "postgres",
        Database: "myapp",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer repo.Close()

    // 2. 定义动态表配置
    config := db.NewDynamicTableConfig("user_profiles").
        WithParentTable("users", "").
        WithStrategy("auto").
        AddField(db.NewDynamicTableField("id", db.TypeInteger).
            AsPrimaryKey().WithAutoinc()).
        AddField(db.NewDynamicTableField("bio", db.TypeString)).
        AddField(db.NewDynamicTableField("avatar_url", db.TypeString)).
        AddField(db.NewDynamicTableField("created_at", db.TypeTime).
            AsNotNull())

    // 3. 创建 hook
    pgAdapter := repo.adapter.(*db.PostgreSQLAdapter)
    hook := db.NewPostgreSQLDynamicTableHook(pgAdapter)

    // 4. 注册配置
    if err := hook.RegisterDynamicTable(ctx, config); err != nil {
        log.Fatal(err)
    }

    // 5. 现在每当插入用户时，user_profiles_* 表会自动创建

    // 6. 查询已创建的表
    tables, _ := hook.ListCreatedDynamicTables(ctx, "user_profiles")
    log.Printf("Created tables: %v", tables)
}
```

### 常见问题速答

**Q: 表名如何生成？**  
A: 默认为 `{配置表名}_{id}`，例如 `project_tasks_1`

**Q: 如何修改命名规则？**  
A: 扩展 Hook 实现，重写 `generateTableName` 方法

**Q: 已创建的表何时删除？**  
A: 不自动删除，需要手动通过 SQL 删除或定期清理任务

**Q: 是否支持外键约束？**  
A: 支持，但跨动态表的外键需要谨慎处理

**Q: 性能如何？**  
A: PostgreSQL 触发器方案最优；MySQL/SQLite Hook 方案受应用逻辑影响

详见 DYNAMIC_TABLE.md 完整文档

*/
