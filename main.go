package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	if err := run(); err != nil {
		logger.Printf("Fatal error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	initializeFlags()

	// Updated usage of isValidArn
	if !isValidArn(roleArn) || !isValidArn(mfaArn) {
		return fmt.Errorf("invalid ARN format")
	}

	myApp := startApplication()
	return checkCredentialsAndLaunchGUI(myApp)
}

func startApplication() fyne.App {
	// Initialize and return a new Fyne application
	myApp := app.NewWithID("com.leanercloud.aws-mfa-helper")
	return myApp
}

func checkCredentialsAndLaunchGUI(myApp fyne.App) error {
	// Get the cache directory path
	cacheDir := myApp.Storage().RootURI().Path()

	// Set up logging
	if err := setupLogging(cacheDir); err != nil {
		return fmt.Errorf("unable to set up logging: %w", err)
	}

	// Check for cached credentials
	creds, valid, err := getCachedCredentials(roleArn)
	if err != nil {
		return fmt.Errorf("error checking cached credentials: %w", err)
	}

	// If valid credentials are found, print them
	if valid {
		credsJSON, err := toJSON(creds)
		if err != nil {
			return fmt.Errorf("error marshaling credentials to JSON: %w", err)
		}
		fmt.Println(credsJSON)
	} else {
		// Otherwise, launch the GUI
		launchGUI(myApp, roleArn, mfaArn)
	}

	return nil
}
