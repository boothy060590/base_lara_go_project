package core

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// QueueConfig represents queue configuration
type QueueConfig struct {
	AccessKey string
	SecretKey string
	Region    string
	Queue     string
	Endpoint  string
}

// QueueService defines the interface for queue operations
type QueueService interface {
	SendMessage(messageBody string) error
	SendMessageToQueue(messageBody string, queueName string) error
	SendMessageWithAttributes(messageBody string, attributes map[string]string) error
	SendMessageToQueueWithAttributes(messageBody string, attributes map[string]string, queueName string) error
	ReceiveMessage() (*sqs.ReceiveMessageOutput, error)
	ReceiveMessageFromQueue(queueName string) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(receiptHandle string) error
	DeleteMessageFromQueue(receiptHandle string, queueName string) error
}

// QueueProvider implements the QueueService interface
type QueueProvider struct {
	config *QueueConfig
	client *sqs.Client
}

// NewQueueProvider creates a new queue provider
func NewQueueProvider(config *QueueConfig, client *sqs.Client) *QueueProvider {
	return &QueueProvider{
		config: config,
		client: client,
	}
}

// SendMessage sends a message to the default SQS queue
func (q *QueueProvider) SendMessage(messageBody string) error {
	_, err := q.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: aws.String(messageBody),
		QueueUrl:    aws.String(fmt.Sprintf("%s/queue/%s", q.config.Endpoint, q.config.Queue)),
	})
	return err
}

// SendMessageToQueue sends a message to a specific queue
func (q *QueueProvider) SendMessageToQueue(messageBody string, queueName string) error {
	_, err := q.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: aws.String(messageBody),
		QueueUrl:    aws.String(fmt.Sprintf("%s/queue/%s", q.config.Endpoint, queueName)),
	})
	return err
}

// SendMessageWithAttributes sends a message with custom attributes to the default queue
func (q *QueueProvider) SendMessageWithAttributes(messageBody string, attributes map[string]string) error {
	sqsAttributes := make(map[string]types.MessageAttributeValue)
	for key, value := range attributes {
		sqsAttributes[key] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
	}

	queueUrl := fmt.Sprintf("%s/queue/%s", q.config.Endpoint, q.config.Queue)

	_, err := q.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody:       aws.String(messageBody),
		MessageAttributes: sqsAttributes,
		QueueUrl:          aws.String(queueUrl),
	})

	if err != nil {
		log.Printf("Error sending message to queue: %v", err)
		return err
	}

	return nil
}

// SendMessageToQueueWithAttributes sends a message with custom attributes to a specific queue
func (q *QueueProvider) SendMessageToQueueWithAttributes(messageBody string, attributes map[string]string, queueName string) error {
	sqsAttributes := make(map[string]types.MessageAttributeValue)
	for key, value := range attributes {
		sqsAttributes[key] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
	}

	queueUrl := fmt.Sprintf("%s/queue/%s", q.config.Endpoint, queueName)

	_, err := q.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody:       aws.String(messageBody),
		MessageAttributes: sqsAttributes,
		QueueUrl:          aws.String(queueUrl),
	})

	if err != nil {
		log.Printf("Error sending message to queue %s: %v", queueName, err)
		return err
	}

	return nil
}

// ReceiveMessage receives a message from the default SQS queue
func (q *QueueProvider) ReceiveMessage() (*sqs.ReceiveMessageOutput, error) {
	queueUrl := fmt.Sprintf("%s/queue/%s", q.config.Endpoint, q.config.Queue)

	result, err := q.client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(queueUrl),
		MaxNumberOfMessages:   10,
		WaitTimeSeconds:       0,
		MessageAttributeNames: []string{"All"},
	})

	if err != nil {
		log.Printf("Error receiving messages: %v", err)
		return nil, err
	}

	return result, nil
}

// ReceiveMessageFromQueue receives a message from a specific queue
func (q *QueueProvider) ReceiveMessageFromQueue(queueName string) (*sqs.ReceiveMessageOutput, error) {
	queueUrl := fmt.Sprintf("%s/queue/%s", q.config.Endpoint, queueName)

	result, err := q.client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(queueUrl),
		MaxNumberOfMessages:   10,
		WaitTimeSeconds:       0,
		MessageAttributeNames: []string{"All"},
	})

	if err != nil {
		log.Printf("Error receiving messages from queue %s: %v", queueName, err)
		return nil, err
	}

	return result, nil
}

// DeleteMessage deletes a message from the default SQS queue
func (q *QueueProvider) DeleteMessage(receiptHandle string) error {
	_, err := q.client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(fmt.Sprintf("%s/queue/%s", q.config.Endpoint, q.config.Queue)),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}

// DeleteMessageFromQueue deletes a message from a specific queue
func (q *QueueProvider) DeleteMessageFromQueue(receiptHandle string, queueName string) error {
	_, err := q.client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(fmt.Sprintf("%s/queue/%s", q.config.Endpoint, queueName)),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}

// Global queue service instance
var QueueServiceInstance QueueService

// SetQueueService sets the global queue service
func SetQueueService(service QueueService) {
	QueueServiceInstance = service
}

// Helper functions for queue operations
func SendMessage(messageBody string) error {
	return QueueServiceInstance.SendMessage(messageBody)
}

func SendMessageToQueue(messageBody string, queueName string) error {
	return QueueServiceInstance.SendMessageToQueue(messageBody, queueName)
}

func SendMessageWithAttributes(messageBody string, attributes map[string]string) error {
	return QueueServiceInstance.SendMessageWithAttributes(messageBody, attributes)
}

func SendMessageToQueueWithAttributes(messageBody string, attributes map[string]string, queueName string) error {
	return QueueServiceInstance.SendMessageToQueueWithAttributes(messageBody, attributes, queueName)
}

func ReceiveMessage() (*sqs.ReceiveMessageOutput, error) {
	return QueueServiceInstance.ReceiveMessage()
}

func ReceiveMessageFromQueue(queueName string) (*sqs.ReceiveMessageOutput, error) {
	return QueueServiceInstance.ReceiveMessageFromQueue(queueName)
}

func DeleteMessage(receiptHandle string) error {
	return QueueServiceInstance.DeleteMessage(receiptHandle)
}

func DeleteMessageFromQueue(receiptHandle string, queueName string) error {
	return QueueServiceInstance.DeleteMessageFromQueue(receiptHandle, queueName)
}
