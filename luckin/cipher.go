package luckin

import (
	"crypto/aes"
	"encoding/base64"
	"fmt"
	"strings"
)

func initKey(value string) []byte {
	key := make([]byte, 16)
	copy(key, []byte(value))
	return key
}

func decodeCiphertext(value string) ([]byte, error) {
	base64Value := strings.NewReplacer(
		"-", "+",
		"_", "/",
	).Replace(value)

	base64Value += strings.Repeat("=", (4-len(base64Value)%4)%4)

	data, err := base64.StdEncoding.DecodeString(base64Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}

	return data, nil
}

func encodeCiphertext(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

func padPKCS7(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 || blockSize > 255 {
		return nil, fmt.Errorf("PKCS#7 block size must be between 1 and 255, got %d", blockSize)
	}

	padding := blockSize - len(data)%blockSize
	padded := make([]byte, len(data)+padding)
	copy(padded, data)

	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}

	return padded, nil
}

func unpadPKCS7(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("decrypted data is empty")
	}

	padding := int(data[len(data)-1])

	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, fmt.Errorf("invalid PKCS#7 padding length: %d", padding)
	}

	for _, value := range data[len(data)-padding:] {
		if int(value) != padding {
			return nil, fmt.Errorf("invalid PKCS#7 padding")
		}
	}

	return data[:len(data)-padding], nil
}

func aesDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	if len(ciphertext) == 0 || len(ciphertext)%block.BlockSize() != 0 {
		return nil, fmt.Errorf(
			"ciphertext length must be a non-zero multiple of %d, got %d",
			block.BlockSize(),
			len(ciphertext),
		)
	}

	plaintext := make([]byte, len(ciphertext))

	for start := 0; start < len(ciphertext); start += block.BlockSize() {
		end := start + block.BlockSize()
		block.Decrypt(plaintext[start:end], ciphertext[start:end])
	}

	return unpadPKCS7(plaintext, block.BlockSize())
}

func aesEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	padded, err := padPKCS7(plaintext, block.BlockSize())
	if err != nil {
		return nil, fmt.Errorf("failed to pad plaintext: %w", err)
	}

	ciphertext := make([]byte, len(padded))
	for start := 0; start < len(padded); start += block.BlockSize() {
		end := start + block.BlockSize()
		block.Encrypt(ciphertext[start:end], padded[start:end])
	}

	return ciphertext, nil
}

func Encrypt(key string, value string) (string, error) {
	keyInit := initKey(key)

	ciphertext, err := aesEncrypt([]byte(value), keyInit)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %w", err)
	}

	return encodeCiphertext(ciphertext), nil
}

func Decrypt(key string, encryptedString string) (string, error) {
	ciphertext, err := decodeCiphertext(encryptedString)
	if err != nil {
		return "", err
	}

	keyInit := initKey(key)

	plaintext, err := aesDecrypt(ciphertext, keyInit)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	return string(plaintext), nil
}
