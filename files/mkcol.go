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

func (api *API) mkcol(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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
	logrus.Info(fmt.Sprintf("api:files user:%s op:mkcol path:%s", authRes.Username, resource))
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
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(metaJSON)
}
