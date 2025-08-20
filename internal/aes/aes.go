package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

func getRandomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

type AES struct {
	key []byte
}

func (c AES) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, _ := getRandomBytes(aesgcm.NonceSize())

	ciphertext := aesgcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func (c AES) EncryptBase64(data []byte) (string, error) {
	ciphertext, err := c.Encrypt(data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c AES) Decrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (c AES) DecryptBase64(data string) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	return c.Decrypt(encrypted)

}

func New(key string) *AES {
	var bytesKey []byte

	if len(key) == 0 {
		bytesKey, _ = getRandomBytes(32)
	} else {
		var err error
		bytesKey, err = base64.URLEncoding.DecodeString(key)
		if err != nil {
			panic(err)
		}
	}
	if len(bytesKey) != 32 {
		panic("")
	}

	return &AES{
		key: bytesKey,
	}
}
