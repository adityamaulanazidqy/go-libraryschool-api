package identity

type User struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	RoleID    int    `json:"role_id"`
	Role      string `json:"role"`
	CreatedAt string `json:"created"`
}
