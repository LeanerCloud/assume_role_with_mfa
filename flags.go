package main

import "flag"

var (
	roleArn    string
	mfaArn     string
	awsProfile string
)

func initializeFlags() {
	flag.StringVar(&roleArn, "role-arn", "", "AWS Role ARN")
	flag.StringVar(&mfaArn, "mfa-arn", "", "MFA device ARN")
	flag.StringVar(&awsProfile, "profile", "default", "AWS Profile")
	flag.Parse()
}
