package files

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/sanato/sanato-lib/storage"
	"net/http"
	"path/filepath"
)

func (api *API) put(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	authRes, err := api.tokenAuth(r, true)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	resource := filepath.Clean(p.ByName("path"))
	if resource == "" {
		resource = "/"
	}
	logrus.Info(fmt.Sprintf("api:files user:%s op:put path:%s", authRes.Username, resource))
	if r.Header.Get("Content-Range") != "" {
		logrus.Error("PUT with Content-Range is not allowed.")
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
	meta, err := api.storageProvider.Stat(resource, false)
	if err != nil {
		// stat will fail if the file does not exists
		// in our case this is ok and we create a new file
		if storage.IsNotExistError(err) {
			err = api.storageProvider.PutFile(resource, r.Body, r.ContentLength)
			if err != nil {
				logrus.Error(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			meta, err = api.storageProvider.Stat(resource, false)
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
			metaJSON, err := json.Marshal(meta)
			if err != nil {
				logrus.Error(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write(metaJSON)
			return
		} else {
			logrus.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	if meta.IsCol {
		logrus.Warn("PUT is not allowed on non-files.")
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}
	err = api.storageProvider.PutFile(resource, r.Body, r.ContentLength)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	meta, err = api.storageProvider.Stat(resource, false)
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
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(metaJSON)
	return
}
