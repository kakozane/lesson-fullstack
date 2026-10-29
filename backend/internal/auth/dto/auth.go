package dto

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResult struct {
	AccessToken string `json:"accessToken"`
}
