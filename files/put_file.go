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

func PutFile() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		log.WithFields(log.Fields{
			"frontend": "api",
			"app":      "files",
			"action":   "put_file",
		}).Debug("files app put file")

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

		resource := path.Join("/", strings.TrimPrefix(r.URL.Path, "/api/v1/core/apps/files/put_file"))
		
		if r.Header.Get("Content-Range") != "" {
			log.Error("PUT with Content-Range is not allowed.")
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		}

		meta, err := storageProxy.Stat(resource, false)
		if err != nil {
			// stat will fail if the file does not exists
			// in our case this is ok and we create a new file
			if storage.IsNotExistError(err) {
				err = storageProxy.PutFile(resource, r.Body, r.ContentLength, "", "")
				if err != nil {
					log.Error(err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				meta, err = storageProxy.Stat(resource, false)
				if err == nil {
					w.Header().Set("ETag", meta.ETag)
				}
				w.WriteHeader(http.StatusCreated)

				return
			} else {
				log.Error(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		if meta.IsCol {
			log.Warn("PUT is not allowed on non-files.")
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		err = storageProxy.PutFile(resource, r.Body, r.ContentLength, "", "")
		if err != nil {
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		meta, err = storageProxy.Stat(resource, false)
		if err == nil {
			w.Header().Set("ETag", meta.ETag)
		}
		w.WriteHeader(http.StatusNoContent)

		return
	}
	return http.HandlerFunc(fn)
}
