package webdav

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/sanato/sanato-lib/storage"
	"net/http"
)

func (api *API) delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

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

	go logrus.Info(fmt.Sprintf("api:webdav user:%s op:delete path:%s", authRes.Username, resource))

	_, err = api.storageProvider.Stat(resource, false)
	if err != nil { //DELETE on null resource gave 500, should be 404 (RFC2518:S3)
		if storage.IsNotExistError(err) {
			logrus.Warn(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = api.storageProvider.Remove(resource, true)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}
