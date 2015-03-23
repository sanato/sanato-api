package files

import (
	"blacksync/lib/blacksync/storage"
	"path/filepath"
	"blacksync/lib/blacksync/auth"
	"blacksync/lib/blacksync/config"
	"blacksync/lib/blacksync/storage/storageproxy"
	"net/http"
	"path"
	"strings"

	"io"

	log "github.com/Sirupsen/logrus"
)

func GetFile() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		log.WithFields(log.Fields{
		"app":      "api/v1/core/files",
		"action":   "GET",
		//"resource": resource,
		}).Info("")

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

		resource := path.Join("/", strings.TrimPrefix(r.URL.Path, "/api/v1/core/apps/files/get_file"))
		
		log.WithFields(log.Fields{
			"app":      "api/v1/core/webdav",
			"action":   "GET",
			"resource": resource,
		}).Info("")
		
		meta, err := storageProxy.Stat(resource, false)
		if err != nil {
			if storage.IsNotExistError(err){
				log.Warn(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if meta.IsCol {
			log.Error("GET is only implemented for file resources")
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		}

		reader, err := storageProxy.GetFile(resource)
		if err != nil {
			if storage.IsNotExistError(err){
				log.Warn(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", meta.MimeType)
		w.Header().Set("Content-Disposition", "attachment; filename=" + filepath.Base(meta.Path))

		w.WriteHeader(http.StatusOK)
		
		io.Copy(w, reader)

		return
	}
	return http.HandlerFunc(fn)
}
