package util

import (
	"math/rand"
	"strings"
	"time"
)

var letters = "abcdefghijklmnopqrstuvwxyz"

func init() {
	// 刷新种子值
	rand.Seed(time.Now().UnixNano())
}

func RandomInt64(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomString(n int) string {
	var sb strings.Builder
	l := len(letters)
	for i := 0; i < n; i++ {
		sb.WriteByte(letters[rand.Intn(l)])
	}
	return sb.String()
}

// RandomOwner 随机生成一个6个字符的accounts.owner属性
func RandomOwner() string {
	return RandomString(6)
}

// RandomBalance 随机生成一个值为0~1000的accounts.balance属性
func RandomBalance() int64 {
	return RandomInt64(0, 1000)
}

// RandomCurrency 随机生成一个account.currency属性
func RandomCurrency() string {
	currncies := []string{"EUR", "USD", "CAD"}
	return currncies[rand.Intn(len(currncies))]
}

// RandomAmount 随机生成一个在-remain~1000的数值用于Entry的生成
func RandomAmount(remain int64) int64 {
	return RandomInt64(-remain, 1000)
}

// RandomAccountId 随机生成一个合法的AccountID
func RandomAccountId(n int64) int64 {
	return RandomInt64(1, n)
}
