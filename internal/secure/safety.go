package secure

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/foximilUno/metrics/internal/types"
)

func encryptString(stringToEncrypt string, keyString string) (string, error) {
	h := hmac.New(sha256.New, []byte(keyString))
	h.Write([]byte(stringToEncrypt))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash), nil
}

func EncryptMetric(m *types.Metrics, key string) (string, error) {
	return encryptString(m.Format(), key)
}

func IsValidHash(msg string, hashString string, keyString string) (bool, error) {
	hash, err := hex.DecodeString(hashString)
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, []byte(keyString))

	_, err = mac.Write([]byte(msg))
	if err != nil {
		return false, err
	}

	expectedHash := mac.Sum(nil)
	return hmac.Equal(hash, expectedHash), nil
}
