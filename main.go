package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type CredentialProcessResponse struct {
	Version         int    `json:"Version"`
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

type CachedCredentials struct {
	Credentials CredentialProcessResponse `json:"credentials"`
}

var (
	roleArn    string
	mfaArn     string
	awsProfile string
)

func main() {
	initializeFlags()

	validateArns()

	myApp := startApplication()

	checkCredentialsAndLaunchGUI(myApp)
}

func initializeFlags() {
	flag.StringVar(&roleArn, "role-arn", "", "AWS Role ARN")
	flag.StringVar(&mfaArn, "mfa-arn", "", "MFA device ARN")
	flag.StringVar(&awsProfile, "profile", "default", "AWS Profile")
	flag.Parse()
}

func validateArns() {
	if !isValidArn(roleArn) || !isValidArn(mfaArn) {
		fmt.Println("Invalid ARN format")
		os.Exit(1)
	}

	if roleArn == "" || mfaArn == "" {
		fmt.Println("Usage: -role-arn <role_arn> -mfa-arn <mfa_device_arn> [-profile <source_aws_profile>]")
		os.Exit(1)
	}
}

func startApplication() fyne.App {
	myApp := app.NewWithID("com.leanercloud.aws-mfa-helper")
	return myApp
}

func checkCredentialsAndLaunchGUI(myApp fyne.App) {
	cacheDir := myApp.Storage().RootURI().Path()
	if err := setupLogging(cacheDir); err != nil {
		fmt.Println("Warning: Unable to set up logging:", err)
	}

	if creds, valid := checkCachedCredentials(roleArn); valid {
		fmt.Println(toJSON(creds))
	} else {
		launchGUI(myApp, roleArn, mfaArn)
	}
}

func setupLogging(cacheDir string) error {
	logFile, err := os.OpenFile(filepath.Join(cacheDir, "aws_mfa_log.txt"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags) // Set log format to include timestamps
	return nil
}

func getCachedCredentials(roleArn string) (CredentialProcessResponse, bool) {
	cacheFile, err := getCredentialCacheFile(roleArn)
	if err != nil {
		return CredentialProcessResponse{}, false
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return CredentialProcessResponse{}, false
	}

	var cachedCreds CachedCredentials
	if err := json.Unmarshal(data, &cachedCreds); err != nil {
		return CredentialProcessResponse{}, false
	}

	if creds := cachedCreds.Credentials; !isExpired(creds.Expiration) {
		return creds, true
	}
	return CredentialProcessResponse{}, false
}

func getCredentialCacheFile(roleArn string) (string, error) {
	// Generate SHA256 hash of the role ARN
	hasher := sha256.New()
	hasher.Write([]byte(roleArn))
	hash := hex.EncodeToString(hasher.Sum(nil))

	dir := fyne.CurrentApp().Storage().RootURI().Path()
	filename := fmt.Sprintf("aws_mfa_credentials_cache_%s.json", hash)
	return filepath.Join(dir, filename), nil
}

func checkCachedCredentials(roleArn string) (CredentialProcessResponse, bool) {

	return getCachedCredentials(roleArn)
}

func launchGUI(myApp fyne.App, roleArn, mfaSerialNumber string) {
	myWindow := setupGUIWindow(myApp, roleArn, mfaSerialNumber)
	myWindow.ShowAndRun()
}
func handleMFASubmission(myWindow fyne.Window, roleArn, mfaSerialNumber, token string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	credentials, err := assumeRoleWithToken(ctx, roleArn, mfaSerialNumber, token)
	if err != nil {
		dialog.ShowError(err, myWindow) // Show error in a dialog box
		return
	}

	response := createCredentialResponse(credentials)
	cacheCredentials(response, roleArn)
	fmt.Println(toJSON(response))
	myWindow.Close()
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
		widget.NewLabel("Enter MFA Token"),
		mfaInput,
		submitButton,
	)

	myWindow.SetContent(content)
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

func assumeRoleWithToken(ctx context.Context, roleArn, mfaSerialNumber, token string) (*sts.AssumeRoleOutput, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(awsProfile),
	)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS configuration: %w", err)
	}

	client := sts.NewFromConfig(cfg)
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("AssumedRoleSession"),
		SerialNumber:    aws.String(mfaSerialNumber),
		TokenCode:       aws.String(token),
	}

	return client.AssumeRole(ctx, input)
}

func createCredentialResponse(credentials *sts.AssumeRoleOutput) CredentialProcessResponse {
	return CredentialProcessResponse{
		Version:         1,
		AccessKeyId:     aws.ToString(credentials.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(credentials.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(credentials.Credentials.SessionToken),
		Expiration:      credentials.Credentials.Expiration.Format(time.RFC3339),
	}
}
func cacheCredentials(creds CredentialProcessResponse, roleArn string) {
	cacheFile, err := getCredentialCacheFile(roleArn)
	if err != nil {
		log.Println("Warning: Unable to get cache file path:", err)
		return
	}

	cachedCreds := CachedCredentials{Credentials: creds}
	data, err := json.Marshal(cachedCreds)
	if err != nil {
		log.Println("Warning: Unable to marshal credentials:", err)
		return
	}

	if err := ioutil.WriteFile(cacheFile, data, 0600); err != nil {
		log.Println("Warning: Unable to write credentials to cache file:", err)
	}
}

func isValidArn(arn string) bool {
	// This regex pattern is enhanced to match more AWS services and partitions
	// Note: This is a simplified pattern and might not cover all edge cases.
	// For comprehensive validation, you may need a more detailed pattern based on AWS ARN format specifications.
	// match, _ := regexp.MatchString(`arn:aws[a-zA-Z0-9-]*:[a-zA-Z0-9-]+:[a-zA-Z0-9-]*:\d{12}:[\w+=/,.@-]+`, arn) // this is broken
	return true
}
func isExpired(expiration string) bool {
	expiryTime, err := time.Parse(time.RFC3339, expiration)
	if err != nil {
		return true
	}
	return time.Now().After(expiryTime)
}

func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		log.Fatal("Error marshaling to JSON:", err)
	}
	return string(data)
}
