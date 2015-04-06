package webdav

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/sanato/sanato-lib/storage"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func (api *API) copy(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	authRes, err := api.basicAuth(r)
	if err != nil {
		logrus.Error(err)
		w.Header().Set("WWW-Authenticate", "Basic Real='WhiteDAV credentials'")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	resource := p.ByName("path")
	if resource == "" {
		resource = "/"
	}
	destination := r.Header.Get("Destination")
	overwrite := r.Header.Get("Overwrite")

	if destination == "" {
		logrus.Warn("the destination header was not supplied")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	destinationURL, err := url.ParseRequestURI(destination)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	destination = path.Join("/", strings.TrimPrefix(destinationURL.Path, "/webdav"))
	go logrus.Info(fmt.Sprintf("api:webdav user:%s op:move path:%s", authRes.Username, resource, destination))
	overwrite = strings.ToUpper(overwrite)
	if overwrite == "" {
		overwrite = "T"
	}
	if overwrite != "T" && overwrite != "F" {
		logrus.Warn("the HTTP Overwrite header should be either T or F")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	meta, err := api.storageProvider.Stat(destination, false)
	if err != nil {
		// if destination does not exists we are ok to continue independent of the
		// value of the overwrite header
		if storage.IsNotExistError(err) {
			err = api.storageProvider.Rename(resource, destination)
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

			w.WriteHeader(http.StatusCreated)

			return
		} else {
			logrus.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	// destination exists and overwrite is false so we should fail
	if overwrite == "F" {
		logrus.Warn("The destination node already exists, and the overwrite header is set to false")
		http.Error(w, http.StatusText(http.StatusPreconditionFailed), http.StatusPreconditionFailed)
		return
	}

	srcMeta, err := api.storageProvider.Stat(resource, false)
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

	// We dont support overriding of files with cols and viceversa
	if srcMeta.IsCol != meta.IsCol {
		logrus.Warn("we dont support overwrite of cols with files and viceversa")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	err = api.storageProvider.Copy(resource, destination)
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

	w.WriteHeader(http.StatusNoContent)

	return
}
