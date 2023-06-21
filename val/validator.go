package val

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidUsername = regexp.MustCompile("^[a-zA-Z0-9_]+$").MatchString
	isValidFullname = regexp.MustCompile("^[\\w\\s]+$").MatchString
)

// 验证string，最大最小长度限制
func ValidateString(value string, min int, max int) error {
	l := len(value)
	if l < min || l > max {
		return fmt.Errorf("只能包括 %d-%d 个字符", min, max)
	}
	return nil
}

// 验证用户名
func ValidateUsername(value string, min int, max int) error {
	// 验证长度
	err := ValidateString(value, min, max)
	if err != nil {
		return err
	}

	// 验证字符类型
	if ok := isValidUsername(value); !ok {
		return fmt.Errorf("用户名字符只能包括大小写英文、数字、下划线")
	}
	return nil
}

// 验证全名
func ValidateFullName(value string, min int, max int) error {
	// 验证长度
	err := ValidateString(value, min, max)
	if err != nil {
		return err
	}

	// 验证字符类型
	if ok := isValidFullname(value); !ok {
		return fmt.Errorf("用户全称字符只能包括大小写英文、空格")
	}
	return nil
}

// 验证密码
func ValidatePassword(value string) error {
	// 验证长度
	err := ValidateString(value, 6, 20)
	if err != nil {
		return err
	}
	return nil
}

// 验证邮箱
func ValidateEmail(value string) error {
	_, err := mail.ParseAddress(value)
	return err
}
