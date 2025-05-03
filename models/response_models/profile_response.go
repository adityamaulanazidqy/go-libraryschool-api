package response_models

type ProfileResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleID   int    `json:"role_id"`
	Role     string `json:"role"`
}
