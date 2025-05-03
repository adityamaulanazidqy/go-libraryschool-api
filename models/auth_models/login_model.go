package auth_models

type LoginRequest struct {
	Email    string `json:"email" example:"user@gmail.com"`
	Password string `json:"password" example:"password"`
}

type LoginResponse struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created"`
}
