package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/roshanlc/send-to-kindle/internal/database"
	_ "modernc.org/sqlite"
)

const testURL = "https://libgen.li/ads.php?md5=7e5412b8ece1fe49f7bfbc6e5ab77809" // stray birds by Tagore
func main() {

	// setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbConn, err := sql.Open("sqlite", "./tmp/store.db")
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

	taskID := "sn90rn389dnd3893dn"
	// fmt.Println(db.AddTask(database.Task{
	// 	ID:    taskID,
	// 	URL:   "Random.com",
	// 	State: database.Ongoing,
	// }))

	t, err := db.GetTask(taskID)
	fmt.Println(t, err)

	fmt.Println("updating:->", db.UpdateTask(database.Task{
		ID:       taskID,
		URL:      "again.com",
		State:    database.Completed,
		ErrorMsg: "none",
	}))

	//
	t, err = db.GetTask(taskID)
	fmt.Println(t, err)

	log.Println("Exiting...")
}
