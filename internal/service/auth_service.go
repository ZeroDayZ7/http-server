package service

import (
	"fmt"

	"github.com/zerodayz7/http-server/internal/errors"
	"github.com/zerodayz7/http-server/internal/model"
	"github.com/zerodayz7/http-server/internal/repository"
	"github.com/zerodayz7/http-server/internal/service/security"
	"github.com/zerodayz7/http-server/internal/shared/logger"

	"go.uber.org/zap"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) IsEmailExists(email string) (bool, error) {
	log := logger.GetLogger()
	log.Debug("IsEmailExists", zap.String("email", email))

	exists, err := s.repo.EmailExists(email)
	if err != nil {
		log.Error("repo.EmailExists failed", zap.Error(err))
		return false, err
	}
	return exists, nil
}

func (s *UserService) IsUsernameExists(username string) (bool, error) {
	log := logger.GetLogger()
	log.Debug("IsUsernameExists", zap.String("username", username))

	exists, err := s.repo.UsernameExists(username)
	if err != nil {
		log.Error("repo.UsernameExists failed", zap.Error(err))
		return false, err
	}
	return exists, nil
}

func (s *UserService) IsEmailOrUsernameExists(email, username string) (bool, bool, error) {
	return s.repo.EmailOrUsernameExists(email, username)
}

func (s *UserService) CheckPassword(email, password string) (bool, bool, error) {
	log := logger.GetLogger()
	log.Debug("CheckPassword", zap.String("email", email))

	u, err := s.repo.GetByEmail(email)
	if err != nil {
		log.Error("GetByEmail failed", zap.Error(err), zap.String("email", email))
		return false, false, err
	}

	valid, err := security.VerifyPassword(password, u.Password)
	if err != nil {
		log.Error("VerifyPassword failed", zap.Error(err))
		return false, false, err
	}
	return valid, u.TwoFactorEnabled, nil
}

func (s *UserService) Verify2FACode(email, code string) (bool, error) {
	log := logger.GetLogger()
	log.Debug("Verify2FACode", zap.String("email", email))

	u, err := s.repo.GetByEmail(email)
	if err != nil {
		log.Error("GetByEmail failed", zap.Error(err))
		return false, err
	}
	if !u.TwoFactorEnabled {
		log.Warn("2FA not enabled", zap.String("email", email))
		return false, nil
	}

	return code == u.TwoFactorSecret, nil
}

func (s *UserService) Register(username, email, rawPassword string) (*model.User, error) {
	log := logger.GetLogger()
	log.Debug("Register", zap.String("email", email), zap.String("username", username))

	emailExists, usernameExists, err := s.repo.EmailOrUsernameExists(email, username)
	if err != nil {
		return nil, fmt.Errorf("checking email/username existence: %w", err)
	}

	if emailExists {
		return nil, errors.ErrEmailExists
	}
	if usernameExists {
		return nil, errors.ErrUsernameExists
	}

	hash, err := security.HashPassword(rawPassword)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	u := &model.User{
		Username: username,
		Email:    email,
		Password: hash,
	}

	if err := s.repo.CreateUser(u); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return u, nil
}
