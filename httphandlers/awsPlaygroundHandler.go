package httphandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"log"
	"net/http"
	"net/url"
	"quic_shell_server/db"
	"time"
)

func awsPlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	awsCredentials := db.GetAvailableCredentials()
	if awsCredentials == nil {
		http.Error(w, "No available AWS credentials", http.StatusInternalServerError)
		return
	}
	loginUrl := testAws(awsCredentials.AccountId, awsCredentials.Username, awsCredentials.AwsAccessKeyId, awsCredentials.AwsAccessSecret, 3600)
	// respond loginUrl in JSON format {"url":${loginUrl}}
	json.NewEncoder(w).Encode(map[string]string{"url": loginUrl})
}

func testAws(accountId, username, accessKeyId, secretAccessKey string, durationSeconds int) (loginUrl string) {
	// Load AWS configuration
	// Define your access key and secret key
	//accessKey := "your-access-key-id"
	//secretKey := "your-secret-access-key"

	// Load the AWS configuration
	//cfg, err := config.LoadDefaultConfig(context.TODO())
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""))),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an STS client
	stsClient := sts.NewFromConfig(cfg)

	// Assume a role
	//roleArn := "arn:aws:iam::807641583053:role/user-us-east-1" // Replace with your role ARN
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountId, username) // Replace with your role ARN
	//roleSessionName := "us-east-1"
	roleSessionName := username
	creds := stscreds.NewAssumeRoleProvider(stsClient, roleArn, func(o *stscreds.AssumeRoleOptions) {
		o.RoleSessionName = roleSessionName
		//o.Duration = 10800 // in seconds (3 hours)
		o.Duration = time.Second * time.Duration(durationSeconds)
	})

	// Retrieve the credentials
	value, err := creds.Retrieve(context.TODO())
	if err != nil {
		log.Fatalf("unable to retrieve credentials, %v", err)
	}

	// Create the URL-encoded JSON with credentials
	sessionJson, err := json.Marshal(map[string]string{
		"sessionId":    value.AccessKeyID,
		"sessionKey":   value.SecretAccessKey,
		"sessionToken": value.SessionToken,
	})
	if err != nil {
		log.Fatalf("error marshaling JSON: %v", err)
	}

	// Generate the sign-in URL
	signInTokenURL := fmt.Sprintf("https://signin.aws.amazon.com/federation?Action=getSigninToken&Session=%s", url.QueryEscape(string(sessionJson)))
	resp, err := http.Get(signInTokenURL)
	if err != nil {
		log.Fatalf("error getting sign-in token: %v", err)
	}
	defer resp.Body.Close()

	var respData struct {
		SigninToken string `json:"SigninToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		log.Fatalf("error decoding response: %v", err)
	}

	// Construct the final sign-in URL
	consoleURL := fmt.Sprintf("https://signin.aws.amazon.com/federation?Action=login&Issuer=Example.org&Destination=%s&SigninToken=%s",
		url.QueryEscape("https://console.aws.amazon.com/"),
		url.QueryEscape(respData.SigninToken),
	)

	fmt.Println("Console login URL:", consoleURL)
	return consoleURL
}
