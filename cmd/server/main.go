package main

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/roshanlc/send-to-kindle/internal/downloader"
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
	err := downloader.Process(client, testURL)
	if err != nil {
		logger.Error(err.Error())
	}
	log.Println("Exiting...")
}
