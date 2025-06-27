package core

import (
	"encoding/json"
	"fmt"

	"gopkg.in/gomail.v2"
)

// MailConfig represents mail configuration
type MailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// SendMailJob represents a mail job for queue processing
type SendMailJob struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

// MailService defines the interface for mail operations
type MailService interface {
	SendMail(to []string, subject, body string) error
	SendMailAsync(to []string, subject, body string, queueName string) error
	ProcessMailJobFromQueue(jobData []byte) error
}

// MailProvider implements the MailService interface
type MailProvider struct {
	config *MailConfig
	mailer *gomail.Dialer
}

// NewMailProvider creates a new mail provider
func NewMailProvider(config *MailConfig, mailer *gomail.Dialer) *MailProvider {
	return &MailProvider{
		config: config,
		mailer: mailer,
	}
}

// SendMail sends an email using the configured mailer
func (m *MailProvider) SendMail(to []string, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", m.config.FromName, m.config.From))
	msg.SetHeader("To", to...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	return m.mailer.DialAndSend(msg)
}

// SendMailAsync sends an email asynchronously via queue
func (m *MailProvider) SendMailAsync(to []string, subject, body string, queueName string) error {
	// Create mail job data
	job := SendMailJob{
		To:      to,
		Subject: subject,
		Body:    body,
	}

	// Marshal job data
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %v", err)
	}

	// Send to queue with job type and queue name attribute
	attributes := map[string]string{
		"job_type": "send_mail",
		"queue":    queueName,
	}

	return SendMessageToQueueWithAttributes(string(jobData), attributes, queueName)
}

// ProcessMailJobFromQueue processes a send mail job from the queue
func (m *MailProvider) ProcessMailJobFromQueue(jobData []byte) error {
	var job SendMailJob
	if err := json.Unmarshal(jobData, &job); err != nil {
		return fmt.Errorf("failed to unmarshal job data: %v", err)
	}

	err := m.SendMail(job.To, job.Subject, job.Body)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// Global mail service instance
var MailServiceInstance MailService

// SetMailService sets the global mail service
func SetMailService(service MailService) {
	MailServiceInstance = service
}

// Helper functions for mail operations
func SendMail(to []string, subject, body string) error {
	return MailServiceInstance.SendMail(to, subject, body)
}

func SendMailAsync(to []string, subject, body string, queueName string) error {
	return MailServiceInstance.SendMailAsync(to, subject, body, queueName)
}

func ProcessMailJobFromQueue(jobData []byte) error {
	return MailServiceInstance.ProcessMailJobFromQueue(jobData)
}
