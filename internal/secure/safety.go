package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/foximilUno/metrics/internal/types"
)

func getAesGSMWithNonce(keyString string) (cipher.AEAD, []byte, error) {
	key32 := sha256.Sum256([]byte(keyString))
	key := key32[:]

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	aesGcm, err := cipher.NewGCMWithTagSize(aesBlock, 16)
	if err != nil {
		return nil, nil, err
	}
	nonce := key[:aesGcm.NonceSize()]
	return aesGcm, nonce, nil
}

func encryptMetric(stringToEncrypt string, keyString string) (string, error) {
	h := hmac.New(sha256.New, []byte(keyString))
	h.Write([]byte(stringToEncrypt))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash), nil
}

func EncryptGaugeMetric(m *types.Metrics, key string) (string, error) {
	newVal := *m.Value
	return encryptMetric(fmt.Sprintf("%s:gauge:%f", m.ID, newVal), key)
}

func EncryptCounterMetric(m *types.Metrics, key string) (string, error) {
	newVal := *m.Delta
	return encryptMetric(fmt.Sprintf("%s:counter:%d", m.ID, newVal), key)
}

func IsValidHash(hashString string, keyString string) (bool, error) {
	sig, err := hex.DecodeString(hashString)
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, []byte(keyString))
	mac.Write([]byte(hashString))

	return hmac.Equal(sig, mac.Sum(nil)), nil
}
