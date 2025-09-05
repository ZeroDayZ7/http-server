package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/zerodayz7/http-server/internal/shared/logger"

	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
)

const (
	memory      = 64 * 1024
	iterations  = 3
	parallelism = 2
	saltLength  = 16
	keyLength   = 32
)

func HashPassword(password string) (string, error) {
	log := logger.GetLogger()
	log.Debug("Hashowanie hasła")

	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		log.Error("Błąd generowania soli", zap.Error(err))
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	log.Debug("Hasło zahashowane", zap.String("encoded_salt", encodedSalt))
	return encodedSalt + "$" + encodedHash, nil
}

func VerifyPassword(password, encoded string) (bool, error) {
	log := logger.GetLogger()
	log.Debug("Sprawdzanie hasła")

	parts := strings.Split(encoded, "$")
	if len(parts) != 2 {
		log.Error("Nieprawidłowy format hasła w bazie", zap.String("encoded", encoded))
		return false, errors.New("nieprawidłowy format hasła w bazie")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		log.Error("Błąd dekodowania soli", zap.Error(err))
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		log.Error("Błąd dekodowania hasha", zap.Error(err))
		return false, err
	}

	computedHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	isValid := subtle.ConstantTimeCompare(hash, computedHash) == 1
	if isValid {
		log.Info("Hasło poprawne")
	} else {
		log.Info("Nieprawidłowe hasło")
	}
	return isValid, nil
}
