package webdav

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (api *API) unlock(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	authRes, err := api.basicAuth(r)
	if err != nil {
		logrus.Error(err)
		w.Header().Set("WWW-Authenticate", "Basic Real='WhiteDAV credentials'")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	logrus.Info(fmt.Sprintf("api:webdav user:%s op:unlock", authRes.Username))
	w.WriteHeader(http.StatusNoContent)
	return
}
