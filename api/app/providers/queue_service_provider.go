package providers

import (
	"context"
	"fmt"
	"log"

	"base_lara_go_project/app/core"
	"base_lara_go_project/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joho/godotenv"
)

func RegisterQueue() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get queue configuration from config package
	queueConfig := config.QueueConfig()
	defaultQueue := queueConfig["default"].(string)
	connections := queueConfig["connections"].(map[string]interface{})
	connectionConfig := connections[defaultQueue].(map[string]interface{})

	accessKey := connectionConfig["key"].(string)
	secretKey := connectionConfig["secret"].(string)
	region := connectionConfig["region"].(string)
	queue := connectionConfig["queue"].(string)
	endpoint := connectionConfig["endpoint"].(string)

	// Create queue configuration
	queueConfigInstance := &core.QueueConfig{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
		Queue:     queue,
		Endpoint:  endpoint,
	}

	// Create custom AWS config for ElasticMQ
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           endpoint,
			SigningRegion: region,
		}, nil
	})

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
		awsconfig.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
			},
		}),
		awsconfig.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	// Create SQS client
	sqsClient := sqs.NewFromConfig(cfg)

	// Create queue if it doesn't exist
	createQueueIfNotExists(sqsClient, queue)

	// Create queue provider and set global instance
	queueProvider := core.NewQueueProvider(queueConfigInstance, sqsClient)
	core.SetQueueService(queueProvider)

	fmt.Printf("Queue service configured for %s (endpoint: %s)\n", queue, endpoint)
}

// createQueueIfNotExists creates the queue if it doesn't exist
func createQueueIfNotExists(client *sqs.Client, queueName string) {
	_, err := client.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})

	if err != nil {
		// If queue already exists, that's fine
		log.Printf("Queue creation result: %v (this is normal if queue already exists)", err)
	} else {
		log.Printf("Queue '%s' created successfully", queueName)
	}
}
