// utils/apple.go

package utils

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/config"
)

func VerifyAppleToken(identityToken string, clientID string) (jwt.MapClaims, error) {
	appleJWKSURL := "https://appleid.apple.com/auth/keys"
	k, err := keyfunc.NewDefault([]string{appleJWKSURL})
	if err != nil {
		return nil, apperr.InternalServerError("failed to create keyfunc", err)
	}

	token, err := jwt.Parse(identityToken, k.Keyfunc)
	if err != nil {
		return nil, apperr.InternalServerError("invalid token", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["iss"] != "https://appleid.apple.com" {
			return nil, apperr.Unauthorized("invalid issuer", nil)
		}

		if claims["aud"] != clientID {
			return nil, apperr.Unauthorized("invalid audience", nil)
		}

		return claims, nil
	}

	return nil, apperr.Unauthorized("invalid token claims", nil)
}

func GenerateAppleClientSecret() (string, error) {
	block, _ := pem.Decode([]byte(config.AppConfig.AppleP8Key))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block containing apple private key")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": config.AppConfig.AppleTeamID,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour * 24 * 30 * 6).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": config.AppConfig.AppleBundleID,
	})

	token.Header["kid"] = config.AppConfig.AppleKeyID

	return token.SignedString(privKey)
}

func GetAppleRefreshToken(code string) (string, error) {
	clientSecret, err := GenerateAppleClientSecret()
	if err != nil {
		return "", err
	}

	data := url.Values{}
	data.Set("client_id", config.AppConfig.AppleBundleID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://appleid.apple.com/auth/token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		RefreshToken string `json:"refresh_token"`
		Error        string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("apple token error: %s", result.Error)
	}

	return result.RefreshToken, nil
}
