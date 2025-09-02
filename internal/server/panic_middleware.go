package server

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// PanicMiddleware recovers from any panic in http handler goroutines
func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if r := recover(); r != nil {
				// log stacktrace internally
				slog.Error("encountered panic, review the stacktrace", slog.Any("stacktrace", debug.Stack()))
				slog.Info("recovered from panic")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}()
		next.ServeHTTP(w, r)
	})

}
