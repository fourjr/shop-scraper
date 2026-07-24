package luckin

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
)

func RefreshLoginToken(email, password string) (string, error) {
	log.Printf("[luckin] warn: refreshing login token for email %s", email)
	encPassword, err := encryptPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt password: %v", err)
	}
	var body = struct {
		Email       string `json:"email"`
		EncPassword string `json:"encPassword"`
	}{
		Email:       email,
		EncPassword: encPassword,
	}

	var response struct {
		PasswordVerified bool `json:"passwordVerified"`
	}

	resp, err := request("/api/capi/resource/isalestradecapi/user/mailLogin", body, "", &response)
	if err != nil {
		return "", fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	if !response.PasswordVerified {
		return "", fmt.Errorf("login failed: password verification failed")
	}
	var token string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "LK_PROD_ILUCKYINWAP_SID" {
			token = cookie.Value
			break
		}
	}
	if token == "" {
		return "", fmt.Errorf("login failed: no auth token received")
	}

	return token, nil
}

func getPublicKey() (*rsa.PublicKey, error) {
	config := struct {
		Keys []string `json:"keys"`
	}{
		Keys: []string{"isalesmarketcapi.app.login.key"},
	}

	var response struct {
		LoginKey string `json:"isalesmarketcapi.app.login.key"`
	}

	_, err := request("/api/capi/resource/isalesmarketingcapi/configValue/configValue", config, "", &response)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}

	var key struct {
		PublicKey string `json:"publicKey"`
	}
	if err := json.Unmarshal([]byte(response.LoginKey), &key); err != nil {
		return nil, fmt.Errorf("failed to decode login key: %v", err)
	}

	der, err := base64.StdEncoding.DecodeString(key.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid Base64: %v", err)
	}

	parsedKey, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %v", err)
	}

	publicKey, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not an RSA public key")
	}

	return publicKey, nil
}

func encryptPassword(password string) (string, error) {
	publicKey, err := getPublicKey()
	if err != nil {
		return "", fmt.Errorf("failed to get public key: %v", err)
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(password))
	if err != nil {
		return "", fmt.Errorf("Encryption failed: %v", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
