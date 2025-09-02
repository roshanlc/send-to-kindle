package server

import (
	"log/slog"
	"net/http"
)

// TaskListHandler returns a list of tasks with details
func (s *Server) TaskListHandler(w http.ResponseWriter, r *http.Request) {
	err := s.Templates.ExecuteTemplate(w, Pages["HistoryPage"], map[string]any{
		// values go here (key : val)
	})
	if err != nil {
		slog.Error("error while excuting history template", slog.String("error", err.Error()))
		http.Error(w, InternalServerError, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// TaskAddHandler adds a task to list
func (s *Server) TaskAddHandler(w http.ResponseWriter, r *http.Request) {
	err := s.Templates.ExecuteTemplate(w, Pages["SubmitResultPage"], map[string]any{
		"isValid": true, //should be dynamic later
	})

	if err != nil {
		slog.Error("error while excuting submit-result template", slog.String("error", err.Error()))
		http.Error(w, InternalServerError, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
