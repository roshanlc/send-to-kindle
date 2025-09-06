package server

import (
	"log/slog"
	"net/http"
)

// HomeHandler serves the homepage (dashboard)
func (s *Server) HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	err := s.Templates.ExecuteTemplate(w, Pages["HomePage"], map[string]any{
		"SendTo": s.Config.SmtpTo,
	})
	if err != nil {
		slog.Error("error while excuting home template", slog.String("error", err.Error()))
		http.Error(w, InternalServerError, http.StatusInternalServerError)
		return
	}
}
