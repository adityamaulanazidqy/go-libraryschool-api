package request_models

type UpdatePasswordRequest struct {
	UserID   int    `json:"user_id"`
	Password string `json:"password"`
}
