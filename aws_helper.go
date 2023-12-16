package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func assumeRoleWithToken(ctx context.Context, roleArn, mfaSerialNumber, token string) (*sts.AssumeRoleOutput, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(awsProfile))
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

func cacheAWSRoleCredentials(credentials *sts.AssumeRoleOutput, roleArn string) error {
	response := createCredentialResponse(credentials)
	cacheFile, err := getCredentialCacheFile(roleArn)
	if err != nil {
		return fmt.Errorf("unable to get cache file path: %w", err)
	}

	cachedCreds := CachedCredentials{Credentials: response}
	data, err := json.Marshal(cachedCreds)
	if err != nil {
		return fmt.Errorf("unable to marshal credentials: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0600); err != nil {
		return fmt.Errorf("unable to write credentials to cache file: %w", err)
	}

	return nil
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

func getCredentialCacheFile(roleArn string) (string, error) {
	// Generate SHA256 hash of the role ARN
	hasher := sha256.New()
	hasher.Write([]byte(roleArn))
	hash := hex.EncodeToString(hasher.Sum(nil))

	dir := fyne.CurrentApp().Storage().RootURI().Path()
	filename := fmt.Sprintf("aws_mfa_credentials_cache_%s.json", hash)
	return filepath.Join(dir, filename), nil
}

func getCachedCredentials(roleArn string) (CredentialProcessResponse, bool, error) {
	cacheFile, err := getCredentialCacheFile(roleArn)
	if err != nil {
		return CredentialProcessResponse{}, false, fmt.Errorf("error getting credential cache file: %w", err)
	}

	// Check if the cache file exists
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		// Cache file does not exist, no cached credentials
		return CredentialProcessResponse{}, false, nil
	}

	// Read the cache file
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return CredentialProcessResponse{}, false, fmt.Errorf("error reading cache file: %w", err)
	}

	var cachedCreds CachedCredentials
	if err := json.Unmarshal(data, &cachedCreds); err != nil {
		return CredentialProcessResponse{}, false, fmt.Errorf("error unmarshaling cached credentials: %w", err)
	}

	// Check if the cached credentials are expired
	if isExpired(cachedCreds.Credentials.Expiration) {
		return CredentialProcessResponse{}, false, nil
	}

	return cachedCreds.Credentials, true, nil
}
