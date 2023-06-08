package token

import "time"

type Maker interface {
	// CreateToken 根据Username与Duration创建Token
	CreateToken(username string, duration time.Duration) (string, *Payload, error)

	// VerifyToken 验证token是否有效
	VerifyToken(token string) (*Payload, error)
}
