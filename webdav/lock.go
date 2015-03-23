package webdav

import (
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (api *API) lock(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	_, err := api.basicAuth(r)
	if err != nil {
		logrus.Error(err)
		w.Header().Set("WWW-Authenticate", "Basic Real='WhiteDAV credentials'")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "text/xml; charset=\"utf-8\"")
	w.Header().Set("Lock-Token", "opaquelocktoken:00000000-0000-0000-0000-000000000000")

	xml := `<?xml version="1.0" encoding="utf-8"?><prop xmlns="DAV:"><lockdiscovery><activelock><allprop/><timeout>Second-604800</timeout><depth>Infinity</depth><locktoken><href>opaquelocktoken:00000000-0000-0000-0000-000000000000</href></locktoken></activelock></lockdiscovery></prop>`

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xml))

	return
}
