package joinsecret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

var (
	mu         sync.RWMutex
	keys       map[int][]byte
	keyVersion int
)

func Configure(encodedKeys string, version int) error {
	if version <= 0 {
		return errors.New("JOIN_CODE_KEY_VERSION debe ser positivo")
	}
	parsed := map[int][]byte{}
	for _, entry := range strings.Split(encodedKeys, ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), ":", 2)
		if len(parts) != 2 {
			return errors.New("JOIN_CODE_ENCRYPTION_KEYS debe usar version:base64")
		}
		v, err := strconv.Atoi(parts[0])
		key, decodeErr := base64.StdEncoding.DecodeString(parts[1])
		if err != nil || v <= 0 || decodeErr != nil || len(key) != 32 {
			return errors.New("cada clave debe tener version positiva y exactamente 32 bytes en base64")
		}
		if _, exists := parsed[v]; exists {
			return errors.New("version de clave duplicada")
		}
		parsed[v] = key
	}
	if _, ok := parsed[version]; !ok {
		return errors.New("JOIN_CODE_KEY_VERSION no existe en el llavero")
	}
	mu.Lock()
	keys = parsed
	keyVersion = version
	mu.Unlock()
	return nil
}

func ActiveVersion() int {
	mu.RLock()
	defer mu.RUnlock()
	return keyVersion
}

func Encrypt(code string, reservationID int) (nonce, ciphertext []byte, version int, err error) {
	mu.RLock()
	key := append([]byte(nil), keys[keyVersion]...)
	version = keyVersion
	mu.RUnlock()
	if len(key) != 32 {
		return nil, nil, 0, errors.New("cifrado de codigos no configurado")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, 0, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, 0, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, 0, err
	}
	ciphertext = gcm.Seal(nil, nonce, []byte(code), aad(reservationID))
	return nonce, ciphertext, version, nil
}

func Decrypt(nonce, ciphertext []byte, version, reservationID int) (string, error) {
	mu.RLock()
	key := append([]byte(nil), keys[version]...)
	mu.RUnlock()
	if len(key) != 32 {
		return "", fmt.Errorf("version de clave %d no disponible", version)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plain, err := gcm.Open(nil, nonce, ciphertext, aad(reservationID))
	if err != nil {
		return "", errors.New("secreto de codigo invalido")
	}
	return string(plain), nil
}

func aad(reservationID int) []byte {
	return []byte("poli-redi:reservation:" + strconv.Itoa(reservationID))
}
