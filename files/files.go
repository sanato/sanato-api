package files

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"github.com/sanato/sanato-lib/auth"
	"github.com/sanato/sanato-lib/config"
	"github.com/sanato/sanato-lib/storage"
	"net/http"
	"strings"
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
	api.router.Handle("GET", "/files_get/*path", api.get)
	api.router.Handle("PUT", "/files_put/*path", api.put)
	api.router.Handle("POST", "/files_delete/*path", api.delete)
	api.router.Handle("POST", "/files_mkcol/*path", api.mkcol)
	api.router.Handle("GET", "/files_stat/*path", api.stat)
	api.router.Handle("POST", "/files_rename", api.move)
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

// tokenAuth validate a token, if inHeader is true it picks the token from the Authorization header.
// If not, it picks the token from the query parameter token
func (api *API) tokenAuth(r *http.Request, inHeader bool) (*auth.AuthResource, error) {
	config, err := api.configProvider.Parse()
	if err != nil {
		return nil, err
	}
	if inHeader == true {
		tokenHeader := strings.Split(r.Header.Get("Authorization"), " ")
		if len(tokenHeader) < 2 {
			return nil, errors.New("no token auth header")
		}
		token, err := jwt.Parse(string(tokenHeader[1]), func(token *jwt.Token) (key interface{}, err error) {
			return []byte(config.TokenSecret), nil
		})
		if err != nil {
			return nil, err
		}
		authRes := &auth.AuthResource{}
		authRes.Username = token.Claims["username"].(string)
		authRes.DisplayName = token.Claims["displayName"].(string)
		authRes.Email = token.Claims["email"].(string)
		return authRes, nil
	} else {
		tokenParam := r.URL.Query().Get("token")
		token, err := jwt.Parse(string(tokenParam), func(token *jwt.Token) (key interface{}, err error) {
			return []byte(config.TokenSecret), nil
		})
		if err != nil {
			return nil, err
		}
		authRes := &auth.AuthResource{}
		authRes.Username = token.Claims["username"].(string)
		authRes.DisplayName = token.Claims["displayName"].(string)
		authRes.Email = token.Claims["email"].(string)
		return authRes, nil
	}

}
