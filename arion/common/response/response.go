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
		ID                 int        `json:"id"`
		Name               string     `json:"name"`
		Email              string     `json:"email"`
		Role               string     `json:"role"`
		SubscriptionActive bool       `json:"subscription_active,omitempty"`
		CreatedAt          time.Time  `json:"created_at"`
		UpdatedAt          *time.Time `json:"updated_at"`
	}

	CustomerData struct {
		ID        int        `json:"id"`
		Name      string     `json:"name"`
		Phone     string     `json:"phone"`
		Address   string     `json:"address"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at"`
	}

	// CustomerCheckActiveOrderByPhone is the response when checking active order by phone (get-or-create customer).
	CustomerCheckActiveOrderByPhone struct {
		CustomerID      int  `json:"customer_id"`
		ActiveOrderID   int  `json:"active_order_id"`
	}

	ProductData struct {
		ID            int        `json:"id"`
		Name          string     `json:"name"`
		Description   string     `json:"description"`
		Price         int        `json:"price"`
		OriginalPrice int        `json:"original_price"`
		ImageURL      string     `json:"image_url"`
		IsActive      bool       `json:"is_active"`
		CreatedAt     time.Time  `json:"created_at"`
		UpdatedAt     *time.Time `json:"updated_at"`
	}

	UploadImageData struct {
		ImageURL string `json:"image_url"`
	}

	OrderData struct {
		ID                int                `json:"id"`
		CustomerName      string             `json:"customer_name"`
		IsCustomerDeleted bool               `json:"is_customer_deleted"`
		TotalPrice        int                `json:"total_price"`
		Status            string             `json:"status"`
		PaymentStatus     string             `json:"payment_status"`
		Notes             string             `json:"notes"`
		OrderItems        []OrderItemData    `json:"order_items,omitempty"`
		OrderPayments     []OrderPaymentData `json:"order_payments,omitempty"`
		CreatedAt         time.Time          `json:"created_at"`
		UpdatedAt         *time.Time         `json:"updated_at"`
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

	OrderPaymentData struct {
		ID        int        `json:"id"`
		OrderID   int        `json:"order_id"`
		Amount    int        `json:"amount"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at"`
	}

	TempOrderData struct {
		ID             int                 `json:"id"`
		CustomerName   string              `json:"customer_name"`
		CustomerPhone  string              `json:"customer_phone"`
		TotalPrice     int                 `json:"total_price"`
		Status         string              `json:"status"`
		TempOrderItems []TempOrderItemData `json:"order_items,omitempty"`
		CreatedAt      time.Time           `json:"created_at"`
		UpdatedAt      *time.Time          `json:"updated_at"`
	}

	TempOrderItemData struct {
		ID          int        `json:"id"`
		TempOrderID int        `json:"temp_order_id,omitempty"`
		ProductID   int        `json:"product_id,omitempty"`
		ProductName string     `json:"product_name"`
		Price       int        `json:"price"`
		Qty         int        `json:"qty"`
		CreatedAt   time.Time  `json:"created_at"`
	}

	PurchaseListProductData struct {
		ProductName string `json:"product_name"`
		Price       int    `json:"price"`
		Qty         int    `json:"qty"`
	}

	PlanData struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		DisplayName   string `json:"display_name"`
		DescriptionEN string `json:"description_en"`
		DescriptionID string `json:"description_id"`
		PriceIDR      int    `json:"price_idr"`
		MaxUsers      int    `json:"max_users"`
	}

	SubscriptionData struct {
		ID                 int        `json:"id"`
		Status             string     `json:"status"`
		Plan               PlanData   `json:"plan"`
		TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
		CurrentPeriodStart time.Time  `json:"current_period_start"`
		CurrentPeriodEnd   time.Time  `json:"current_period_end"`
	}

	CheckoutData struct {
		OrderID     string `json:"order_id"`
		RedirectURL string `json:"redirect_url"`
		SnapToken   string `json:"snap_token"`
	}

	OrderStatsData struct {
		TotalRevenue int `json:"total_revenue"`
		NetSales     int `json:"net_sales"`
	}

	SystemStatsData struct {
		TotalShops    int `json:"total_shops"`
		SubsTrialing  int `json:"subs_trialing"`
		SubsActive    int `json:"subs_active"`
		SubsExpired   int `json:"subs_expired"`
		SubsCancelled int `json:"subs_cancelled"`
		MRRIDR        int `json:"mrr_idr"`
	}

	SystemShopData struct {
		ShopID      int        `json:"shop_id"`
		ShopName    string     `json:"shop_name"`
		OwnerName   string     `json:"owner_name"`
		OwnerEmail  string     `json:"owner_email"`
		PlanName    string     `json:"plan_name"`
		SubStatus   string     `json:"sub_status"`
		TrialEndsAt *time.Time `json:"trial_ends_at,omitempty"`
		PeriodEnd   time.Time  `json:"period_end"`
		JoinedAt    time.Time  `json:"joined_at"`
	}

	InvitationData struct {
		Email    string `json:"email"`
		ShopName string `json:"shop_name"`
	}

	SystemPaymentData struct {
		ShopName        string     `json:"shop_name"`
		PlanName        string     `json:"plan_name"`
		AmountIDR       int        `json:"amount_idr"`
		Status          string     `json:"status"`
		MidtransOrderID string     `json:"midtrans_order_id"`
		PaidAt          *time.Time `json:"paid_at,omitempty"`
		CreatedAt       time.Time  `json:"created_at"`
	}
)
