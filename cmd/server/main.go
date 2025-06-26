package main

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/roshanlc/send-to-kindle/internal/downloader"
	"github.com/roshanlc/send-to-kindle/internal/email"
	"github.com/roshanlc/send-to-kindle/internal/helper"
	"resty.dev/v3"
)

const testURL = "https://libgen.li/ads.php?md5=7e5412b8ece1fe49f7bfbc6e5ab77809" // stray birds by Tagore
func main() {

	// setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	client := resty.New().
		SetRetryCount(3).
		SetTimeout(5 * time.Minute)

	ctx := helper.GenerateIDWithContext()
	ctx, err := downloader.Process(ctx, client, testURL)
	if err != nil {
		logger.Error(err.Error())
	}

	godotenv.Load(".env")
	user := os.Getenv("USERID")
	pw := os.Getenv("PASSWORD")

	details := email.EmailDetails{
		From:        "",
		To:          "",
		Host:        "",
		Port:        587,
		Subject:     "Hello",
		Body:        "Receive this",
		Username:    user,
		Password:    pw,
		Attachments: []string{helper.GetFilepathFromContext(ctx)},
	}

	err = email.Send(ctx, details)
	if err != nil {
		slog.Error(err.Error())
	}

	log.Println("Exiting...")
}
