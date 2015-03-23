package webdav

import (
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/sanato/sanato-lib/storage"
	"net/http"
)

func (api *API) options(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	_, err := api.basicAuth(r)
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

	meta, err := api.storageProvider.Stat(resource, false)
	if err != nil {
		if storage.IsNotExistError(err) {
			logrus.Error(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	allow := "OPTIONS, LOCK, GET, HEAD, POST, DELETE, PROPPATCH, COPY, MOVE, UNLOCK, PROPFIND"
	if !meta.IsCol {
		allow += ", PUT"
	}

	w.Header().Set("Allow", allow)
	w.Header().Set("DAV", "1, 2")
	w.Header().Set("MS-Author-Via", "DAV")
	//w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusOK)

	return
}
