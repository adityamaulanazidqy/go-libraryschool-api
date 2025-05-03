package request_models

type ProfileUpdate struct {
	Id       int    `json:"id,omitempty"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
