package files

import (
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/sanato/sanato-lib/storage"
	"io"
	"net/http"
	"path/filepath"
)

func (api *API) get(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	_, err := api.tokenAuth(r, false)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	resource := filepath.Clean(p.ByName("path"))
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
		logrus.Error("GET is only implemented for file resources")
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
	reader, err := api.storageProvider.GetFile(resource)
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
	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(meta.Path))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, reader)
	return
}
