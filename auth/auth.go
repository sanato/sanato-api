package auth

import (
	"github.com/julienschmidt/httprouter"
	"github.com/sanato/sanato-lib/auth"
	"github.com/sanato/sanato-lib/config"
	"github.com/sanato/sanato-lib/storage"
)

func NewAPI(router *httprouter.Router, cp *config.ConfigProvider, ap *auth.AuthProvider, sp *storage.StorageProvider) (*API, error) {
	return &API{router, cp, ap, sp}, nil
}

type API struct {
	router          *httprouter.Router
	configProvider  *config.ConfigProvider
	authProvider    *auth.AuthProvider
	storageProvider *storage.StorageProvider
}

func (api *API) Start() {
	api.router.Handle("POST", "/auth/login", api.login)
}
