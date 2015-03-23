package webdav

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/whitedav/lib/whitedav/storage"
	"net/http"
)

func (api *API) head(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

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
			logrus.Warn(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if meta.IsCol {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", meta.Size))
	w.Header().Set("Last-Modified", fmt.Sprintf("%d", meta.Modified))
	w.Header().Set("ETag", meta.ETag)
	w.WriteHeader(http.StatusOK)

	return
}
