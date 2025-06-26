package providers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

type MailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

type SendMailJob struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

var Mailer *gomail.Dialer
var MailConfigInstance *MailConfig

func RegisterMailer() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get mail configuration from environment variables
	host := os.Getenv("MAIL_HOST")
	portStr := os.Getenv("MAIL_PORT")
	username := os.Getenv("MAIL_USERNAME")
	password := os.Getenv("MAIL_PASSWORD")
	from := os.Getenv("MAIL_FROM_ADDRESS")
	fromName := os.Getenv("MAIL_FROM_NAME")

	// Convert port string to integer
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid MAIL_PORT: %s", portStr)
	}

	// Create mail configuration
	MailConfigInstance = &MailConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		FromName: fromName,
	}

	// Create mailer dialer
	Mailer = gomail.NewDialer(host, port, username, password)

	fmt.Printf("Mailer configured for %s:%d\n", host, port)
}

// SendMail sends an email using the configured mailer
func SendMail(to []string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", MailConfigInstance.FromName, MailConfigInstance.From))
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	return Mailer.DialAndSend(m)
}

// SendMailAsync sends an email asynchronously via queue
func SendMailAsync(to []string, subject, body string) error {
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

	// Send to queue with job type attribute
	attributes := map[string]string{
		"job_type": "send_mail",
	}

	return SendMessageWithAttributes(string(jobData), attributes)
}

// ProcessMailJobFromQueue processes a send mail job from the queue
func ProcessMailJobFromQueue(jobData []byte) error {
	var job SendMailJob
	if err := json.Unmarshal(jobData, &job); err != nil {
		return fmt.Errorf("failed to unmarshal job data: %v", err)
	}

	return SendMail(job.To, job.Subject, job.Body)
}
