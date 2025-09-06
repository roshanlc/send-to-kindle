package main

import (
	"log/slog"
	"time"

	"github.com/roshanlc/send-to-kindle/internal/downloader"
	"github.com/roshanlc/send-to-kindle/internal/email"
	"github.com/roshanlc/send-to-kindle/internal/helper"
	"resty.dev/v3"
)

// process takes the url, downloads the file and emails it as an attachment
func process(config *Config, url string) {
	downloader.SetDownloadDirectory(config.DownloadsDir) // set the downloads directory
	ctx := helper.GenerateIDWithContext()
	slog.Info("trying to download file", slog.String("url", url), slog.Any("taskID", helper.GetIDFromContext(ctx)))

	client := resty.New().
		SetRetryCount(3).
		SetTimeout(5 * time.Minute)

	filename, ctx, err := downloader.Process(ctx, client, url)
	if err != nil {
		slog.Error("process failed", slog.String("error", err.Error()))
		return
	}

	slog.Info("attempting to email downloaded file:- "+filename, slog.Any("taskID", helper.GetIDFromContext(ctx)))

	details := email.EmailDetails{
		From:        config.From,
		To:          config.To,
		Host:        config.Host,
		Port:        config.Port,
		Subject:     helper.GetIDFromContext(ctx).String(),
		Body:        "Save the attached file(s).",
		Attachments: []string{helper.GetFilepathFromContext(ctx)},
		Username:    config.User,
		Password:    config.Password,
	}

	err = email.Send(ctx, details)
	if err != nil {
		slog.Error("process failed while sending email", slog.String("error", err.Error()))
		return
	}
	err = client.Close()
	if err != nil {
		slog.Error("process failed during http client closure", slog.String("error", err.Error()))
		return
	}
}
