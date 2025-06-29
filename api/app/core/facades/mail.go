package facades_core

import (
	mail_core "base_lara_go_project/app/core/mail"
	"base_lara_go_project/config"
)

// MailFacade provides a facade for mail operations
type MailFacade struct{}

// Send sends an email
func (m *MailFacade) Send(to []string, subject, body string) error {
	return mail_core.SendMail(to, subject, body)
}

// SendAsync sends an email asynchronously
func (m *MailFacade) SendAsync(to []string, subject, body string, queueName string) error {
	return mail_core.SendMailAsync(to, subject, body, queueName)
}

// Mail sends an email synchronously
func Mail(to []string, subject, body string) error {
	return mail_core.SendMail(to, subject, body)
}

// MailAsync sends an email asynchronously via the mail queue from config
func MailAsync(to []string, subject, body string) error {
	queueConfig := config.QueueConfig()
	queues := queueConfig["queues"].(map[string]interface{})
	queueName := queues["mail"].(string)
	return mail_core.SendMailAsync(to, subject, body, queueName)
}
