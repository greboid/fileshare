package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/goji/httpauth"
	"github.com/gorilla/handlers"
	"github.com/greboid/fileshare"
)

func authFunc(key string) func(string, string, *http.Request) bool {
	if key == "" {
		return func(string, string, *http.Request) bool {
			return true
		}
	}
	return func(_ string, password string, request *http.Request) bool {
		headerKey := request.Header.Get("X-API-KEY")
		if headerKey == key || password == key {
			return true
		}
		return false
	}
}

func Auth(apiKey string) func(handler http.Handler) http.Handler {
	return httpauth.BasicAuth(httpauth.AuthOptions{
		AuthFunc: authFunc(apiKey),
		Realm:    "fileshare",
	})
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
