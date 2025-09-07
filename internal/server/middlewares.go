package server

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

var allowedOrigins = []string{
	"http://localhost:9009",
	"https://your-production-site.com",
}

func isAllowedOrigin(origin string) bool {
	for _, o := range allowedOrigins {
		if origin == o {
			return true
		}
	}
	return false
}

// Middleware to protect routes
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Don't protect the login and static assets
		// if r.URL.Path == "/login" || r.URL.Path == "/logout" {
		// 	next.ServeHTTP(w, r)
		// 	return
		// }
		session, err := s.CookieStore.Get(r, "session")

		if err != nil {
			slog.Error("error while excuting checking sessions", slog.String("error", err.Error()))
			http.Error(w, InternalServerError, http.StatusInternalServerError)
			return
		}

		auth, ok := session.Values["authenticated"].(bool)
		if (!ok || !auth) && r.URL.Path == "/login" {
			// if not logged in and hits login page, go to login page
			next.ServeHTTP(w, r)
			return
		} else if !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusMovedPermanently)
			return
		} else if auth && ok && r.URL.Path == "/login" {
			// if  logged in and hits login page, go to home page
			http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		}
		next(w, r)
	}
}

// PanicMiddleware recovers from any panic in http handler goroutines
func (s *Server) panicMiddleware(next http.HandlerFunc) http.HandlerFunc {
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
