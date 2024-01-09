package db

import (
	"sync"
	"time"
)

type AwsCredentials struct {
	Username          string
	Password          string
	Region            string
	AccountId         string
	AwsAccessKeyId    string
	AwsAccessSecret   string
	TimeOfReservation time.Time
}

type isAvailable bool

var awsUsersPool = map[AwsCredentials]isAvailable{
	AwsCredentials{
		Username:        "user-us-east-1",
		Region:          "us-east-1",
		AccountId:       "807641583053",
		AwsAccessKeyId:  "AKIA3YCZQRXGXM7SMQUW",
		AwsAccessSecret: "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",
	}: true,
	AwsCredentials{
		Username:        "user-us-east-2",
		Region:          "us-east-2",
		AccountId:       "807641583053",
		AwsAccessKeyId:  "AKIA3YCZQRXGU2SYOZ3G",
		AwsAccessSecret: "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",
	}: true,
}

var awsUsersPoolMutex = &sync.Mutex{}

func GetAvailableCredentials() *AwsCredentials {
	awsUsersPoolMutex.Lock()
	defer awsUsersPoolMutex.Unlock()
	for credentials, isAvailable := range awsUsersPool {
		if isAvailable {
			awsUsersPool[credentials] = false
			return &credentials
		}
	}
	return nil
}

func FreeUpCredentials(credentials AwsCredentials) {
	awsUsersPoolMutex.Lock()
	defer awsUsersPoolMutex.Unlock()
	awsUsersPool[credentials] = true
}
