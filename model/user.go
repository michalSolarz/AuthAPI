package model

type (
	User struct {
		ID       int `json:"id"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
		UUID     string `json:"uuid"`
	}
)
