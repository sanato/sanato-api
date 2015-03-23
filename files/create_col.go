package files

import (
	"encoding/json"
	"blacksync/lib/blacksync/storage"
	"blacksync/lib/blacksync/auth"
	"blacksync/lib/blacksync/config"
	"blacksync/lib/blacksync/storage/storageproxy"
	"net/http"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func CreateCol() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		log.WithFields(log.Fields{
			"version":  "v1",
			"app_type": "core",
			"app":      "files",
			"action":   "create_col",
		}).Debug("files app create col")

		cfg, err := config.Get()
		if err != nil {
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		authRes := auth.GetAuthResourceFromRequest(r)

		storageProxy, err := storageproxy.NewStorageProxy(cfg.Storage, authRes)
		if err != nil {
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		resource := path.Join("/", strings.TrimPrefix(r.URL.Path, "/api/v1/core/apps/files/create_col"))

		err = storageProxy.CreateCol(resource, false)
		if err != nil {
			if storage.IsNotExistError(err){
				log.Error(err)
				http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
				return
			}
			if storage.IsExistError(err) {
				log.Error(err)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		meta, err := storageProxy.Stat(resource, false)
		if err != nil {
			if storage.IsNotExistError(err){
				log.Error(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		metaJSON, err := json.Marshal(meta)
		if err != nil {
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Write(metaJSON)
		w.WriteHeader(http.StatusCreated)
	}
	return http.HandlerFunc(fn)
}
