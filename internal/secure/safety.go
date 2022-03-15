package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/foximilUno/metrics/internal/types"
	"strconv"
	"strings"
)

func getAesGSMWithNonce(keyString string) (cipher.AEAD, []byte, error) {
	key32 := sha256.Sum256([]byte(keyString))
	key := key32[:16]

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	aesGcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, nil, err
	}
	nonce := key[:aesGcm.NonceSize()]
	return aesGcm, nonce, nil
}

func encryptMetric(stringToEncrypt string, keyString string) (string, error) {
	aesGcm, nonce, err := getAesGSMWithNonce(keyString)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(aesGcm.Seal(nil, nonce, []byte(stringToEncrypt), nil)), nil
}

func EncryptGaugeMetric(m *types.Metrics, key string) (string, error) {
	newVal := *m.Value
	return encryptMetric(fmt.Sprintf("%s:gauge:%f", m.ID, newVal), key)
}

func EncryptCounterMetric(m *types.Metrics, key string) (string, error) {
	newVal := *m.Delta
	return encryptMetric(fmt.Sprintf("%s:counter:%d", m.ID, newVal), key)
}

func DecryptMetric(stringToDecrypt string, keyString string) (*types.Metrics, error) {
	aesGcm, nonce, err := getAesGSMWithNonce(keyString)
	if err != nil {
		return nil, err
	}
	encrypted, err := hex.DecodeString(stringToDecrypt)
	if err != nil {
		return nil, err
	}

	decrypted, err := aesGcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}
	decryptedString := string(decrypted)
	values := strings.Split(decryptedString, ":")
	newMetric := &types.Metrics{
		ID:    values[0],
		MType: values[1],
	}
	switch newMetric.MType {
	case "gauge":
		newVar, err := strconv.ParseFloat(values[2], 64)
		if err != nil {
			return nil, err
		}
		newMetric.Value = &newVar
	case "counter":
		newVar, err := strconv.ParseInt(values[2], 10, 64)
		if err != nil {
			return nil, err
		}
		newMetric.Delta = &newVar
	default:
		return nil, fmt.Errorf("need detail new type=%s", newMetric.MType)
	}

	return newMetric, nil
}
