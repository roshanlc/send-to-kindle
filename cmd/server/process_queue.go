package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/roshanlc/send-to-kindle/config"
	"github.com/roshanlc/send-to-kindle/internal/database"
	"github.com/roshanlc/send-to-kindle/internal/downloader"
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
		ctx, err := downloader.Process(ctx, client, testURL)
		if err != nil {
			slog.Error("error occured while downloading for task", slog.String("taskID", task.ID.String()), slog.String("error", err.Error()))
			continue
		}

		slog.Info("attempting to email downloaded file", slog.Any("taskID", task.ID.String()))

		// details := email.EmailDetails{
		// 	From:        config.SmtpFrom,
		// 	To:          config.SmtpTo,
		// 	Host:        config.SmtpHost,
		// 	Port:        config.SmtpPort,
		// 	Subject:     helper.GetIDFromContext(ctx).String(),
		// 	Body:        "Save the attached file(s).",
		// 	Attachments: []string{helper.GetFilepathFromContext(ctx)},
		// 	Username:    config.SmtpUserID,
		// 	Password:    config.SmtpPassword,
		// }

		// err = email.Send(ctx, details)
		// if err != nil {
		// 	slog.Error("process failed while sending email", slog.Any("taskID", task.ID.String()), slog.String("error", err.Error()))
		// 	return
		// }
		err = client.Close()
		if err != nil {
			slog.Error("process failed during http client closure", slog.Any("taskID", task.ID.String()), slog.String("error", err.Error()))
			return
		}
	}
}
