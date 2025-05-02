package jwt

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/weeb-vip/gateway-proxy/internal/poller"
)

type customClaims struct {
	jwt.RegisteredClaims
	Purpose *string `json:"purpose"`
}

type ParsedJWT struct {
	Subject  *string
	Audience *string
	Purpose  *string
}

type Parser interface {
	Parse(token string) (*ParsedJWT, error)
}

type parser struct {
	keyManager poller.KeyManager
}
