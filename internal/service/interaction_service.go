package service

import (
	"time"

	"github.com/zerodayz7/http-server/internal/repository/mysql"
)

type InteractionService struct {
	repo *mysql.MySQLInteractionRepo
}

func NewInteractionService(repo *mysql.MySQLInteractionRepo) *InteractionService {
	return &InteractionService{repo: repo}
}

// Record zapisuje interakcję (visit, like, comment)
func (s *InteractionService) Record(ip string, userID *uint, typ string, value int, content *string) error {
	return s.repo.Add(ip, userID, typ, value, content, time.Now())
}

// CountByType zwraca liczbę interakcji danego typu
func (s *InteractionService) CountByType(typ string) (int, error) {
	// używamy istniejącej metody Count w repo
	return s.repo.Count(typ)
}

func (s *InteractionService) GetLastVisitByIP(ip string) (time.Time, error) {
	return s.repo.GetLastVisit(ip)
}

func (s *InteractionService) GetLastInteractionByIP(ip, typ string) (time.Time, error) {
	return s.repo.GetLastInteraction(ip, typ)
}
