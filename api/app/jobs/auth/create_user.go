package auth

// CreateUserJob handles user creation
type CreateUserJob struct {
	UserData  map[string]interface{} `json:"user_data"`
	RoleNames []string               `json:"role_names"`
}

// Handle processes the user creation job
func (j *CreateUserJob) Handle() (any, error) {
	// TODO: Implement user creation using go_core services
	// For now, return success
	return map[string]interface{}{
		"message":   "User creation job processed",
		"user_data": j.UserData,
		"roles":     j.RoleNames,
	}, nil
}
