package webdav

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/sanato/sanato-lib/storage"
	"net/http"
)

func (api *API) mkcol(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	authRes, err := api.basicAuth(r)
	if err != nil {
		logrus.Error(err)
		w.Header().Set("WWW-Authenticate", "Basic Real='WhiteDAV credentials'")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	resource := p.ByName("path")
	if resource == "" {
		resource = "/"
	}
	logrus.Info(fmt.Sprintf("api:webdav user:%s op:mkcol path:%s", authRes.Username, resource))
	if r.ContentLength > 0 { // MKCOL with weird body must fail with 415 (RFC2518:8.3.1)
		// we dont accept mkcol with body, this is against the estandar
		logrus.Error(errors.New("we do not accept MKCOL with body"))
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	err = api.storageProvider.CreateCol(resource, false)
	if err != nil {
		if storage.IsNotExistError(err) {
			logrus.Error(err)
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		if storage.IsExistError(err) {
			logrus.Error(err)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	return
	return
}
