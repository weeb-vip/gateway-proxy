package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type customClaims struct {
	jwt.RegisteredClaims
	Purpose *string `json:"purpose"`
}

func main() {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	// Create custom claims
	purpose := "access"
	claims := customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user123",
			Audience:  []string{"gateway-proxy"},
			Issuer:    "test-issuer",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        "test-jwt-id",
		},
		Purpose: &purpose,
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Add kid header
	token.Header["kid"] = "test-key-id"

	// Sign token
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// Output the token
	fmt.Println("Generated JWT Token:")
	fmt.Println(tokenString)

	// Output the public key for verification
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	fmt.Println("\nPublic Key (for key management service):")
	fmt.Println(string(publicKeyPEM))
}