package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/roshanlc/send-to-kindle/internal/database"
	"github.com/roshanlc/send-to-kindle/internal/helper"
	"github.com/roshanlc/send-to-kindle/internal/queue"
)

// TaskListHandler returns a list of tasks with details
func (s *Server) TaskListHandler(w http.ResponseWriter, r *http.Request) {

	var tasks []database.Task
	// data from history
	tasks, err := s.DB.ListTask("")

	if err != nil {
		slog.Error("error while fetching tasks list", slog.String("error", err.Error()))
	}

	w.WriteHeader(http.StatusOK)
	err = s.Templates.ExecuteTemplate(w, Pages["HistoryPage"], tasks)
	if err != nil {
		slog.Error("error while excuting history template", slog.String("error", err.Error()))
		http.Error(w, InternalServerError, http.StatusInternalServerError)
		return
	}
}

// TaskAddHandler adds a task to list
func (s *Server) TaskAddHandler(w http.ResponseWriter, r *http.Request) {
	var isValid bool = true
	var errMsg string
	var taskID uuid.UUID
	values := map[string]any{}
	// form parsing
	err := r.ParseForm()
	if err != nil {
		isValid = false
		errMsg = err.Error()
		values["isValid"] = isValid
		values["error"] = errMsg
		w.WriteHeader(http.StatusBadRequest)
		s.execSubmitResponse(values, w, r)
		return
	}

	url := strings.TrimSpace(r.Form.Get("url"))
	if url == "" {
		errMsg = "URL input should not be empty."
		isValid = false
		values["isValid"] = isValid
		values["error"] = errMsg
		w.WriteHeader(http.StatusBadRequest)
		s.execSubmitResponse(values, w, r)
		return
	}

	// TODO: also verify url thoroughly

	urls := strings.Split(url, ",")
	tIDs := make([]string, 0, len(urls))

	for _, u := range urls {
		u := strings.TrimSpace(u)
		if !helper.IsURLValid(u) {
			errMsg = "URL input should be valid."
			isValid = false
			values["isValid"] = isValid
			values["error"] = errMsg
			w.WriteHeader(http.StatusBadRequest)
			s.execSubmitResponse(values, w, r)
			return
		}

		// task additon to db and then to queue

		// generate new id
		taskID = helper.GenerateID()

		// add to database
		err = s.DB.AddTask(database.Task{
			ID:    taskID.String(),  // id of the task
			URL:   u,                // url of task
			State: database.Ongoing, // state of task
		})

		var task queue.Task
		if err != nil {
			isValid = false
			slog.Error("error while adding task to db", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			values["isValid"] = isValid
			values["error"] = errMsg
			s.execSubmitResponse(values, w, r)
			continue
		}
		task = queue.NewTask(taskID, u)
		// enqueue the task
		s.TaskQueue.Enqueue(task)
		tIDs = append(tIDs, taskID.String())
	}
	values["isValid"] = isValid
	values["error"] = errMsg
	values["taskID"] = tIDs
	w.WriteHeader(http.StatusOK)
	s.execSubmitResponse(values, w, r)
}

func (s *Server) execSubmitResponse(values map[string]any, w http.ResponseWriter, r *http.Request) {
	err := s.Templates.ExecuteTemplate(w, Pages["SubmitResultPage"], values)
	if err != nil {
		slog.Error("error while excuting submit-result template", slog.String("error", err.Error()))
		http.Error(w, InternalServerError, http.StatusInternalServerError)
		return
	}
}

// TaskAddHandler removes completed tasks from history
func (s *Server) TaskRemoveCompletedHandler(w http.ResponseWriter, r *http.Request) {
	err := s.DB.DeleteCompletedTasks()
	if err != nil {
		slog.Error("error while deleting completed tasks", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error while deleting completed tasks."))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task executed successfully."))
}
