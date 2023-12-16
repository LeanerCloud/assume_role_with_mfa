package main

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
