package auth

// GetLoggedInUserJob handles retrieving logged in user
type GetLoggedInUserJob struct {
	UserID uint `json:"user_id"`
}

// Handle processes the get logged in user job
func (j *GetLoggedInUserJob) Handle() (any, error) {
	// TODO: Implement user retrieval using go_core services
	// For now, return placeholder data
	return map[string]interface{}{
		"message": "Get logged in user job processed",
		"user_id": j.UserID,
	}, nil
}
