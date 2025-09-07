package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/roshanlc/send-to-kindle/config"
	"github.com/roshanlc/send-to-kindle/internal/database"
	"github.com/roshanlc/send-to-kindle/internal/downloader"
	"github.com/roshanlc/send-to-kindle/internal/email"
	"github.com/roshanlc/send-to-kindle/internal/helper"
	"github.com/roshanlc/send-to-kindle/internal/queue"
	"resty.dev/v3"
)

func processTasks(config *config.ServerConfig, q *queue.TaskQueue, db *database.DB) {
	client := resty.New().
		SetRetryCount(3).
		SetTimeout(5 * time.Minute)

	defer client.Close() // clean up

	for {
		task := q.Dequeue()
		slog.Info("taking up task", slog.String("taskID", task.ID.String()))
		ctx := helper.NewContextWithUUID(context.Background(), task.ID)
		taskDB, err := db.GetTask(task.ID.String()) // get task item from db
		// TODO: add failure state to task and maybe re-add to queue again ?? also how to handle error of failure state update
		if err != nil {
			slog.Error("error occured while fetching task from db", slog.String("taskID", task.ID.String()), slog.String("error", err.Error()))
			continue
		}
		filename, ctx, err := downloader.Process(ctx, client, task.URL)

		if err != nil {
			slog.Error("error occured while downloading for task", slog.String("taskID", task.ID.String()), slog.String("error", err.Error()))
			continue
		}

		taskDB.Title = filename
		err = db.UpdateTask(taskDB)
		if err != nil {
			slog.Error("process failed while updating task title", slog.Any("taskID", task.ID.String()), slog.String("error", err.Error()))
		}

		slog.Info("attempting to email downloaded file", slog.Any("taskID", task.ID.String()))

		details := email.EmailDetails{
			From:        config.SmtpFrom,
			To:          config.SmtpTo,
			Host:        config.SmtpHost,
			Port:        config.SmtpPort,
			Subject:     helper.GetIDFromContext(ctx).String(),
			Body:        "Save the attached file(s).",
			Attachments: []string{helper.GetFilepathFromContext(ctx)},
			Username:    config.SmtpUserID,
			Password:    config.SmtpPassword,
		}

		err = email.Send(ctx, details)
		if err != nil {
			slog.Error("process failed while sending email", slog.Any("taskID", task.ID.String()), slog.String("error", err.Error()))
			taskDB.State = database.Failed
			err := db.UpdateTask(taskDB)
			if err != nil {
				slog.Error("process failed while updating task state to failure", slog.Any("taskID", task.ID.String()), slog.String("error", err.Error()))
			}
			continue
		}

		// update the status of task in DB
		taskDB.State = database.Completed
		err = db.UpdateTask(taskDB)
		if err != nil {
			slog.Error("process failed while updating task state to completion", slog.Any("taskID", task.ID.String()), slog.String("error", err.Error()))
		}

		// delete the file now
		path := helper.GetFilepathFromContext(ctx)
		slog.Info("attempting to deleted downloaded file", slog.Any("taskID", task.ID.String()), slog.String("filepath", path))
		err = downloader.DeleteDownloadedFile(path)
		if err != nil {
			slog.Error("error while deleting downloaded file", slog.Any("taskID", task.ID.String()), slog.String("error", err.Error()),
				slog.String("filepath", path))
		}
	}
}
