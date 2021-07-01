package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/greboid/fileshare"
)

func Auth(apiKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			headerKey := request.Header.Get("X-API-KEY")
			if headerKey == apiKey {
				next.ServeHTTP(writer, request)
				return
			}

			_, password, ok := request.BasicAuth()
			if !ok || password != apiKey {
				writer.Header().Add("WWW-Authenticate", `Basic realm="fileshare"`)
				writer.WriteHeader(401)
				return
			}

			next.ServeHTTP(writer, request)
		})
	}
}

func LoggingHandler(dst io.Writer) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(dst, h)
	}
}

func StripSlashes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" && strings.HasSuffix(request.URL.Path, "/") {
			http.Redirect(writer, request, strings.TrimSuffix(request.URL.Path, "/"), http.StatusPermanentRedirect)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func checkExpiry(db *fileshare.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			db.CheckFileName(r.URL.Path, "/raw/")
			next.ServeHTTP(w, r)
		})
	}
}
