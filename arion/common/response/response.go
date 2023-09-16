package response

type (
	HealthCheck struct {
		Status string `json:"status"`
	}

	TokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
)
