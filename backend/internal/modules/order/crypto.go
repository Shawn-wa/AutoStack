package order

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var cryptoKey []byte

// InitCrypto 初始化加密模块
func InitCrypto(secretKey string) error {
	if len(secretKey) != 32 {
		return errors.New("加密密钥必须为32字节")
	}
	cryptoKey = []byte(secretKey)
	return nil
}

// Encrypt 加密字符串
func Encrypt(plaintext string) (string, error) {
	if cryptoKey == nil {
		return "", errors.New("加密模块未初始化")
	}

	block, err := aes.NewCipher(cryptoKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密字符串
func Decrypt(ciphertext string) (string, error) {
	if cryptoKey == nil {
		return "", errors.New("加密模块未初始化")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(cryptoKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文长度不足")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
