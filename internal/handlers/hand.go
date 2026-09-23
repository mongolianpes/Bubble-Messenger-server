package handlers

import (
	"bubble/internal/users"
)

type Handler struct {
	UsersService users.UsersService
}

func NewHand() (*Handler, error) {
	usersService, err := users.NewClient()
	if err != nil {
		return nil, err
	}

	return &Handler{
		UsersService: usersService,
	}, nil
}
