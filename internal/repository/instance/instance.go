package instance

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func GetConnection() *dynamodb.DynamoDB {
	config := aws.Config{}

	if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
		config.Endpoint = aws.String(endpoint)
		config.Region = aws.String(awsRegion())
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            config,
	}))

	return dynamodb.New(sess)
}

func awsRegion() string {
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}

	return "us-east-1"
}
