package middleware

import (
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/handlers"
)

func WrapGzip(handler http.Handler) http.Handler {
	return gziphandler.GzipHandler(handler)
}

func WrapLogger(handler http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stderr, handler)
}
