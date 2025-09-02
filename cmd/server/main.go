package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/roshanlc/send-to-kindle/config"
	"github.com/roshanlc/send-to-kindle/internal/database"
	"github.com/roshanlc/send-to-kindle/internal/queue"
	"github.com/roshanlc/send-to-kindle/internal/server"
	_ "modernc.org/sqlite"
)

const testURL = "https://libgen.li/ads.php?md5=7e5412b8ece1fe49f7bfbc6e5ab77809" // stray birds by Tagore

// server will use a single mail client and send from it to the clients
// just display the email from which it will be sent

var templates = template.Must(template.ParseGlob("templates/*.html"))

const DBNAME = "kindle-server.db"

func main() {
	// setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// config setup
	config, err := readConfig()
	if err != nil {
		slog.Error("error while reading config from .env", slog.String("error", err.Error()))
		return
	}

	err = config.Verify()
	if err != nil {
		slog.Error("error while verifying config values", slog.String("error", err.Error()))
		return
	}

	// database setup
	dbConn, err := sql.Open("sqlite", filepath.Join(config.DBPath, DBNAME))
	if err != nil {
		slog.Error(err.Error())
	} else {
		slog.Info("success at db connection")
		fmt.Println(dbConn.Ping())
	}
	defer dbConn.Close()

	db, err := database.New(dbConn)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	slog.Info("attempting to setup database")
	err = db.Setup()
	if err != nil {
		slog.Error(err.Error())
	} else {
		slog.Info("completed setup database")

	}

	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		slog.Error("error while generating random key", slog.String("error", err.Error()))
		return
	}

	// start the server
	svr := server.Server{
		Config:    &config,
		DB:        db,
		Templates: templates,
		TaskQueue: *queue.NewTaskQueue(),
	}

	err = svr.Verify()
	if err != nil {
		slog.Error("error while setting up server", slog.String("error", err.Error()))
		return
	}

	log.Println("Exiting...")
}

func readConfig() (config.ServerConfig, error) {
	config := config.ServerConfig{}
	err := godotenv.Load()
	if err != nil {
		return config, err
	}

	config.SmtpHost = os.Getenv("SMTPHOST")
	smtpPort, err := strconv.Atoi(os.Getenv("SMTPPORT"))
	if err != nil {
		return config, err
	}
	config.SmtpPort = smtpPort
	config.SmtpUserID = os.Getenv("USERID")
	config.SmtpPassword = os.Getenv("PASSWORD")
	config.SmtpFrom = os.Getenv("SMTPFROM")
	config.ServerPort = os.Getenv("SERVERPORT")
	config.DBPath = os.Getenv("DBPATH")

	return config, nil
}
