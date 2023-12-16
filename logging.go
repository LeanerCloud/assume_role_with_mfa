package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

func setupLogging(cacheDir string) error {
	logFile, err := os.OpenFile(filepath.Join(cacheDir, "aws_mfa_log.txt"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}
	defer func() {
		if cerr := logFile.Close(); cerr != nil {
			logger.Printf("Error closing log file: %v", cerr)
		}
	}()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags)
	return nil
}
