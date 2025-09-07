package handler

import "github.com/zerodayz7/http-server/internal/service"

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService, sessionService *service.SessionService) *UserHandler {
	return &UserHandler{userService: userService}
}
