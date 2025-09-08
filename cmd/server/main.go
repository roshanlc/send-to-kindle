package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/roshanlc/send-to-kindle/config"
	"github.com/roshanlc/send-to-kindle/internal/database"
	"github.com/roshanlc/send-to-kindle/internal/downloader"
	"github.com/roshanlc/send-to-kindle/internal/helper"
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
	dbConn, err := sql.Open("sqlite", fmt.Sprint(filepath.Join(config.DBPath, DBNAME), "?_busy_timeout=5000&_journal_mode=WAL"))
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

	// task queue
	q := queue.NewTaskQueue()

	// cookie store
	store := sessions.NewCookieStore([]byte(config.SecretKey))

	// start the server
	svr := server.Server{
		Config:      &config,
		DB:          db,
		Templates:   templates,
		TaskQueue:   q,
		CookieStore: store,
	}

	err = svr.Verify()
	if err != nil {
		slog.Error("error while setting up server", slog.String("error", err.Error()))
		return
	}

	// set download directory
	downloader.SetDownloadDirectory(config.STOREPATH)

	// fetch ongoing tasks from db and add to queue (remaining ones from last run)
	tasks, err := db.ListTask([]database.TaskState{database.Pending, database.Ongoing})
	if err != nil {
		slog.Error("error while fetching pending and ongoing tasks from db", slog.String("error", err.Error()))
		slog.Warn("skipping adding leftover tasks due to error")
	} else {
		fmt.Println("leftover tasks:", tasks)
	}

	for _, t := range tasks {
		u, err := helper.GetUUIDFromID(t.ID)
		if err == nil {
			// update corresponding taks status in db
			_ = db.UpdateTask(database.Task{
				ID:    t.ID,
				State: database.Pending,
			})
			ta := queue.NewTask(u, t.URL)
			q.Enqueue(ta)
		}
	}

	// run in waitgroup

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		slog.Info("spinned up a goroutine for server")

		defer wg.Done()
		svr.Start()
	}()
	go func() {
		slog.Info("spinned up a goroutine for task queue processing")
		defer wg.Done()
		processTasks(&config, q, db)
	}()

	wg.Wait()
	slog.Info("Exiting...")
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
	config.SmtpUserID = os.Getenv("SMTPUSERID")
	config.SmtpPassword = os.Getenv("SMTPPASSWORD")
	config.SmtpFrom = os.Getenv("SMTPFROM")
	config.ServerPort = os.Getenv("SERVERPORT")
	config.DBPath = os.Getenv("DBPATH")
	config.STOREPATH = os.Getenv("STOREPATH")
	config.Username = os.Getenv("USERNAME")
	config.Password = os.Getenv("PASSWORD")
	config.SecretKey = os.Getenv("SECRETKEY")

	to := os.Getenv("SMTPTO")
	var emails []string
	if to != "" {
		val := strings.Split(strings.TrimSpace(to), ",")
		for _, v := range val {
			emails = append(emails, strings.TrimSpace(v))
		}
	}

	config.SmtpTo = emails

	return config, nil
}
