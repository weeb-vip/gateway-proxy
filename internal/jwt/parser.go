package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/weeb-vip/gateway-proxy/internal/poller"
)

func (p parser) Parse(token string) (*ParsedJWT, error) {
	claims := &customClaims{}
	t, err := jwt.ParseWithClaims(token, claims, p.keyFunc)
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, errors.New("invalid token")
	}

	return &ParsedJWT{
		Subject:  &claims.Subject,
		Audience: &claims.Audience[0],
		Purpose:  claims.Purpose,
	}, nil
}

func (p parser) keyFunc(token *jwt.Token) (any, error) {
	if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
		return nil, fmt.Errorf("expected signing method: %s, got: %s", jwt.SigningMethodRS256.Alg(), token.Method.Alg())
	}
	err := token.Claims.Valid()
	if err != nil {
		return nil, err
	}
	keyID := token.Header["kid"]
	if keyID == nil || keyID.(string) == "" {
		return nil, errors.New("no kid in jwt")
	}
	key, err := p.keyManager.FindKeyByID(keyID.(string))
	if err != nil {
		return nil, errors.New("couldn't retrieve the signing key")
	}

	rsaKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(key.Body))
	if err != nil {
		return nil, errors.New("public key malformed")
	}

	return rsaKey, nil
}

func NewParser(keyManager poller.KeyManager) Parser {
	return parser{keyManager: keyManager}
}
