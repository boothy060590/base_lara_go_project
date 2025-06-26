package UserTransformer

import "base_lara_go_project/app/models"

func Transform(user models.User) map[string]interface{} {
	roles := []string{}
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
	}
	return map[string]interface{}{
		"id":             user.ID,
		"first_name":     user.FirstName,
		"last_name":      user.LastName,
		"email":          user.Email,
		"mobile_number":  user.MobileNumber,
		"reset_password": user.ResetPassword,
		"roles":          roles,
	}
}
