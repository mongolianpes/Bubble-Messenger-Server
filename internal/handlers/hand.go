package handlers

import (
	"bubble/internal/files"
	"bubble/internal/messenger"
	"bubble/internal/users"
)

type Handler struct {
	UsersService     users.UsersService
	MessengerService messenger.MessengerService
	FilesService     files.FilesService
}

func NewHand() (*Handler, error) {
	usersService, err := users.NewClient()
	if err != nil {
		return nil, err
	}

	messengerService, err := messenger.NewClient()
	if err != nil {
		return nil, err
	}

	filesService, err := files.NewClient()
	if err != nil {
		return nil, err
	}

	return &Handler{
		UsersService:     usersService,
		MessengerService: messengerService,
		FilesService:     filesService,
	}, nil
}
