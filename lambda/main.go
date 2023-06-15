package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	dryRunModeEnvVar = "DRY_RUN_MODE"

	disableKeysToggleEnvVar = "DISABLE_KEYS_TOGGLE"
	disableKeysDaysEnvVar   = "DISABLE_KEYS_DAYS"
	defaultDisableKeysDays  = 30
	disableKeysErrorMessage = "Disabling keys feature is not enabled"

	disableOnlyUnusedKeysToggleEnvVar = "DISABLE_UNUSED_ONLY_KEYS_TOGGLE"
	disableOnlyUnusedKeysDaysEnvVar   = "DISABLE_UNUSED_ONLY_KEYS_DAYS"
	defaultOnlyUnusedKeysDays         = 30
	disableOnlyUnusedKeysErrorMessage = "Disabling only unused keys feature is not enabled"
)

// For local testing
func main() {
	disableAllKeys(20, true)
	disableInactiveKeys(40, false)
}

func handler(ctx context.Context, event events.CloudWatchEvent) error {
	dryRunMode := EnvVarAsBool(dryRunModeEnvVar)
	fmt.Printf("Dry runmode =  %t\n", dryRunMode)

	disableeOnlyUnusedKeysToggle := EnvVarAsBool(disableOnlyUnusedKeysToggleEnvVar)
	if disableeOnlyUnusedKeysToggle {
		disableeOnlyUnusedKeysDays := getEnvVarAsInt(disableOnlyUnusedKeysDaysEnvVar, defaultOnlyUnusedKeysDays)
		disableInactiveKeys(disableeOnlyUnusedKeysDays, dryRunMode)
	} else {
		fmt.Println(disableOnlyUnusedKeysErrorMessage)
	}

	disableKeysToggle := EnvVarAsBool(disableKeysToggleEnvVar)
	if disableKeysToggle {
		disableKeysDays := getEnvVarAsInt(disableKeysDaysEnvVar, defaultDisableKeysDays)
		disableAllKeys(disableKeysDays, dryRunMode)
	} else {
		fmt.Println(disableKeysErrorMessage)
	}

	return nil
}

func disableAllKeys(keyAgeInDays int, dryRunMode bool) {
	svc, shouldReturn := CreateNewIamClientSession()

	if shouldReturn {
		return
	}

	usersOutput, err := svc.ListUsers(&iam.ListUsersInput{})

	if err != nil {
		fmt.Println(err)
		return
	}

	currentTime := time.Now()
	for _, user := range usersOutput.Users {
		accessKeysOutput := getAccessKeysForUser(user, svc)

		for _, key := range accessKeysOutput.AccessKeyMetadata {
			if key.CreateDate.AddDate(0, 0, keyAgeInDays).Before(currentTime) {
				markKeysAsDisabled(key, user, dryRunMode, svc)
			}
		}
	}
}

func disableInactiveKeys(keyInactivityPeriodInDays int, dryRunMode bool) {
	svc, shouldReturn := CreateNewIamClientSession()

	if shouldReturn {
		return
	}

	usersOutput, err := svc.ListUsers(&iam.ListUsersInput{})

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, user := range usersOutput.Users {
		accessKeysOutput := getAccessKeysForUser(user, svc)

		for _, key := range accessKeysOutput.AccessKeyMetadata {

			lastUsedInput := &iam.GetAccessKeyLastUsedInput{
				AccessKeyId: key.AccessKeyId,
			}
			lastUsedResult, err := svc.GetAccessKeyLastUsed(lastUsedInput)
			if err != nil {
				log.Fatal(err)
			}

			if lastUsedResult.AccessKeyLastUsed.LastUsedDate != nil {
				lastUsedDate := *lastUsedResult.AccessKeyLastUsed.LastUsedDate
				inactiveDays := time.Since(lastUsedDate).Hours() / 24

				if inactiveDays > float64(keyInactivityPeriodInDays) {
					markKeysAsDisabled(key, user, dryRunMode, svc)
				}
			}
		}
	}
}

func CreateNewIamClientSession() (*iam.IAM, bool) {
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println(err)
		return nil, true
	}

	svc := iam.New(sess)

	return svc, false
}

func getAccessKeysForUser(user *iam.User, svc *iam.IAM) *iam.ListAccessKeysOutput {
	username := *user.UserName
	fmt.Printf("Checking credentials for user: %s\n", username)

	keysInput := &iam.ListAccessKeysInput{
		UserName: user.UserName,
	}
	accessKeysOutput, err := svc.ListAccessKeys(keysInput)
	if err != nil {
		log.Fatal(err)
	}
	return accessKeysOutput
}

func markKeysAsDisabled(key *iam.AccessKeyMetadata, user *iam.User, dryRunMode bool, svc *iam.IAM) {
	disableKeyInput := &iam.UpdateAccessKeyInput{
		AccessKeyId: key.AccessKeyId,
		Status:      aws.String("Inactive"),
		UserName:    user.UserName,
	}

	if dryRunMode != true {
		_, err := svc.UpdateAccessKey(disableKeyInput)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Printf("Disabled access key %s for user %s\n", *key.AccessKeyId, *user.UserName)
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
