package gateway

import (
	"io"
	"net/http"
)

func CopyResponse(w http.ResponseWriter, r *http.Response) {
	for k, values := range r.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(r.StatusCode)
	_, _ = io.Copy(w, r.Body)
}
