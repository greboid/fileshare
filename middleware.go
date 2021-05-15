package fileshare

import (
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
)

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
