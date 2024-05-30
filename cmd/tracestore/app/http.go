package app

import (
	"net/http"

	"example.com/acme/kit/middleware"
	"github.com/klauspost/compress/gzhttp"
)

func httpGzipMiddleware() middleware.Interface {
	return middleware.Func(func(handler http.Handler) http.Handler {
		return gzhttp.GzipHandler(handler)
	})
}
