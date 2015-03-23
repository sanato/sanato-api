package webdav

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/sanato/sanato-lib/auth"
	"github.com/sanato/sanato-lib/config"
	"github.com/sanato/sanato-lib/storage"
	"net/http"
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
	api.router.Handle("HEAD", "/webdav/*path", api.head)
	api.router.Handle("GET", "/webdav/*path", api.get)
	api.router.Handle("PUT", "/webdav/*path", api.put)
	api.router.Handle("DELETE", "/webdav/*path", api.delete)
	api.router.Handle("MKCOL", "/webdav/*path", api.mkcol)
	api.router.Handle("OPTIONS", "/webdav/*path", api.options)
	api.router.Handle("LOCK", "/webdav/*path", api.lock)
	api.router.Handle("UNLOCK", "/webdav/*path", api.unlock)
	api.router.Handle("PROPFIND", "/webdav/*path", api.propfind)
	api.router.Handle("COPY", "/webdav/*path", api.copy)
	api.router.Handle("MOVE", "/webdav/*path", api.move)
}

func (api *API) basicAuth(r *http.Request) (*auth.AuthResource, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, errors.New("no basic auth provided")
	}
	authRes, err := api.authProvider.Authenticate(username, password)
	if err != nil {
		return nil, err
	}
	return authRes, err
}
