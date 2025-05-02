package jwt_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/weeb-vip/gateway-proxy/internal/jwt"
	"github.com/weeb-vip/gateway-proxy/internal/keys"
	"testing"
	"time"
)

var keyPair, _ = generateKeyPair()

type signingKeys struct {
	k   string
	kid string
	err error
}

func (s signingKeys) Fetch() error { return nil }

func (s signingKeys) SetupBackgroundPolling(pollingDuration time.Duration) {}

func (s signingKeys) FindKeyByID(id string) (*keys.Key, error) {
	if s.err != nil {
		return nil, s.err
	}

	return &keys.Key{
		ID:   s.kid,
		Body: s.k,
	}, nil
}

func TestNewParser(t *testing.T) {
	t.Run("if it sees wrong algorithm, it returns error", func(t *testing.T) {
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

		claims, err := jwt.NewParser(nil).Parse(token)

		assert.EqualError(t, err, "expected signing method: RS256, got: HS256")
		assert.Nil(t, claims)
	})
	t.Run("if token is expired, it returns error", func(t *testing.T) {
		token, _ := generateTestJWTCredentials(time.Hour*-1, "sub", "purpose", nil, nil)

		claims, err := jwt.NewParser(nil).Parse(token)

		assert.ErrorContains(t, err, "token is expired")
		assert.Nil(t, claims)
	})
	t.Run("if JWT doesn't contain kid, it returns an error", func(t *testing.T) {
		token, _ := generateTestJWTCredentials(time.Second*10, "sub", "purpose", nil, nil)

		claims, err := jwt.NewParser(nil).Parse(token)

		assert.EqualError(t, err, "no kid in jwt")
		assert.Nil(t, claims)
	})
	t.Run("if signing key couldn't be found, it returns an error", func(t *testing.T) {
		keyID := "non-existing-key"
		token, _ := generateTestJWTCredentials(time.Second*10, "sub", "purpose", &keyID, nil)
		keyManager := signingKeys{k: keyPair.PublicKey, kid: "valid-key-id", err: errors.New("my-error")}

		claims, err := jwt.NewParser(keyManager).Parse(token)

		assert.EqualError(t, err, "couldn't retrieve the signing key")
		assert.Nil(t, claims)
	})

	t.Run("if somehow rsa public key is invalid, it returns a parsing error", func(t *testing.T) {
		keyID := "my-kid"
		token, _ := generateTestJWTCredentials(time.Second*10, "sub", "purpose", &keyID, nil)
		keyManager := signingKeys{k: "some-invalid-key", kid: "my-kid", err: nil}

		claims, err := jwt.NewParser(keyManager).Parse(token)

		assert.EqualError(t, err, "public key malformed")
		assert.Nil(t, claims)
	})
	t.Run("given a valid key, we get a valid parsed JWT back", func(t *testing.T) {
		keyID := "my-kid"
		token, _ := generateTestJWTCredentials(time.Second*10, "sub", "purpose", &keyID, nil)
		keyManager := signingKeys{k: keyPair.PublicKey, kid: "my-kid", err: nil}

		claims, err := jwt.NewParser(keyManager).Parse(token)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "getweed", *claims.Audience)
		assert.Equal(t, "sub", *claims.Subject)
		assert.Equal(t, "purpose", *claims.Purpose)
	})
	t.Run("if a key is not valid yet (nbf), it returns an error", func(t *testing.T) {
		keyID := "my-kid"
		nbf := time.Second * 5
		token, _ := generateTestJWTCredentials(time.Second*10, "sub", "purpose", &keyID, &nbf)
		keyManager := signingKeys{k: keyPair.PublicKey, kid: "my-kid", err: nil}

		claims, err := jwt.NewParser(keyManager).Parse(token)

		assert.EqualError(t, err, "token is not valid yet")
		assert.Nil(t, claims)
	})
}

func generateTestJWTCredentials(ttl time.Duration, subject string, purpose string, kid *string, nbf *time.Duration) (string, error) {
	return generateValidJWT(keyPair.PrivateKey, ttl, subject, purpose, kid, nbf)
}
