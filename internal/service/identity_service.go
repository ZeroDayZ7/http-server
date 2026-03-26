package service

import (
	"crypto/sha256"
	"encoding/hex"
)

type IdentityService interface {
	GenerateFingerprint(ip, ua, lang string) string
}

type identityService struct {
	salt string
}

func NewIdentityService(salt string) IdentityService {
	return &identityService{salt: salt}
}

func (s *identityService) GenerateFingerprint(ip, ua, lang string) string {
	raw := ip + "|" + ua + "|" + lang + "|" + s.salt
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
