package files

import (
	"blacksync/lib/blacksync/storage"
	"blacksync/lib/blacksync/auth"
	"blacksync/lib/blacksync/config"
	"blacksync/lib/blacksync/storage/storageproxy"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func Stat() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		log.WithFields(log.Fields{
			"frontend": "api",
			"app":      "files",
			"action":   "stat",
		}).Debug("files app stat resource")

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

		var children bool
		queryChildren := r.URL.Query().Get("children")
		if queryChildren != "" {
			children, err = strconv.ParseBool(queryChildren)
			if err != nil {
				children = false
			}
		}

		if err != nil {
			log.Warn(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		resource := filepath.Join("/", strings.TrimPrefix(r.URL.Path, "/api/v1/core/apps/files/stat"))
		
		meta, err := storageProxy.Stat(resource, children)
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
		w.WriteHeader(http.StatusOK)
		w.Write(metaJSON)
		return
	}
	return http.HandlerFunc(fn)
}
