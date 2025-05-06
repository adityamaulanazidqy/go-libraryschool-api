package request_models

type ProfileUpdate struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}
