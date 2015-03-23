package files

import (
	"blacksync/lib/blacksync/storage"
	"blacksync/lib/blacksync/auth"
	"blacksync/lib/blacksync/config"
	"blacksync/lib/blacksync/storage/storageproxy"
	"net/http"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func Remove() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		log.WithFields(log.Fields{
			"frontend": "api",
			"app":      "files",
			"action":   "remove",
		}).Debug("files app remove")

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
		resource := path.Join("/", strings.TrimPrefix(r.URL.Path, "/api/v1/core/apps/files/remove"))
		
		_, err = storageProxy.Stat(resource, false)
		if err != nil { //DELETE on null resource gave 500, should be 404 (RFC2518:S3)
			if storage.IsNotExistError(err){
				log.Warn(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}	
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = storageProxy.Remove(resource, true)
		if err != nil {
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}
	return http.HandlerFunc(fn)
}
