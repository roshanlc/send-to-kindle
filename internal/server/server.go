package server

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	"github.com/roshanlc/send-to-kindle/config"
	"github.com/roshanlc/send-to-kindle/internal/database"
	"github.com/roshanlc/send-to-kindle/internal/queue"
)

type Server struct {
	Config    *config.ServerConfig // reference to a server configuration
	DB        *database.DB         // reference to a db instance
	Templates *template.Template   // templates
	TaskQueue *queue.TaskQueue     // Queue
	mux       *http.ServeMux       // multiplexer
}

// NOTE: These should be corresponding to the files under templales folder
var Pages = map[string]string{
	"HomePage":         "dashboard-base.html",
	"HistoryPage":      "history.html",
	"SubmitPage":       "submit-form.html",
	"SubmitResultPage": "submit-result.html",
}

const (
	InternalServerError = "Something went wrong. Please try again later."
)

// Verify checks if all server configuration values are alright
func (s *Server) Verify() error {
	if s.Config == nil {
		return fmt.Errorf("Config field is nil")
	}
	if err := s.Config.Verify(); err != nil {
		return fmt.Errorf("Config values is not proper: %w", err)
	}
	if s.DB == nil {
		return fmt.Errorf("DB connection cannot be nil")
	}
	if err := s.DB.Database.Ping(); err != nil {
		return fmt.Errorf("Database ping failed: %w", err)
	}
	if s.Templates == nil {
		return fmt.Errorf("Tempaltes reference should be non-nil")
	}
	// check if necessary templates are added
	var allTemplates []string
	for _, t := range s.Templates.Templates() {
		allTemplates = append(allTemplates, t.Name())
	}
	all := strings.Join(allTemplates, ",")
	for _, val := range Pages {
		if !strings.Contains(all, val) {
			return fmt.Errorf("%s template is not found", val)
		}
	}

	return nil
}

func (s *Server) setupRouter() {

	mux := http.NewServeMux()
	// setup routes

	mux.HandleFunc("GET /", s.HomeHandler)
	mux.HandleFunc("GET /history", s.TaskListHandler)
	mux.HandleFunc("POST /submit", s.TaskAddHandler)
	mux.HandleFunc("DELETE /history/clear", s.TaskRemoveCompletedHandler)

	s.mux = mux
}

func (s *Server) Start() {
	// setup routes and stuff
	slog.Info("setting up router")
	s.setupRouter()

	slog.Info("starting server...", slog.String("port", s.Config.ServerPort))
	// start the server
	err := http.ListenAndServe(":"+s.Config.ServerPort, s.mux)
	if err != nil {
		slog.Error("error while starting server", slog.String("error", err.Error()))
		return
	}
}
