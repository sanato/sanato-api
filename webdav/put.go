package webdav

import (
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/sanato/sanato-lib/storage"
	"net/http"
)

func (api *API) put(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

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

	/*
	   Content-Range is dangerous for PUT requests:  PUT per definition
	   stores a full resource.  draft-ietf-httpbis-p2-semantics-15 says
	   in section 7.6:
	     An origin server SHOULD reject any PUT request that contains a
	     Content-Range header field, since it might be misinterpreted as
	     partial content (or might be partial content that is being mistakenly
	     PUT as a full representation).  Partial content updates are possible
	     by targeting a separately identified resource with state that
	     overlaps a portion of the larger resource, or by using a different
	     method that has been specifically defined for partial updates (for
	     example, the PATCH method defined in [RFC5789]).
	   This clarifies RFC2616 section 9.6:
	     The recipient of the entity MUST NOT ignore any Content-*
	     (e.g. Content-Range) headers that it does not understand or implement
	     and MUST return a 501 (Not Implemented) response in such cases.
	   OTOH is a PUT request with a Content-Range currently the only way to
	   continue an aborted upload request and is supported by curl, mod_dav,
	   Tomcat and others.  Since some clients do use this feature which results
	   in unexpected behaviour (cf PEAR::HTTP_WebDAV_Client 1.0.1), we reject
	   all PUT requests with a Content-Range for now.
	*/
	if r.Header.Get("Content-Range") != "" {
		logrus.Error("PUT with Content-Range is not allowed.")
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}

	// Intercepting the Finder problem
	if r.Header.Get("X-Expected-Entity-Length") != "" {
		/*
		   Many webservers will not cooperate well with Finder PUT requests,
		   because it uses 'Chunked' transfer encoding for the request body.

		   The symptom of this problem is that Finder sends files to the
		   server, but they arrive as 0-length files in PHP.

		   If we don't do anything, the user might think they are uploading
		   files successfully, but they end up empty on the server. Instead,
		   we throw back an error if we detect this.

		   The reason Finder uses Chunked, is because it thinks the files
		   might change as it's being uploaded, and therefore the
		   Content-Length can vary.

		   Instead it sends the X-Expected-Entity-Length header with the size
		   of the file at the very start of the request. If this header is set,
		   but we don't get a request body we will fail the request to
		   protect the end-user.
		*/
		logrus.Warn("Intercepting the Finder problem. Content-Length:", r.Header.Get("Content-Length"), "X-Expected-Entity-Length:", r.Header.Get("X-Expected-Entity-Length"))
		/*
			        	TODO:
						// Only reading first byte
			            $firstByte = fread($body,1);
			            if (strlen($firstByte)!==1) {
			                throw new Exception\Forbidden('This server is not compatible with OS/X finder. Consider using a different WebDAV client or webserver.');
			            }

			            // The body needs to stay intact, so we copy everything to a
			            // temporary stream.

			            $newBody = fopen('php://temp','r+');
			            fwrite($newBody,$firstByte);
			            stream_copy_to_stream($body, $newBody);
			            rewind($newBody);

			            $body = $newBody;
		*/
	}
	meta, err := api.storageProvider.Stat(resource, false)
	if err != nil {
		// stat will fail if the file does not exists
		// in our case this is ok and we create a new file
		if storage.IsNotExistError(err) {
			err = api.storageProvider.PutFile(resource, r.Body, r.ContentLength)
			if err != nil {
				logrus.Error(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			meta, err = api.storageProvider.Stat(resource, false)
			if err == nil {
				w.Header().Set("ETag", meta.ETag)
			}
			w.WriteHeader(http.StatusCreated)

			return
		} else {
			logrus.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	if meta.IsCol {
		logrus.Warn("PUT is not allowed on non-files.")
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	err = api.storageProvider.PutFile(resource, r.Body, r.ContentLength)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	meta, err = api.storageProvider.Stat(resource, false)
	if err == nil {
		w.Header().Set("ETag", meta.ETag)
	}
	w.WriteHeader(http.StatusNoContent)

	return
}
