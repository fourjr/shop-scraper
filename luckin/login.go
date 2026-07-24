package luckin

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"shops/http"
)

func RefreshLoginToken(email, password string) (string, error) {
	tokenBytes := make([]byte, 128)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

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

	reader, err := json.Marshal(body)
	resp, err := http.DoPost(baseUrl+"/api/capi/resource/isalestradecapi/user/mailLogin", reader, map[string]string{
		"Content-Type":    "application/json",
		"Accept-Language": "en-US",
		"X-LK-Tenant":     "LKSG",
	})

	if err != nil {
		return "", fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		BusiCode string `json:"busiCode"`
		Code     int    `json:"code"`
		Content  struct {
			PasswordVerified bool `json:"passwordVerified"`
		} `json:"content"`
	}
	rawContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	if err := json.Unmarshal(rawContent, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}
	if response.Code != 1 {
		return "", fmt.Errorf("api request failed with error code %d: %s - %s", response.Code, response.BusiCode, string(rawContent))
	}
	if response.BusiCode != "200" {
		return "", fmt.Errorf("api request failed with busi code %s - %s", response.BusiCode, string(rawContent))
	}
	if !response.Content.PasswordVerified {
		return "", fmt.Errorf("login failed: password verification failed - %s", string(rawContent))
	}
	var token string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "LK_PROD_ILUCKYINWAP_SID" {
			token = cookie.Value
			log.Printf("authenticated SID received: %s...", cookie.Value)
		}
	}
	if token == "" {
		return "", fmt.Errorf("login failed: no SID cookie received - %s", string(rawContent))
	}
	log.Println(string(rawContent))
	if err := getUserInfo(token); err != nil {
		return "", fmt.Errorf("failed to fetch user info: %v", err)
	}
	return token, nil
}

func getUserInfo(token string) error {
	// Implementation for fetching user info
	reader, err := json.Marshal(struct {
		Version string `json:"version"`
	}{
		Version: "1.4.35",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}
	resp, err := http.DoPost(baseUrl+"/api/capi/resource/isalestradecapi/user/queryUserInfo", reader, map[string]string{
		"Content-Type":    "application/json",
		"Accept-Language": "en-US",
		"Cookie":          fmt.Sprintf("LK_PROD_ILUCKYINWAP_SID=%s; ", token),
		"X-LK-Tenant":     "LKSG",
	})

	if err != nil {
		return fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		BusiCode string `json:"busiCode"`
		Code     int    `json:"code"`
		Content  struct {
		} `json:"content"`
	}
	rawContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	if err := json.Unmarshal(rawContent, &response); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	if response.Code != 1 {
		return fmt.Errorf("api request failed with error code %d: %s - %s", response.Code, response.BusiCode, string(rawContent))
	}
	if response.BusiCode != "200" {
		return fmt.Errorf("api request failed with busi code %s - %s", response.BusiCode, string(rawContent))
	}
	return nil
}

func getPublicKey() (*rsa.PublicKey, error) {
	config := struct {
		AppVersion string   `json:"appVersion"`
		Keys       []string `json:"keys"`
	}{
		AppVersion: "1.4.35",
		Keys:       []string{"isalesmarketcapi.app.login.key"},
	}
	reader, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}
	resp, err := http.DoPost(baseUrl+"/api/capi/resource/isalesmarketingcapi/configValue/configValue", reader, map[string]string{
		"Content-Type":    "application/json",
		"Accept-Language": "en-US",
		"X-LK-Tenant":     "LKSG",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()
	var response struct {
		BusiCode string `json:"busiCode"`
		Code     int    `json:"code"`
		Content  struct {
			LoginKey string `json:"isalesmarketcapi.app.login.key"`
		} `json:"content"`
	}
	rawContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	if err := json.Unmarshal(rawContent, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	if response.Code != 1 {
		return nil, fmt.Errorf("api request failed with error code %d: %s - %s", response.Code, response.BusiCode, string(rawContent))
	}
	if response.BusiCode != "200" {
		return nil, fmt.Errorf("api request failed with busi code %s - %s", response.BusiCode, string(rawContent))
	}

	var key struct {
		PublicKey string `json:"publicKey"`
	}
	if err := json.Unmarshal([]byte(response.Content.LoginKey), &key); err != nil {
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
