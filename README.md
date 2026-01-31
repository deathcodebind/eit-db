# EIT-DB

一个受 Ecto (Elixir) 启发的 Go 数据库抽象层,提供类型安全的 Schema、Changeset 和跨数据库适配器支持。

## 特性

- **Schema 系统**: 声明式数据结构定义,支持验证器和默认值
- **Changeset**: 数据变更追踪和验证,类似 Ecto.Changeset
- **跨数据库适配器**: 支持 MySQL, PostgreSQL, SQLite
- **查询构建器**: 类型安全的查询接口
- **迁移系统**: 自动从 Schema 生成数据库迁移
- **GORM 集成**: 可与 GORM 无缝配合使用

## 安装

```bash
go get pathologyenigma/eit-db
```

## 快速开始

### 1. 定义 Schema

```go
import "pathologyenigma/eit-db"

func BuildUserSchema() db.Schema {
    schema := db.NewBaseSchema("users")
    
    schema.AddField(
        db.NewField("id", db.TypeInteger).
            PrimaryKey().
            Build(),
    )
    
    schema.AddField(
        db.NewField("email", db.TypeString).
            Null(false).
            Unique().
            Validate(&db.EmailValidator{}).
            Build(),
    )
    
    schema.AddField(
        db.NewField("created_at", db.TypeTime).Build(),
    )
    
    return schema
}
```

### 2. 使用 Changeset 验证数据

```go
schema := BuildUserSchema()
data := map[string]interface{}{
    "email": "user@example.com",
    "created_at": time.Now(),
}

cs := db.NewChangeset(schema)
cs.Cast(data).Validate()

if !cs.IsValid() {
    for field, errors := range cs.Errors() {
        fmt.Printf("%s: %v\n", field, errors)
    }
}
```

### 3. 查询构建器

```go
// 初始化适配器
repo, _ := db.InitFromConfig("./config/database.yaml")

// 构建查询
qb := db.NewQueryBuilder(schema, repo)
result := qb.Query("email = ?", "user@example.com")

// 插入数据
cs := db.NewChangeset(schema).Cast(data).Validate()
qb.Insert(cs)

// 更新数据
updates := map[string]interface{}{"email": "new@example.com"}
cs := db.NewChangeset(schema).Cast(updates)
qb.Update(cs, "id = ?", userID)
```

### 4. 数据库迁移

```go
// 自动从 Schema 生成迁移
schemas := []db.Schema{BuildUserSchema(), BuildPostSchema()}
migrator := db.NewMigrator(repo)
migrator.AutoMigrate(schemas)
```

## 架构

EIT-DB 采用三层架构:

1. **Schema 层**: 定义数据结构和验证规则
2. **Changeset 层**: 管理数据变更和验证
3. **Adapter 层**: 抽象不同数据库的操作

这种设计使得你可以:
- 在不同数据库间轻松切换
- 在业务层使用统一的 API
- 轻松添加自定义验证器
- 保持代码的可测试性

## 支持的数据库

- MySQL 5.7+
- PostgreSQL 10+
- SQLite 3+

## 文档

详细文档请查看 [docs](./docs) 目录:

- Schema 定义指南
- Changeset 使用指南
- 查询构建器 API
- 自定义适配器开发

## License

MIT License
