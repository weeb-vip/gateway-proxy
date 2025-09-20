package middlewares

import (
	"net/http"
)

// Tracer is a no-op middleware (DataDog removed)
func Tracer() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return h
	}
}
