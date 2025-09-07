package server

import (
	"log/slog"
	"net/http"
	"strings"
)

// showLoginPageHandler returns login page html template
func (s *Server) ShowLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// GET: show login page
	w.WriteHeader(http.StatusOK)
	err := s.Templates.ExecuteTemplate(w, Pages["LoginPage"], nil)
	if err != nil {
		slog.Error("error while excuting LoginPage template", slog.String("error", err.Error()))
		http.Error(w, InternalServerError, http.StatusInternalServerError)
		return
	}
}

// Login page and handler
func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		slog.Error("error while parsing form", slog.String("error", err.Error()))
		http.Error(w, InternalServerError, http.StatusInternalServerError)
		return
	}
	user := strings.TrimSpace(r.Form.Get("username"))
	pass := strings.TrimSpace(r.Form.Get("password"))

	if user == s.Config.Username && pass == s.Config.Password {
		session, _ := s.CookieStore.Get(r, "session")
		session.Values["authenticated"] = true
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	err = s.Templates.ExecuteTemplate(w, Pages["LoginPage"], "Invalid username or password")
	if err != nil {
		slog.Error("error while excuting LoginPage template", slog.String("error", err.Error()))
		http.Error(w, InternalServerError, http.StatusInternalServerError)
		return
	}
}

// Logout handler
func (s *Server) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := s.CookieStore.Get(r, "session")
	session.Values["authenticated"] = false
	session.Save(r, w)
	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}
