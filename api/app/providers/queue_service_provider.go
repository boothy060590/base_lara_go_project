package providers

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/joho/godotenv"
)

type QueueConfig struct {
	AccessKey string
	SecretKey string
	Region    string
	Queue     string
	Endpoint  string
}

var SQSClient *sqs.Client
var QueueConfigInstance *QueueConfig

func RegisterQueue() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get queue configuration from environment variables
	accessKey := os.Getenv("SQS_ACCESS_KEY")
	secretKey := os.Getenv("SQS_SECRET_KEY")
	region := os.Getenv("SQS_REGION")
	queue := os.Getenv("SQS_QUEUE")
	endpoint := os.Getenv("SQS_ENDPOINT")

	// Create queue configuration
	QueueConfigInstance = &QueueConfig{
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

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
			},
		}),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	// Create SQS client
	SQSClient = sqs.NewFromConfig(cfg)

	// Create queue if it doesn't exist
	createQueueIfNotExists()

	fmt.Printf("Queue service configured for %s (endpoint: %s)\n", queue, endpoint)
}

// createQueueIfNotExists creates the queue if it doesn't exist
func createQueueIfNotExists() {
	_, err := SQSClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: aws.String(QueueConfigInstance.Queue),
	})

	if err != nil {
		// If queue already exists, that's fine
		log.Printf("Queue creation result: %v (this is normal if queue already exists)", err)
	} else {
		log.Printf("Queue '%s' created successfully", QueueConfigInstance.Queue)
	}
}

// SendMessage sends a message to the SQS queue
func SendMessage(messageBody string) error {
	_, err := SQSClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: aws.String(messageBody),
		QueueUrl:    aws.String(fmt.Sprintf("%s/queue/%s", QueueConfigInstance.Endpoint, QueueConfigInstance.Queue)),
	})
	return err
}

// SendMessageWithAttributes sends a message with custom attributes
func SendMessageWithAttributes(messageBody string, attributes map[string]string) error {
	log.Printf("Sending message to queue %s: %s", QueueConfigInstance.Queue, messageBody)
	log.Printf("Message attributes: %+v", attributes)

	sqsAttributes := make(map[string]types.MessageAttributeValue)
	for key, value := range attributes {
		sqsAttributes[key] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
		log.Printf("Setting attribute %s = %s", key, value)
	}

	queueUrl := fmt.Sprintf("%s/queue/%s", QueueConfigInstance.Endpoint, QueueConfigInstance.Queue)
	log.Printf("Queue URL: %s", queueUrl)

	_, err := SQSClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody:       aws.String(messageBody),
		MessageAttributes: sqsAttributes,
		QueueUrl:          aws.String(queueUrl),
	})

	if err != nil {
		log.Printf("Error sending message to queue: %v", err)
		return err
	}

	log.Printf("Message sent successfully to queue %s", QueueConfigInstance.Queue)
	return nil
}

// ReceiveMessage receives a message from the SQS queue
func ReceiveMessage() (*sqs.ReceiveMessageOutput, error) {
	queueUrl := fmt.Sprintf("%s/queue/%s", QueueConfigInstance.Endpoint, QueueConfigInstance.Queue)
	log.Printf("Receiving messages from queue: %s", queueUrl)

	result, err := SQSClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
	})

	if err != nil {
		log.Printf("Error receiving messages: %v", err)
		return nil, err
	}

	log.Printf("Received %d messages from queue", len(result.Messages))
	return result, nil
}

// DeleteMessage deletes a message from the SQS queue
func DeleteMessage(receiptHandle string) error {
	_, err := SQSClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(fmt.Sprintf("%s/queue/%s", QueueConfigInstance.Endpoint, QueueConfigInstance.Queue)),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}
