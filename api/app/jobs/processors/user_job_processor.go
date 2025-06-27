package processors

import (
	"log"
)

// UserJobProcessor handles user-related job processing
type UserJobProcessor struct{}

// NewUserJobProcessor creates a new user job processor
func NewUserJobProcessor() *UserJobProcessor {
	return &UserJobProcessor{}
}

// CanProcess checks if this processor can handle the given job type
func (u *UserJobProcessor) CanProcess(jobType string) bool {
	return jobType == "user_created"
}

// Process processes a user created job
func (u *UserJobProcessor) Process(jobData []byte) error {
	// This would typically dispatch events or perform other user creation tasks
	log.Printf("Processing user created job: %s", string(jobData))

	// For now, we'll just log the job data
	// In a real implementation, this would parse the job data and perform specific actions
	// such as sending welcome emails, creating user profiles, etc.
	return nil
}
