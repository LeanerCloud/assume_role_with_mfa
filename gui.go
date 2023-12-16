package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// In gui.go or a similar file

func launchGUI(myApp fyne.App, roleArn, mfaSerialNumber string) {
	myWindow := setupGUIWindow(myApp, roleArn, mfaSerialNumber)
	myWindow.ShowAndRun()
}

func setupGUIWindow(myApp fyne.App, roleArn, mfaSerialNumber string) fyne.Window {
	myWindow := myApp.NewWindow("AWS MFA Input")

	mfaInput := createMFAInputEntry(func(token string) {
		handleMFASubmission(myWindow, roleArn, mfaSerialNumber, token)
	})
	submitButton := createSubmitButton(func() {
		handleMFASubmission(myWindow, roleArn, mfaSerialNumber, mfaInput.Text)
	})

	content := container.New(layout.NewVBoxLayout(),
		layout.NewSpacer(), // Spacer before the elements
		widget.NewLabel("Enter MFA Token"),
		mfaInput,
		layout.NewSpacer(), // Spacer before the elements
		submitButton,
		layout.NewSpacer(), // Spacer after the elements
	)

	myWindow.SetContent(content)

	// Set focus to the MFA input edit box
	myWindow.Canvas().Focus(mfaInput)

	return myWindow
}

func createMFAInputEntry(onSubmit func(string)) *widget.Entry {
	mfaInput := widget.NewEntry()
	mfaInput.SetPlaceHolder("xxxxxx")
	mfaInput.Validator = func(s string) error {
		if len(s) != 6 {
			return fmt.Errorf("MFA token must be 6 digits")
		}
		return nil
	}

	mfaInput.OnSubmitted = onSubmit

	return mfaInput
}

func createSubmitButton(onClick func()) *widget.Button {
	return widget.NewButton("Submit", onClick)
}
