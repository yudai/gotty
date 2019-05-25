package middleware

import (
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler"
)

func WrapGzip(handler http.Handler) http.Handler {
	return gziphandler.GzipHandler(handler)
}

func WrapLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		log.Printf("%s %d %s %s", r.RemoteAddr, 200, r.Method, r.URL.Path)
	})
}
