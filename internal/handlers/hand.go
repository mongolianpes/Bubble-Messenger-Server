package handlers

import (
	"bubble/internal/audiodialog"
	"bubble/internal/files"
	"bubble/internal/ipblocker"
	"bubble/internal/messenger"
	"bubble/internal/users"
	"sync"
	"time"
)

type Handler struct {
	UsersService       users.UsersService
	MessengerService   messenger.MessengerService
	FilesService       files.FilesService
	AudioDialogService audiodialog.AudioDialogService
	Blocker            ipblocker.Blocker
}

type PagesHandler struct {
	homeHTMLPage  string
	authHTMLPage  string
	chatsHTMLPage string
	sync.Mutex
}

type TLSHandler struct {
	Handlers                *Handler
	PagesHandlers           *PagesHandler
	Auth                    users.Auth
	Blocker                 ipblocker.Blocker
	UsersUsedRequestsNonces map[string]string
}

func NewHand() (*TLSHandler, error) {
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

	audioDialogService, err := audiodialog.NewClient()
	if err != nil {
		return nil, err
	}

	blocker := ipblocker.New(10, 30*time.Minute)

	router := &TLSHandler{
		Handlers: &Handler{
			UsersService:       usersService,
			MessengerService:   messengerService,
			FilesService:       filesService,
			AudioDialogService: audioDialogService,
			Blocker:            blocker,
		},
		PagesHandlers:           &PagesHandler{},
		Blocker:                 blocker,
		Auth:                    usersService,
		UsersUsedRequestsNonces: make(map[string]string),
	}

	router.PagesHandlers.readHTMLPages()

	return router, nil
}
