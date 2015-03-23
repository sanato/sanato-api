package files

import (
	"blacksync/lib/blacksync/storage"
	"path/filepath"
	"blacksync/lib/blacksync/auth"
	"blacksync/lib/blacksync/config"
	"blacksync/lib/blacksync/storage/storageproxy"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func Rename() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		log.WithFields(log.Fields{
			"frontend": "api",
			"app":      "files",
			"action":   "rename",
		}).Debug("files app rename resource")

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

		if err != nil {
			log.Warn(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		from := filepath.Clean(r.URL.Query().Get("from"))
		to := filepath.Clean(r.URL.Query().Get("to"))

		err = storageProxy.Rename(from, to)
		if err != nil {
			if storage.IsNotExistError(err) { 
				log.Warn(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	return http.HandlerFunc(fn)
}
