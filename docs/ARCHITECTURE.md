# EIT-DB 架构文档

> 本文为当前架构与 v1.0.0 路线图目标的一致性说明，描述现状、设计原则与演进方向。

## 1. 设计目标（与 v1.0.0 路线图对齐）

- **零 ORM 泄露**：用户只通过 Repository / Schema / Changeset 交互，底层 ORM 不对外暴露。
- **跨数据库一致性**：通过特性声明与派发机制，统一能力差异与降级策略。
- **可演进的查询构造器**：三层架构支撑多 SQL 方言与未来非 SQL（如搜索引擎）。
- **多适配器协同**：主库 + 次级搜索/分析/缓存适配器的可扩展路径（路线图 v0.9+）。
- **可观测性与测试覆盖**：功能特性与数据库能力可验证、可对比、可回归。

## 2. 核心分层

### 2.1 Domain 层（Schema / Changeset / Repository）

- **Schema**：声明式字段、约束、验证规则；支持手动定义与 Go 结构体反射生成。
- **Changeset**：统一的变更跟踪与验证入口（最终目标：所有 CRUD 基于 Changeset）。
- **Repository**：统一数据访问入口，屏蔽具体数据库差异与实现细节。

### 2.2 Adapter 层（数据库适配与能力声明）

- **Adapter 接口**：统一 `Connect / Query / Exec / Transaction` 等行为。
- **特性表**：
  - `DatabaseFeatures`：数据库原生能力声明（如 JSON、数组、生成列）。
  - `QueryFeatures`：查询能力声明与降级策略（如 FULL OUTER JOIN、CTE）。
- **派发与降级**：根据特性表路由能力，必要时在应用层降级。

#### 功能派发流程（示意）

```
┌─────────────┐      ┌──────────────────┐      ┌────────────────────┐
│ Repository  │ ───► │ Feature Registry │ ───► │ Adapter Capability │
└─────────────┘      └──────────────────┘      └────────────────────┘
    │                        │                         │
    │                        │                         ▼
    │                        │               原生能力？是 → 直接执行
    │                        │                         │
    │                        │                         └─ 否 → 查找降级策略
    │                        │                                   │
    │                        │                                   ├─ application_layer
    │                        │                                   ├─ alternative_syntax
    │                        │                                   └─ none → 返回错误
    ▼
   SQL / Query API
```

#### 功能派发伪代码

```go
func DispatchFeature(adapter Adapter, feature string) (Strategy, error) {
  if adapter.GetDatabaseFeatures().HasFeature(feature) {
    return StrategyNative, nil
  }

  fallback := adapter.GetDatabaseFeatures().GetFallbackStrategy(feature)
  switch fallback {
  case FallbackApplicationLayer:
    return StrategyApplicationLayer, nil
  case FallbackCheckConstraint:
    return StrategyCheckConstraint, nil
  case FallbackDynamicTable:
    return StrategyDynamicTable, nil
  case FallbackNone:
    return StrategyNone, fmt.Errorf("feature not supported: %s", feature)
  default:
    return StrategyNone, fmt.Errorf("unknown fallback: %s", fallback)
  }
}
```

#### QueryFeatures 示例（查询能力派发）

**示例 1：MySQL 的 FULL OUTER JOIN**

```go
qf := db.GetQueryFeatures("mysql")
if qf.HasQueryFeature("full_outer_join") {
  // 直接生成 FULL OUTER JOIN
} else {
  switch qf.GetFallbackStrategy("full_outer_join") {
  case db.QueryFallbackMultiQuery:
    // 使用 LEFT JOIN UNION RIGHT JOIN 模拟
  case db.QueryFallbackApplicationLayer:
    // 查询后在应用层合并结果
  default:
    return fmt.Errorf("full outer join not supported")
  }
}
```

**示例 2：SQLite 的 JSON 路径查询**

```go
qf := db.GetQueryFeatures("sqlite")
if qf.HasQueryFeature("json_path") {
  // 使用 JSON_EXTRACT / json_extract
} else {
  // 降级：拉取 JSON 字段后在应用层解析
}
```

### 2.3 Query Constructor 三层架构

1. **用户 API 层**：链式条件、排序、分页等对外接口。
2. **Adapter 转义层**：各适配器定义能力与语法映射。
3. **执行层**：SQL / 方言 / 参数占位符实现。

该结构为未来接入非 SQL 数据源（搜索引擎、图数据库等）预留扩展点。

### 2.4 Migration 与 Schema 变更

- Schema-based 与 Raw SQL 双模式。
- 未来将与**数据版本控制**与**自动分表**能力联动（路线图 v0.8+）。

## 3. Roadmap 对齐情况

### 已达成（当前）

- ✅ Repository + Schema + Changeset 基础架构
- ✅ Query Constructor 三层架构（SQL 方言）
- ✅ DatabaseFeatures / QueryFeatures 能力声明与测试
- ✅ Adapter 级别的能力验证与回归测试体系
- ✅ 不再对外暴露 GORM

### 正在推进 / 计划中

- 🔄 Schema 字段类型扩展（数据库特化/降级/方言）
- 🔄 多适配器（Multi-Adapter）架构设计与实现（路线图 v0.9+）
- 🔄 数据版本控制与自动分表（路线图 v0.8+）
- 🔄 性能与可观测性完善（路线图 v0.10+）

## 4. 关键设计约束

- **API 稳定性优先**：对外 API 不暴露底层实现细节。
- **功能可验证**：所有特性声明必须可测试（后端与应用层测试区分）。
- **跨数据库一致性**：能力差异通过声明、派发与降级统一。

## 5. 与其他文档的关系

- Query Constructor 架构详解：请参考 .dev-docs/QUERY_CONSTRUCTOR_ARCHITECTURE.md
- Query Features 系统：请参考 .dev-docs/QUERY_FEATURES.md
- Adapter 注册与工作流：请参考 .dev-docs/ADAPTER_WORKFLOW.md
- v1.0.0 路线图：请参考 .dev-docs/v1.0.0_ROADMAP.md
