package response

import "time"

type (
	HealthCheck struct {
		Status string `json:"status"`
	}

	TokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	UserData struct {
		ID        int        `json:"id"`
		Name      string     `json:"name"`
		Email     string     `json:"email"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at"`
	}

	CustomerData struct {
		ID        int        `json:"id"`
		Name      string     `json:"name"`
		Phone     string     `json:"phone"`
		Address   string     `json:"address"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at"`
	}
)
