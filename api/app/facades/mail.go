package facades

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/config"
)

// Mail sends an email synchronously
func Mail(to []string, subject, body string) error {
	return core.SendMail(to, subject, body)
}

// MailAsync sends an email asynchronously via the mail queue from config
func MailAsync(to []string, subject, body string) error {
	queueConfig := config.QueueConfig()
	queues := queueConfig["queues"].(map[string]interface{})
	queueName := queues["mail"].(string)
	return core.SendMailAsync(to, subject, body, queueName)
}
