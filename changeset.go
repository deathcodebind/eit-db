package db

import (
	"fmt"
	"sync"
)

// Changeset 代表对数据的变更（参考 Ecto.Changeset）
type Changeset struct {
	// 原始数据
	data map[string]interface{}
	
	// 变更的数据
	changes map[string]interface{}
	
	// 验证错误
	errors map[string][]string
	
	// 关联的模式
	schema Schema
	
	// 是否有效
	valid bool
	
	// 变更前的值（用于追踪）
	previousValues map[string]interface{}
	
	// 锁
	mu sync.RWMutex
}

// NewChangeset 创建新的 Changeset
func NewChangeset(schema Schema) *Changeset {
	return &Changeset{
		data:            make(map[string]interface{}),
		changes:         make(map[string]interface{}),
		errors:          make(map[string][]string),
		schema:          schema,
		valid:           true,
		previousValues:  make(map[string]interface{}),
	}
}

// FromMap 从 map 创建 Changeset
func FromMap(schema Schema, dataMap map[string]interface{}) *Changeset {
	cs := NewChangeset(schema)
	for k, v := range dataMap {
		cs.data[k] = v
	}
	return cs
}

// Cast 设置字段值（类似 Ecto 的 cast）
func (cs *Changeset) Cast(data map[string]interface{}) *Changeset {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for key, value := range data {
		field := cs.schema.GetField(key)
		if field == nil {
			continue // 忽略未定义的字段
		}

		// 保存原始值
		if oldValue, exists := cs.data[key]; exists {
			cs.previousValues[key] = oldValue
		}

		// 应用转换器
		transformedValue := value
		for _, transformer := range field.Transformers {
			transformed, err := transformer.Transform(transformedValue)
			if err != nil {
				cs.addError(key, fmt.Sprintf("转换器错误: %v", err))
				continue
			}
			transformedValue = transformed
		}

		// 类型转换
		convertedValue, err := ConvertValue(transformedValue, field.Type)
		if err != nil {
			cs.addError(key, fmt.Sprintf("类型转换失败: %v", err))
			continue
		}

		cs.changes[key] = convertedValue
		cs.data[key] = convertedValue
	}

	return cs
}

// Validate 验证 Changeset
func (cs *Changeset) Validate() *Changeset {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.errors = make(map[string][]string) // 清空之前的错误

	for _, field := range cs.schema.Fields() {
		value, exists := cs.data[field.Name]

		// 检查必填字段
		if !field.Null && (!exists || value == nil || value == "") {
			cs.addError(field.Name, "字段为必填项")
			cs.valid = false
			continue
		}

		// 应用验证器
		if exists && value != nil {
			for _, validator := range field.Validators {
				if err := validator.Validate(value); err != nil {
					cs.addError(field.Name, err.Error())
					cs.valid = false
				}
			}
		}
	}

	return cs
}

// ValidateChange 验证特定字段的变更
func (cs *Changeset) ValidateChange(fieldName string, validator Validator) *Changeset {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	value, exists := cs.changes[fieldName]
	if !exists {
		return cs
	}

	if err := validator.Validate(value); err != nil {
		cs.addError(fieldName, err.Error())
		cs.valid = false
	}

	return cs
}

// IsValid 检查 Changeset 是否有效
func (cs *Changeset) IsValid() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.valid && len(cs.errors) == 0
}

// Errors 获取所有错误
func (cs *Changeset) Errors() map[string][]string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.errors
}

// GetError 获取字段的错误
func (cs *Changeset) GetError(fieldName string) []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.errors[fieldName]
}

// Data 获取所有数据
func (cs *Changeset) Data() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	result := make(map[string]interface{})
	for k, v := range cs.data {
		result[k] = v
	}
	return result
}

// Changes 获取变更的数据
func (cs *Changeset) Changes() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	result := make(map[string]interface{})
	for k, v := range cs.changes {
		result[k] = v
	}
	return result
}

// Get 获取字段值
func (cs *Changeset) Get(fieldName string) interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.data[fieldName]
}

// GetChanged 获取变更的字段值
func (cs *Changeset) GetChanged(fieldName string) (interface{}, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	val, ok := cs.changes[fieldName]
	return val, ok
}

// GetPrevious 获取变更前的值
func (cs *Changeset) GetPrevious(fieldName string) interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.previousValues[fieldName]
}

// HasChanged 检查字段是否被修改
func (cs *Changeset) HasChanged(fieldName string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	_, ok := cs.changes[fieldName]
	return ok
}

// addError 添加验证错误
func (cs *Changeset) addError(fieldName string, message string) {
	if _, ok := cs.errors[fieldName]; !ok {
		cs.errors[fieldName] = make([]string, 0)
	}
	cs.errors[fieldName] = append(cs.errors[fieldName], message)
}

// PutChange 手动添加变更
func (cs *Changeset) PutChange(fieldName string, value interface{}) *Changeset {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	field := cs.schema.GetField(fieldName)
	if field == nil {
		return cs
	}

	// 保存原始值
	if oldValue, exists := cs.data[fieldName]; exists {
		cs.previousValues[fieldName] = oldValue
	}

	cs.changes[fieldName] = value
	cs.data[fieldName] = value

	return cs
}

// ClearError 清除错误
func (cs *Changeset) ClearError(fieldName string) *Changeset {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	delete(cs.errors, fieldName)
	if len(cs.errors) == 0 {
		cs.valid = true
	}

	return cs
}

// ForceChanges 强制所有字段为变更状态（用于插入操作）
func (cs *Changeset) ForceChanges() *Changeset {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for k, v := range cs.data {
		cs.changes[k] = v
	}

	return cs
}

// GetChangedFields 获取所有被修改的字段名列表
func (cs *Changeset) GetChangedFields() []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	fields := make([]string, 0, len(cs.changes))
	for k := range cs.changes {
		fields = append(fields, k)
	}
	return fields
}

// ErrorString 返回格式化的错误字符串
func (cs *Changeset) ErrorString() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if len(cs.errors) == 0 {
		return ""
	}

	errorStr := ""
	for field, messages := range cs.errors {
		errorStr += field + ": " + fmt.Sprintf("%v", messages) + "; "
	}
	return errorStr
}

// ToMap 转换为 map（用于数据库操作）
func (cs *Changeset) ToMap() map[string]interface{} {
	return cs.Changes()
}
