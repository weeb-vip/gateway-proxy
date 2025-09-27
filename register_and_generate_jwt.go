package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type customClaims struct {
	jwt.RegisteredClaims
	Purpose *string `json:"purpose"`
}

type GraphQLResponse struct {
	Data struct {
		RegisterPublicKey struct {
			ID   string `json:"id"`
			Body string `json:"body"`
		} `json:"registerPublicKey"`
	} `json:"data"`
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

func main() {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	// Convert public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Register the public key with the key management service
	mutation := `mutation RegisterPublicKey($publicKey: String!) {
		registerPublicKey(publicKey: $publicKey) {
			id
			body
		}
	}`

	request := GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"publicKey": string(publicKeyPEM),
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post("http://localhost:5001/graphql", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var graphqlResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&graphqlResp); err != nil {
		log.Fatal(err)
	}

	keyID := graphqlResp.Data.RegisterPublicKey.ID
	fmt.Printf("Registered key with ID: %s\n", keyID)

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

	// Add kid header with the registered key ID
	token.Header["kid"] = keyID

	// Sign token
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// Output the token
	fmt.Println("\nGenerated JWT Token (registered with key service):")
	fmt.Println(tokenString)
}