package dto

type LoginRequest struct {
	Username          string             `json:"username"`
	Password          string             `json:"password"`
	DeviceFingerprint *DeviceFingerprint `json:"device_fingerprint,omitempty"`
}

type DeviceFingerprint struct {
	UserAgent  string            `json:"user_agent"`
	Attributes map[string]string `json:"attributes"`
}

type LoginResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresIn    int             `json:"expires_in"`
	Account      AccountResponse `json:"account"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}
