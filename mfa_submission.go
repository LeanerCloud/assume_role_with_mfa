package main

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func handleMFASubmission(myWindow fyne.Window, roleArn, mfaSerialNumber, token string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	credentials, err := assumeRoleWithToken(ctx, roleArn, mfaSerialNumber, token)
	if err != nil {
		dialog.ShowError(err, myWindow) // Show error in a dialog box
		return
	}

	if err := cacheAWSRoleCredentials(credentials, roleArn); err != nil {
		logger.Println("Error caching credentials:", err)
		return
	}

	handleSubmissionResponse(myWindow, credentials)
}

func handleSubmissionResponse(myWindow fyne.Window, credentials *sts.AssumeRoleOutput) {
	// Logic to handle the response after successful submission.
	response := createCredentialResponse(credentials)

	// Convert the response to JSON
	jsonResponse, err := toJSON(response)
	if err != nil {
		// Log the error and return without closing the window
		logger.Printf("Error marshaling credentials to JSON: %v\n", err)
		return
	}

	// Print the successfully marshaled JSON response
	fmt.Println(jsonResponse)

	// Close the window
	myWindow.Close()
}
