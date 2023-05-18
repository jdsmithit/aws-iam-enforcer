package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	disableKeysToggleEnvVar = "DISABLE_KEYS_TOGGLE"
	disableKeysDaysEnvVar   = "DISABLE_KEYS_DAYS"
	defaultDisableKeysDays  = 30
	disableKeysErrorMessage = "Disabling keys feature is not enabled"
)

// For local testing
//func main() {
//	disableKeys(20)
//}

func handler(ctx context.Context, event events.CloudWatchEvent) error {
	disableKeysToggle := EnvVarAsBool(disableKeysToggleEnvVar)
	if disableKeysToggle {
		disableKeysDays := getEnvVarAsInt(disableKeysDaysEnvVar, defaultDisableKeysDays)
		disableKeys(disableKeysDays)
	} else {
		fmt.Println(disableKeysErrorMessage)
	}

	return nil
}

func disableKeys(days int) {
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println(err)
		return
	}

	svc := iam.New(sess)

	usersOutput, err := svc.ListUsers(&iam.ListUsersInput{})
	if err != nil {
		fmt.Println(err)
		return
	}

	currentTime := time.Now()
	for _, user := range usersOutput.Users {
		accessKeysOutput, err := svc.ListAccessKeys(&iam.ListAccessKeysInput{
			UserName: user.UserName,
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, key := range accessKeysOutput.AccessKeyMetadata {
			if key.CreateDate.AddDate(0, 0, days).Before(currentTime) {
				disableKeyInput := &iam.UpdateAccessKeyInput{
					AccessKeyId: key.AccessKeyId,
					Status:      aws.String("Inactive"),
					UserName:    user.UserName,
				}
				_, err := svc.UpdateAccessKey(disableKeyInput)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("Disabled access key %s for user %s\n", *key.AccessKeyId, *user.UserName)
			}
		}
	}
}

func EnvVarAsBool(key string) bool {
	value := os.Getenv(key)
	if value == "" {
		return false
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}

	return boolValue
}

func getEnvVarAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}
