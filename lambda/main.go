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
	dryRunMode = "DRY_RUN_MODE"
	
	disableKeysToggleEnvVar = "DISABLE_KEYS_TOGGLE"
	disableKeysDaysEnvVar   = "DISABLE_KEYS_DAYS"
	defaultDisableKeysDays  = 30
	disableKeysErrorMessage = "Disabling keys feature is not enabled"
	
	disableOnlyUnusedKeysToggleEnvVar = "DISABLE_KEYS_TOGGLE"
	disableOnlyUnusedKeysDaysEnvVar   = "DISABLE_KEYS_DAYS"
	defaultOnlyUnusedDisableKeysDays  = 30
	disableOnlyUnusedKeysErrorMessage = "Disabling only unused keys feature is not enabled"
)

// For local testing
//func main() {
//	disableKeys(20)
//}

func handler(ctx context.Context, event events.CloudWatchEvent) error {
	disableeOnlyUnusedKeysToggle := EnvVarAsBool(disableeOnlyUnusedKeysToggleEnvVar)
	if disableeOnlyUnusedKeysToggle {
		disableeOnlyUnusedKeysDays := getEnvVarAsInt(disableeOnlyUnusedKeysDaysEnvVar, defaultDisableeOnlyUnusedKeysDays)
		disableeOnlyUnusedKeys(disableeOnlyUnusedKeysDays)
	} else {
		fmt.Println(disableeOnlyUnusedKeysErrorMessage)
	}
	
	disableKeysToggle := EnvVarAsBool(disableKeysToggleEnvVar)
	if disableKeysToggle {
		disableKeysDays := getEnvVarAsInt(disableKeysDaysEnvVar, defaultDisableKeysDays)
		disableAllKeys(disableKeysDays)
	} else {
		fmt.Println(disableKeysErrorMessage)
	}

	return nil
}

func disableAllKeys(days int) {
	fmt.Printf("Dry runmode =  %s\n", dryRunMode)
	
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
				
				if dryRunMode != true {
					_, err := svc.UpdateAccessKey(disableKeyInput)
					if err != nil {
						fmt.Println(err)
						continue
					}
				}
				
				fmt.Printf("Disabled access key %s for user %s\n", *key.AccessKeyId, *user.UserName)
			}
		}
	}
}

func disableInactiveCredentials(days int) {
	
	fmt.Printf("Dry runmode =  %s\n", dryRunMode)
	
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println(err)
		return
	}

	svc := iam.New(sess)

	input := &iam.ListUsersInput{}
	result, err := svc.ListUsers(input)
	if err != nil {
		log.Fatal(err)
	}

	// Disable inactive credentials for each user
	for _, user := range result.Users {
		username := *user.UserName
		fmt.Printf("Checking credentials for user: %s\n", username)

		// Get user's access keys
		keysInput := &iam.ListAccessKeysInput{
			UserName: user.UserName,
		}
		keysResult, err := svc.ListAccessKeys(keysInput)
		if err != nil {
			log.Fatal(err)
		}

		// Disable inactive access keys
		for _, key := range keysResult.AccessKeyMetadata {
			accessKeyID := *key.AccessKeyId

			// Get access key last used information
			lastUsedInput := &iam.GetAccessKeyLastUsedInput{
				AccessKeyId: key.AccessKeyId,
			}
			lastUsedResult, err := svc.GetAccessKeyLastUsed(lastUsedInput)
			if err != nil {
				log.Fatal(err)
			}

			// Check if access key is inactive for more than 30 days
			if lastUsedResult.AccessKeyLastUsed.LastUsedDate != nil {
				lastUsedDate := *lastUsedResult.AccessKeyLastUsed.LastUsedDate
				inactiveDays := time.Since(lastUsedDate).Hours() / 24

				if inactiveDays > days {
					// Disable the access key
					disableInput := &iam.UpdateAccessKeyInput{
						AccessKeyId:    aws.String(accessKeyID),
						Status:         aws.String("Inactive"),
						UserName:       user.UserName,
					}
					
					if dryRunMode != true {
						_, err := svc.UpdateAccessKey(disableKeyInput)
						if err != nil {
							log.Fatal(err)
						}
					}
					fmt.Printf("Disabled access key %s for user %s\n", accessKeyID, username)
				}
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
