package notifications

type RegisterTokenRequest struct {
	ProfileID int    `json:"profile_id"`
	Token     string `json:"token"`
	Platform  string `json:"platform"` // "android", "ios", "web"
}
