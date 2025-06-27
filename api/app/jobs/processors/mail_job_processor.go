package processors

import (
	"base_lara_go_project/app/core"
)

// MailJobProcessor handles mail job processing
type MailJobProcessor struct{}

// NewMailJobProcessor creates a new mail job processor
func NewMailJobProcessor() *MailJobProcessor {
	return &MailJobProcessor{}
}

// CanProcess checks if this processor can handle the given job type
func (m *MailJobProcessor) CanProcess(jobType string) bool {
	return jobType == "send_mail"
}

// Process processes a mail job
func (m *MailJobProcessor) Process(jobData []byte) error {
	return core.ProcessMailJobFromQueue(jobData)
}
