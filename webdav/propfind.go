package webdav

import (
	"encoding/xml"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/whitedav/lib/whitedav/storage"
	"net/http"
	"time"
)

func (api *API) propfind(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	_, err := api.basicAuth(r)
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

	var children bool
	depth := r.Header.Get("Depth")
	if depth == "1" {
		children = true
	}

	meta, err := api.storageProvider.Stat(resource, children)
	if err != nil {
		if storage.IsNotExistError(err) {
			logrus.Error(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responses := getPropFindFromMeta(meta)
	responsesXML, err := xml.Marshal(&responses)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")

	w.WriteHeader(207)
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><d:multistatus xmlns:d="DAV:">`))
	w.Write(responsesXML)
	w.Write([]byte(`</d:multistatus>`))

	return
}

func getPropFindFromMeta(meta *storage.MetaData) []ResponseXML {
	responses := []ResponseXML{}
	responses = append(responses, getResponseFromMeta(meta))

	if len(meta.Children) > 0 {
		for _, m := range meta.Children {
			responses = append(responses, getResponseFromMeta(m))
		}
	}

	return responses
}

func getResponseFromMeta(meta *storage.MetaData) ResponseXML {
	propList := []PropertyXML{}

	t := time.Unix(meta.Modified, 0)
	lasModifiedString := t.Format(time.RFC1123)

	getContentLegnth := PropertyXML{xml.Name{"", "d:getcontentlength"}, "", []byte(fmt.Sprintf("%d", meta.Size))}
	getLastModified := PropertyXML{xml.Name{"", "d:getlastmodified"}, "", []byte(lasModifiedString)}
	getETag := PropertyXML{xml.Name{"", "d:getetag"}, "", []byte(meta.ETag)}
	getContentType := PropertyXML{xml.Name{"", "d:getcontenttype"}, "", []byte(meta.MimeType)}
	if meta.IsCol {
		getResourceType := PropertyXML{xml.Name{"", "d:resourcetype"}, "", []byte("<d:collection/>")}
		getContentType.InnerXML = []byte("inode/directory")
		propList = append(propList, getResourceType)
	}
	propList = append(propList, getContentLegnth, getLastModified, getETag, getContentType)

	propStatList := []PropStatXML{}

	propStat := PropStatXML{}
	propStat.Prop = propList
	propStat.Status = "HTTP/1.1 200 OK"
	propStatList = append(propStatList, propStat)

	response := ResponseXML{}
	response.Href = "/webdav" + meta.Path
	response.Propstat = propStatList

	return response

}

type ResponseXML struct {
	XMLName             xml.Name      `xml:"d:response"`
	Href                string        `xml:"d:href"`
	Propstat            []PropStatXML `xml:"d:propstat"`
	Status              string        `xml:"d:status,omitempty"`
	Error               *ErrorXML     `xml:"d:error"`
	ResponseDescription string        `xml:"d:responsedescription,omitempty"`
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_propstat
type PropStatXML struct {
	// Prop requires DAV: to be the default namespace in the enclosing
	// XML. This is due to the standard encoding/xml package currently
	// not honoring namespace declarations inside a xmltag with a
	// parent element for anonymous slice elements.
	// Use of multistatusWriter takes care of this.
	Prop                []PropertyXML `xml:"d:prop>_ignored_"`
	Status              string        `xml:"d:status"`
	Error               *ErrorXML     `xml:"d:error"`
	ResponseDescription string        `xml:"d:responsedescription,omitempty"`
}

// Property represents a single DAV resource property as defined in RFC 4918.
// See http://www.webdav.org/specs/rfc4918.html#data.model.for.resource.properties
type PropertyXML struct {
	// XMLName is the fully qualified name that identifies this property.
	XMLName xml.Name

	// Lang is an optional xml:lang attribute.
	Lang string `xml:"xml:lang,attr,omitempty"`

	// InnerXML contains the XML representation of the property value.
	// See http://www.webdav.org/specs/rfc4918.html#property_values
	//
	// Property values of complex type or mixed-content must have fully
	// expanded XML namespaces or be self-contained with according
	// XML namespace declarations. They must not rely on any XML
	// namespace declarations within the scope of the XML document,
	// even including the DAV: namespace.
	InnerXML []byte `xml:",innerxml"`
}

// http://www.webdav.org/specs/rfc4918.html#ELEMENT_error
type ErrorXML struct {
	XMLName  xml.Name `xml:"d:error"`
	InnerXML []byte   `xml:",innerxml"`
}
