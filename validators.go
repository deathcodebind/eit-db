package db

import (
	"fmt"
	"regexp"
)

// ==================== 常用验证器 ====================

// EmailValidator 邮箱格式验证器
type EmailValidator struct{}

func (v *EmailValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return NewValidationError("email", "邮箱必须是字符串")
	}

	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, str)
	if !matched {
		return NewValidationError("email", "邮箱格式不正确")
	}
	return nil
}

// MinLengthValidator 最小长度验证器
type MinLengthValidator struct {
	Length int
}

func (v *MinLengthValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return NewValidationError("min_length", "字段必须是字符串")
	}

	if len(str) < v.Length {
		return NewValidationError("min_length", fmt.Sprintf("字段长度不能小于 %d", v.Length))
	}
	return nil
}

// MaxLengthValidator 最大长度验证器
type MaxLengthValidator struct {
	Length int
}

func (v *MaxLengthValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return NewValidationError("max_length", "字段必须是字符串")
	}

	if len(str) > v.Length {
		return NewValidationError("max_length", fmt.Sprintf("字段长度不能大于 %d", v.Length))
	}
	return nil
}
