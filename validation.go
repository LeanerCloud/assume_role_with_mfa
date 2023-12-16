package main

import "regexp"

func isValidArn(arn string) bool {
	matched, err := regexp.MatchString(`arn:aws[a-zA-Z0-9-]*:[a-zA-Z0-9-]+:[a-zA-Z0-9-]*:\d{12}:[\w+=/,.@-]+`, arn)
	if err != nil {
		logger.Printf("Error validating ARN: %v", err)
		return false
	}
	return matched
}
