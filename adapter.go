package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

// Adapter 定义通用的数据库适配器接口 (参考 Ecto 设计)
// 每个数据库实现都必须满足这个接口
type Adapter interface {
	// 连接管理
	Connect(ctx context.Context, config *Config) error
	Close() error
	Ping(ctx context.Context) error

	// 事务管理
	Begin(ctx context.Context, opts ...interface{}) (Tx, error)

	// 查询接口 (用于 SELECT)
	Query(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) *sql.Row

	// 执行接口 (用于 INSERT/UPDATE/DELETE)
	Exec(ctx context.Context, sql string, args ...interface{}) (sql.Result, error)

	// 获取底层连接 (用于 GORM 等高级操作)
	GetRawConn() interface{}
}

// Tx 定义事务接口
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	
	// 事务中的查询和执行
	Query(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) *sql.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (sql.Result, error)
}

// Config 数据库配置结构 (参考 Ecto 的 Repo 配置)
type Config struct {
	// 适配器类型: "sqlite" | "postgres" | "mysql" | "sqlserver"
	Adapter string `json:"adapter" yaml:"adapter"`

	// SQLite 特定配置
	Database string `json:"database" yaml:"database"` // 数据库文件路径或数据库名

	// PostgreSQL/MySQL/SQL Server 通用配置
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`

	// PostgreSQL 特定配置
	SSLMode string `json:"ssl_mode" yaml:"ssl_mode"`

	// 连接池配置
	Pool *PoolConfig `json:"pool" yaml:"pool"`

	// 其他参数 (可选的适配器特定参数)
	Options map[string]interface{} `json:"options" yaml:"options"`
}

// PoolConfig 连接池配置 (参考 Ecto 的设计)
type PoolConfig struct {
	MaxConnections int `json:"max_connections" yaml:"max_connections"`
	MinConnections int `json:"min_connections" yaml:"min_connections"`
	ConnectTimeout int `json:"connect_timeout" yaml:"connect_timeout"` // 秒
	IdleTimeout    int `json:"idle_timeout" yaml:"idle_timeout"`       // 秒
	MaxLifetime    int `json:"max_lifetime" yaml:"max_lifetime"`       // 秒
}

// AdapterFactory 适配器工厂接口
type AdapterFactory interface {
	Name() string
	Create(config *Config) (Adapter, error)
}

// Repository 数据库仓储对象 (类似 Ecto.Repo)
type Repository struct {
	adapter Adapter
	mu      sync.RWMutex
}

// 全局适配器工厂注册表
var (
	adapterFactories = make(map[string]AdapterFactory)
	factoriesMutex   sync.RWMutex
)

// RegisterAdapter 注册适配器工厂
func RegisterAdapter(factory AdapterFactory) {
	factoriesMutex.Lock()
	defer factoriesMutex.Unlock()
	adapterFactories[factory.Name()] = factory
}

// NewRepository 创建新的仓储实例 (通过配置注入)
func NewRepository(config *Config) (*Repository, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.Adapter == "" {
		return nil, fmt.Errorf("adapter type must be specified")
	}

	// 从工厂注册表中获取适配器工厂
	factoriesMutex.RLock()
	factory, ok := adapterFactories[config.Adapter]
	factoriesMutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unsupported adapter: %s", config.Adapter)
	}

	// 使用工厂创建适配器
	adapter, err := factory.Create(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	return &Repository{adapter: adapter}, nil
}

// Connect 连接数据库
func (r *Repository) Connect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.adapter == nil {
		return fmt.Errorf("adapter is not initialized")
	}
	return r.adapter.Connect(ctx, nil)
}

// Close 关闭数据库连接
func (r *Repository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.adapter == nil {
		return nil
	}
	return r.adapter.Close()
}

// Ping 测试数据库连接
func (r *Repository) Ping(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.adapter == nil {
		return fmt.Errorf("adapter is not initialized")
	}
	return r.adapter.Ping(ctx)
}

// Query 执行查询
func (r *Repository) Query(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.adapter == nil {
		return nil, fmt.Errorf("adapter is not initialized")
	}
	return r.adapter.Query(ctx, sql, args...)
}

// QueryRow 执行单行查询
func (r *Repository) QueryRow(ctx context.Context, sql string, args ...interface{}) *sql.Row {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.adapter == nil {
		return nil
	}
	return r.adapter.QueryRow(ctx, sql, args...)
}

// Exec 执行操作
func (r *Repository) Exec(ctx context.Context, sql string, args ...interface{}) (sql.Result, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.adapter == nil {
		return nil, fmt.Errorf("adapter is not initialized")
	}
	return r.adapter.Exec(ctx, sql, args...)
}

// Begin 开始事务
func (r *Repository) Begin(ctx context.Context, opts ...interface{}) (Tx, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.adapter == nil {
		return nil, fmt.Errorf("adapter is not initialized")
	}
	return r.adapter.Begin(ctx, opts...)
}

// GetAdapter 获取底层适配器 (用于高级操作)
func (r *Repository) GetAdapter() Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.adapter
}
