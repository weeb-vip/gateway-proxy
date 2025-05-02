package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type key struct {
	PublicKey  string
	PrivateKey string
}

const KeySize = 2048

func generateKeyPair() (*key, error) {
	// There's no need to handle the error since rsa.GenerateKey with the given parameter can never produce error.
	keyPair, _ := rsa.GenerateKey(rand.Reader, KeySize)

	privateKey, err := getEncodedPrivateKey(keyPair)
	if err != nil {
		return nil, err
	}

	publicKey, err := getEncodedPublicKey(keyPair)
	if err != nil {
		return nil, err
	}

	return &key{PublicKey: publicKey, PrivateKey: privateKey}, nil
}

func generateValidJWT(signingKey string, ttl time.Duration, sub string, purpose string, kid *string, nbf *time.Duration) (string, error) {
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(signingKey))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, buildClaims(ttl, sub, purpose, nbf))
	if kid != nil {
		token.Header["kid"] = *kid
	}
	signedToken, err := token.SignedString(signKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func getEncodedPrivateKey(keyPair *rsa.PrivateKey) (string, error) {
	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(keyPair)
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: pkcs8Key})), nil
}

func getEncodedPublicKey(keyPair *rsa.PrivateKey) (string, error) {
	publicKey := keyPair.Public()

	marshalledPublicKey, err := x509.MarshalPKIXPublicKey(publicKey.(*rsa.PublicKey))
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: marshalledPublicKey})), nil
}

func buildClaims(ttl time.Duration, sub string, purpose string, nbf *time.Duration) jwt.MapClaims {
	getNbf := func(nbf *time.Duration) time.Duration {
		if nbf == nil {
			return time.Second * 0
		}

		return *nbf
	}
	return jwt.MapClaims{
		"nbf":     time.Now().Add(getNbf(nbf)).Unix(),
		"iss":     "smokey",
		"aud":     "getweed",
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(ttl).Unix(),
		"sub":     sub,
		"purpose": purpose,
	}
}
