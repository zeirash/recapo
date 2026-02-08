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

	ProductData struct {
		ID            int         `json:"id"`
		Name          string      `json:"name"`
		Description   string      `json:"description"`
		Price         int         `json:"price"`
		OriginalPrice int         `json:"original_price"`
		CreatedAt     time.Time   `json:"created_at"`
		UpdatedAt     *time.Time  `json:"updated_at"`
	}

	OrderData struct {
		ID           int             `json:"id"`
		CustomerName string          `json:"customer_name"`
		TotalPrice   int             `json:"total_price"`
		Status       string          `json:"status"`
		Notes        string          `json:"notes"`
		OrderItems   []OrderItemData `json:"order_items,omitempty"`
		CreatedAt    time.Time       `json:"created_at"`
		UpdatedAt    *time.Time      `json:"updated_at"`
	}

	OrderItemData struct {
		ID          int        `json:"id"`
		OrderID     int        `json:"order_id,omitempty"`
		ProductName string     `json:"product_name"`
		Price       int        `json:"price"`
		Qty         int        `json:"qty"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   *time.Time `json:"updated_at"`
	}
)
