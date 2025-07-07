package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/roshanlc/send-to-kindle/internal/helper"
	"gopkg.in/yaml.v3"
)

const configPath = ".config/send-to-kindle/config.yaml" // from the home directory

// Config holds details about program configuration
type Config struct {
	Host         string   `yaml:"HOST"`
	Port         int      `yaml:"PORT"`
	From         string   `yaml:"FROM"`
	To           []string `yaml:"TO"`
	User         string   `yaml:"USERID"`
	Password     string   `yaml:"PASSWORD"`
	DownloadsDir string   `yaml:"DOWNLOADSDIR"`
}

func main() {
	// setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config, err := readConfig()
	if err != nil {
		slog.Error(err.Error())
		return //exit while running deferred functions
	}

	if len(os.Args) == 1 {
		slog.Error("no url provided, please provide a single url")
		fmt.Println("usage: ./send-to-kindle <url>")
		return
	}

	// extract  url from args
	url := extractURL(os.Args[1])
	if url == "" {
		slog.Error("please provide a valid url")
		return
	}

	slog.Info("extracted url", slog.String("url", url))

	// TODO: add this to database later
	// process the url
	process(config, url)
}

// extractURL takes value from arguments
func extractURL(rawURL string) string {
	if helper.IsURLValid(rawURL) {
		return rawURL
	}
	return ""
}

// readConfig attempts to read config file
func readConfig() (*Config, error) {
	var config Config

	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch user directory value: %w", err)

	}
	file, err := os.ReadFile(userDir + "/" + configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %w", err)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	// TODO: also check the validity of Config object
	// TODO: maybe move config related stuff to internal package
	return &config, nil
}
