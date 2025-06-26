package facades

import (
	"base_lara_go_project/app/providers"
)

// Mail sends an email synchronously
func Mail(to []string, subject, body string) error {
	return providers.SendMail(to, subject, body)
}

// MailAsync sends an email asynchronously via queue
func MailAsync(to []string, subject, body string) error {
	return providers.SendMailAsync(to, subject, body)
}
