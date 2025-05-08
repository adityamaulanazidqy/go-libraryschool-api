package identity

type UserFavoriteBook struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleID   int    `json:"role_id"`
	Role     string `json:"role"`
}

type FavoriteBook struct {
	User UserFavoriteBook `json:"user"`
	Book Book             `json:"book"`
}
