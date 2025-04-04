package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// EncryptFileInfo mã hóa bot token và file ID thành một chuỗi an toàn
func EncryptFileInfo(botToken, fileID string) (string, error) {
	secretKey := []byte(getEncryptionKey())

	// Kết hợp thông tin với ký tự phân cách
	plaintext := []byte(fmt.Sprintf("%s|%s", botToken, fileID))

	// Tạo block cipher
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	// Tạo GCM (Galois/Counter Mode)
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Tạo nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Mã hóa
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	// Mã hóa base64 để sử dụng an toàn trong URL
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptFileInfo giải mã chuỗi đã mã hóa để lấy bot token và file ID
func DecryptFileInfo(encryptedData string) (botToken, fileID string, err error) {
	secretKey := []byte(getEncryptionKey())

	// Giải mã base64
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", "", err
	}

	// Tạo block cipher
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", "", err
	}

	// Tạo GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	// Xác định nonce size
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", "", errors.New("ciphertext too short")
	}

	// Trích xuất nonce và ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Giải mã
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", "", err
	}

	// Tách bot token và file ID
	parts := strings.Split(string(plaintext), "|")
	if len(parts) != 2 {
		return "", "", errors.New("invalid data format after decryption")
	}

	return parts[0], parts[1], nil
}

// getEncryptionKey lấy key từ biến môi trường hoặc sử dụng key mặc định
func getEncryptionKey() string {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		// Tạo một key mặc định với độ dài 32 bytes (256 bits) cho AES-256
		// Trong môi trường production, đảm bảo sử dụng biến môi trường
		key = "telerealm-default-encryption-key-32b"
	}

	// Đảm bảo key có đúng độ dài cho AES-256
	if len(key) < 32 {
		key = key + strings.Repeat("0", 32-len(key))
	}

	return key[:32]
}
