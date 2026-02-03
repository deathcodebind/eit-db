package db

import (
	"context"
	"reflect"
	"time"
)

// FieldType 字段类型定义
type FieldType string

const (
	TypeString    FieldType = "string"
	TypeInteger   FieldType = "integer"
	TypeFloat     FieldType = "float"
	TypeBoolean   FieldType = "boolean"
	TypeTime      FieldType = "time"
	TypeBinary    FieldType = "binary"
	TypeDecimal   FieldType = "decimal"
	TypeMap       FieldType = "map"
	TypeArray     FieldType = "array"
	TypeJSON      FieldType = "json"
)

// Field 定义模式中的字段
type Field struct {
	Name         string
	Type         FieldType
	Default      interface{}
	Null         bool
	Primary      bool
	Autoinc      bool
	Index        bool
	Unique       bool
	Validators   []Validator
	Transformers []Transformer
}

// Schema 定义数据模式接口 (参考 Ecto.Schema)
type Schema interface {
	// 获取模式名称（表名）
	TableName() string
	
	// 获取所有字段
	Fields() []*Field
	
	// 获取字段
	GetField(name string) *Field
	
	// 获取主键字段
	PrimaryKeyField() *Field
}

// BaseSchema 基础模式实现
type BaseSchema struct {
	tableName string
	fields    map[string]*Field
	fieldList []*Field
}

// NewBaseSchema 创建基础模式
func NewBaseSchema(tableName string) *BaseSchema {
	return &BaseSchema{
		tableName: tableName,
		fields:    make(map[string]*Field),
		fieldList: make([]*Field, 0),
	}
}

// TableName 返回表名
func (s *BaseSchema) TableName() string {
	return s.tableName
}

// AddField 添加字段
func (s *BaseSchema) AddField(field *Field) *BaseSchema {
	s.fields[field.Name] = field
	s.fieldList = append(s.fieldList, field)
	return s
}

// Fields 返回所有字段
func (s *BaseSchema) Fields() []*Field {
	return s.fieldList
}

// GetField 获取字段
func (s *BaseSchema) GetField(name string) *Field {
	return s.fields[name]
}

// PrimaryKeyField 获取主键字段
func (s *BaseSchema) PrimaryKeyField() *Field {
	for _, field := range s.fieldList {
		if field.Primary {
			return field
		}
	}
	return nil
}

// FieldBuilder 字段构造器
type FieldBuilder struct {
	field *Field
}

// NewField 创建新字段
func NewField(name string, fieldType FieldType) *FieldBuilder {
	return &FieldBuilder{
		field: &Field{
			Name:         name,
			Type:         fieldType,
			Validators:   make([]Validator, 0),
			Transformers: make([]Transformer, 0),
		},
	}
}

// Default 设置默认值
func (fb *FieldBuilder) Default(value interface{}) *FieldBuilder {
	fb.field.Default = value
	return fb
}

// Null 设置是否允许为空
func (fb *FieldBuilder) Null(allow bool) *FieldBuilder {
	fb.field.Null = allow
	return fb
}

// PrimaryKey 标记为主键
func (fb *FieldBuilder) PrimaryKey() *FieldBuilder {
	fb.field.Primary = true
	fb.field.Autoinc = true
	return fb
}

// Index 添加索引
func (fb *FieldBuilder) Index() *FieldBuilder {
	fb.field.Index = true
	return fb
}

// Unique 添加唯一约束
func (fb *FieldBuilder) Unique() *FieldBuilder {
	fb.field.Unique = true
	return fb
}

// Validate 添加验证器
func (fb *FieldBuilder) Validate(validator Validator) *FieldBuilder {
	fb.field.Validators = append(fb.field.Validators, validator)
	return fb
}

// Transform 添加转换器
func (fb *FieldBuilder) Transform(transformer Transformer) *FieldBuilder {
	fb.field.Transformers = append(fb.field.Transformers, transformer)
	return fb
}

// Build 构建字段
func (fb *FieldBuilder) Build() *Field {
	return fb.field
}

// Validator 验证器接口
type Validator interface {
	// 验证值，返回错误信息或 nil
	Validate(value interface{}) error
}

// Transformer 转换器接口
type Transformer interface {
	// 转换值
	Transform(value interface{}) (interface{}, error)
}

// 内置验证器

// RequiredValidator 必填验证器
type RequiredValidator struct{}

func (v *RequiredValidator) Validate(value interface{}) error {
	if value == nil || value == "" {
		return NewValidationError("required", "字段为必填项")
	}
	return nil
}

// LengthValidator 长度验证器
type LengthValidator struct {
	Min int
	Max int
}

func (v *LengthValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return NewValidationError("length", "字段类型必须为字符串")
	}
	
	len := len(str)
	if v.Min > 0 && len < v.Min {
		return NewValidationError("length", "字段长度不能小于 "+string(rune(v.Min)))
	}
	if v.Max > 0 && len > v.Max {
		return NewValidationError("length", "字段长度不能大于 "+string(rune(v.Max)))
	}
	return nil
}

// PatternValidator 正则验证器
type PatternValidator struct {
	Pattern string
}

func (v *PatternValidator) Validate(value interface{}) error {
	// 实现正则验证
	return nil
}

// UniqueValidator 唯一性验证器
type UniqueValidator struct {
	Schema Schema
	Field  string
}

func (v *UniqueValidator) Validate(value interface{}) error {
	// 需要从数据库查询
	return nil
}

// 内置转换器

// TrimTransformer 字符串修剪转换器
type TrimTransformer struct{}

func (t *TrimTransformer) Transform(value interface{}) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}
	// 修剪字符串
	return str, nil
}

// LowercaseTransformer 小写转换器
type LowercaseTransformer struct{}

func (t *LowercaseTransformer) Transform(value interface{}) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}
	// 转换为小写
	return str, nil
}

// TypeConversionError 类型转换错误
type TypeConversionError struct {
	From string
	To   string
}

func (e *TypeConversionError) Error() string {
	return "cannot convert " + e.From + " to " + e.To
}

// ConvertValue 值类型转换
func ConvertValue(value interface{}, targetType FieldType) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch targetType {
	case TypeString:
		return valueToString(value), nil
	case TypeInteger:
		return valueToInt64(value)
	case TypeFloat:
		return valueToFloat64(value)
	case TypeBoolean:
		return valueToBoolean(value)
	case TypeTime:
		return valueToTime(value)
	default:
		return value, nil
	}
}

func valueToString(value interface{}) interface{} {
	return value
}

func valueToInt64(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int(), nil
	case float32, float64:
		return int64(reflect.ValueOf(v).Float()), nil
	case string:
		// TODO: 实现字符串到 int64 的转换
		return nil, &TypeConversionError{From: "string", To: "int64"}
	default:
		return nil, &TypeConversionError{From: reflect.TypeOf(value).String(), To: "int64"}
	}
}

func valueToFloat64(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case float32, float64:
		return reflect.ValueOf(v).Float(), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), nil
	default:
		return nil, &TypeConversionError{From: reflect.TypeOf(value).String(), To: "float64"}
	}
}

func valueToBoolean(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return v == "true" || v == "1" || v == "yes", nil
	case int:
		return v != 0, nil
	default:
		return nil, &TypeConversionError{From: reflect.TypeOf(value).String(), To: "bool"}
	}
}

func valueToTime(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// 尝试解析时间字符串
		t, err := time.Parse(time.RFC3339, v)
		return t, err
	default:
		return nil, &TypeConversionError{From: reflect.TypeOf(value).String(), To: "time.Time"}
	}
}

// ValidationError 验证错误
type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Code + ": " + e.Message
}

// NewValidationError 创建验证错误
func NewValidationError(code, message string) *ValidationError {
	return &ValidationError{
		Code:    code,
		Message: message,
	}
}

// ==================== Schema Registry ====================

// SchemaRegistry Schema 注册表，便于查找和管理多个 Schema
type SchemaRegistry struct {
	schemas map[string]Schema
}

// NewSchemaRegistry 创建空的 Schema 注册表
func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		schemas: make(map[string]Schema),
	}
}

// Register 注册一个 Schema
func (r *SchemaRegistry) Register(name string, schema Schema) {
	r.schemas[name] = schema
}

// Get 获取指定名称的 Schema
func (r *SchemaRegistry) Get(name string) Schema {
	return r.schemas[name]
}

// GetAllSchemaNames 获取所有已注册的 Schema 名称
func (r *SchemaRegistry) GetAllSchemaNames() []string {
	names := make([]string, 0, len(r.schemas))
	for name := range r.schemas {
		names = append(names, name)
	}
	return names
}

// Timestamp 获取当前时间（用于 created_at/updated_at 字段）
func Timestamp() time.Time {
	return time.Now()
}

// ==================== Query Builder (v0.4.1) ====================

// QueryConstructor 查询构造器接口 - 顶层 API
// 用户通过此接口构建查询，具体实现由适配器提供
type QueryConstructor interface {
	// 条件查询
	Where(condition Condition) QueryConstructor
	
	// 多条件 AND 组合
	WhereAll(conditions ...Condition) QueryConstructor
	
	// 多条件 OR 组合
	WhereAny(conditions ...Condition) QueryConstructor
	
	// 字段选择
	Select(fields ...string) QueryConstructor
	
	// 排序
	OrderBy(field string, direction string) QueryConstructor // direction: "ASC" | "DESC"
	
	// 分页
	Limit(count int) QueryConstructor
	Offset(count int) QueryConstructor
	
	// 构建查询
	Build(ctx context.Context) (string, []interface{}, error)
	
	// 获取底层查询构造器（用于 Adapter 特定优化）
	GetNativeBuilder() interface{}
}

// Condition 条件接口 - 中层转义
// Adapter 实现此接口将条件转换为数据库特定的形式
type Condition interface {
	// 获取条件类型
	Type() string
	
	// 将条件转换为 SQL/Cypher/etc
	Translate(translator ConditionTranslator) (string, []interface{}, error)
}

// ConditionTranslator 条件转义器接口
// 由每个 Adapter 的 QueryConstructor 实现
type ConditionTranslator interface {
	TranslateCondition(condition Condition) (string, []interface{}, error)
	TranslateComposite(operator string, conditions []Condition) (string, []interface{}, error)
}

// ==================== 内置 Condition 实现 ====================

// SimpleCondition 简单条件（字段 操作符 值）
type SimpleCondition struct {
	Field    string
	Operator string // "eq", "ne", "gt", "lt", "gte", "lte", "in", "like", "between"
	Value    interface{}
}

func (c *SimpleCondition) Type() string {
	return "simple"
}

func (c *SimpleCondition) Translate(translator ConditionTranslator) (string, []interface{}, error) {
	return translator.TranslateCondition(c)
}

// CompositeCondition 复合条件（AND/OR）
type CompositeCondition struct {
	Operator   string        // "and" | "or"
	Conditions []Condition
}

func (c *CompositeCondition) Type() string {
	return "composite"
}

func (c *CompositeCondition) Translate(translator ConditionTranslator) (string, []interface{}, error) {
	return translator.TranslateComposite(c.Operator, c.Conditions)
}

// NotCondition 非条件
type NotCondition struct {
	Condition Condition
}

func (c *NotCondition) Type() string {
	return "not"
}

func (c *NotCondition) Translate(translator ConditionTranslator) (string, []interface{}, error) {
	innerSQL, args, err := c.Condition.Translate(translator)
	if err != nil {
		return "", nil, err
	}
	return "NOT (" + innerSQL + ")", args, nil
}

// ==================== Condition Builder (Fluent API) ====================

// ConditionBuilder 条件构造器 - 流式 API
type ConditionBuilder struct {
	field    string
	operator string
	value    interface{}
}

// Eq 等于条件
func Eq(field string, value interface{}) Condition {
	return &SimpleCondition{
		Field:    field,
		Operator: "eq",
		Value:    value,
	}
}

// Ne 不等于条件
func Ne(field string, value interface{}) Condition {
	return &SimpleCondition{
		Field:    field,
		Operator: "ne",
		Value:    value,
	}
}

// Gt 大于条件
func Gt(field string, value interface{}) Condition {
	return &SimpleCondition{
		Field:    field,
		Operator: "gt",
		Value:    value,
	}
}

// Lt 小于条件
func Lt(field string, value interface{}) Condition {
	return &SimpleCondition{
		Field:    field,
		Operator: "lt",
		Value:    value,
	}
}

// Gte 大于等于条件
func Gte(field string, value interface{}) Condition {
	return &SimpleCondition{
		Field:    field,
		Operator: "gte",
		Value:    value,
	}
}

// Lte 小于等于条件
func Lte(field string, value interface{}) Condition {
	return &SimpleCondition{
		Field:    field,
		Operator: "lte",
		Value:    value,
	}
}

// In IN 条件
func In(field string, values ...interface{}) Condition {
	return &SimpleCondition{
		Field:    field,
		Operator: "in",
		Value:    values,
	}
}

// Between BETWEEN 条件
func Between(field string, min, max interface{}) Condition {
	return &SimpleCondition{
		Field:    field,
		Operator: "between",
		Value:    []interface{}{min, max},
	}
}

// Like LIKE 条件（模糊匹配）
func Like(field string, pattern string) Condition {
	return &SimpleCondition{
		Field:    field,
		Operator: "like",
		Value:    pattern,
	}
}

// And AND 条件
func And(conditions ...Condition) Condition {
	return &CompositeCondition{
		Operator:   "and",
		Conditions: conditions,
	}
}

// Or OR 条件
func Or(conditions ...Condition) Condition {
	return &CompositeCondition{
		Operator:   "or",
		Conditions: conditions,
	}
}

// Not NOT 条件
func Not(condition Condition) Condition {
	return &NotCondition{
		Condition: condition,
	}
}
