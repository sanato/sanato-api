package webdav

import (
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (api *API) unlock(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	_, err := api.basicAuth(r)
	if err != nil {
		logrus.Error(err)
		w.Header().Set("WWW-Authenticate", "Basic Real='WhiteDAV credentials'")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	return
}
