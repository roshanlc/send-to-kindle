package config

import (
	"fmt"
	"os"
)

// holds the necessary configuration details for the server to operate
type ServerConfig struct {
	SmtpHost     string
	SmtpPort     int
	SmtpFrom     string
	SmtpUserID   string
	SmtpPassword string
	ServerPort   string
	DBPath       string // location to store sqlite db
}

// Verify checks the values
func (c *ServerConfig) Verify() error {

	if c.DBPath == "" {
		return fmt.Errorf("DBPATH is empty. Please provide a storage location.")
	} else {
		// check if location is proper

		_, err := os.Stat(c.DBPath)
		if err != nil {
			return err
		}
	}

	// TODO: add verification steps
	return nil
}
