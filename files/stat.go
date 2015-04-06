package files

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/sanato/sanato-lib/storage"
	"net/http"
	"path/filepath"
	"strconv"
)

func (api *API) stat(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	authRes, err := api.tokenAuth(r, true)
	if err != nil {
		go logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	resource := filepath.Clean(p.ByName("path"))
	if resource == "" {
		resource = "/"
	}
	logrus.Info(fmt.Sprintf("api:files user:%s op:stat path:%s", authRes.Username, resource))
	var children bool
	queryChildren := r.URL.Query().Get("children")
	if queryChildren != "" {
		children, err = strconv.ParseBool(queryChildren)
		if err != nil {
			children = false
		}
	}
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	meta, err := api.storageProvider.Stat(resource, children)
	if err != nil {
		if storage.IsNotExistError(err) {
			go logrus.Error(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		go logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		go logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(metaJSON)
	return
}
