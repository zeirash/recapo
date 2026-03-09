package model

import (
	"database/sql"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

/********************* Token ************************/
type (
	JwtCustomClaims struct {
		Name       string `json:"name"`
		UserID     int    `json:"user_id"`
		ShopID     int    `json:"shop_id"`
		SystemMode bool   `json:"system_mode"`
		jwt.RegisteredClaims
	}

	JwtCustomRefreshClaims struct {
		UserID     int  `json:"user_id"`
		ShopID     int  `json:"shop_id"`
		SystemMode bool `json:"system_mode"`
		jwt.RegisteredClaims
	}

	TokenData struct {
		Name       string `json:"name"`
		UserID     int    `json:"user_id"`
		ShopID     int    `json:"shop_id"`
		SystemMode bool   `json:"system_mode"`
	}

	/****************** Shared filter *****************/
	// FilterOptions holds common list filters (search + sort). Reused across products, customers, etc.
	FilterOptions struct {
		SearchQuery *string
		Sort        *string // value: column,order. E.g. created_at,desc
	}

	/********************* User ************************/
	User struct {
		ID        int          `db:"id"`
		ShopID    int          `db:"shop_id"`
		Name      string       `db:"name"`
		Email     string       `db:"email"`
		Password  string       `db:"password"`
		Role      string       `db:"role"`
		CreatedAt time.Time    `db:"created_at"`
		UpdatedAt sql.NullTime `db:"updated_at"`
	}

	/********************* Shop ************************/
	Shop struct {
		ID         int          `db:"id"`
		Name       string       `db:"name"`
		ShareToken string       `db:"share_token"`
		CreatedAt  time.Time    `db:"created_at"`
		UpdatedAt  sql.NullTime `db:"updated_at"`
	}

	/******************* Customer *********************/
	Customer struct {
		ID        int          `db:"id"`
		ShopID    int          `db:"shop_id"`
		Name      string       `db:"name"`
		Phone     string       `db:"phone"`
		Address   string       `db:"address"`
		CreatedAt time.Time    `db:"created_at"`
		UpdatedAt sql.NullTime `db:"updated_at"`
	}

	/******************* Product *********************/
	Product struct {
		ID            int          `db:"id"`
		ShopID        int          `db:"shop_id"`
		Name          string       `db:"name"`
		Description   string       `db:"description"`
		Price         int          `db:"price"`
		OriginalPrice int          `db:"original_price"`
		ImageURL      string       `db:"image_url"`
		CreatedAt     time.Time    `db:"created_at"`
		UpdatedAt     sql.NullTime `db:"updated_at"`
	}

	PurchaseProduct struct {
		ProductName string `db:"name"`
		Price       int    `db:"price"`
		Qty         int    `db:"qty"`
	}

	/******************** Order **********************/
	// OrderFilterOptions holds optional filters for listing orders.
	// Used by handler and service; store consumes it. Add fields as needed (e.g. Status, CustomerID).
	OrderFilterOptions struct {
		SearchQuery *string
		DateFrom    *time.Time
		DateTo      *time.Time
		Status      []string
		Sort        *string // value: column,order. E.g. created_at,desc
	}

	Order struct {
		ID           int          `db:"id"`
		ShopID       int          `db:"shop_id"`
		CustomerName string       `db:"customer_name"`
		TotalPrice   int          `db:"total_price"`
		Status       string       `db:"status"`
		Notes        string       `db:"notes"`
		CreatedAt    time.Time    `db:"created_at"`
		UpdatedAt    sql.NullTime `db:"updated_at"`
	}

	OrderItem struct {
		ID          int          `db:"id"`
		OrderID     int          `db:"order_id"`
		ProductName string       `db:"product_name"`
		Price       int          `db:"price"`
		Qty         int          `db:"qty"`
		CreatedAt   time.Time    `db:"created_at"`
		UpdatedAt   sql.NullTime `db:"updated_at"`
	}

	TempOrder struct {
		ID            int          `db:"id"`
		ShopID        int          `db:"shop_id"`
		CustomerName  string       `db:"customer_name"`
		CustomerPhone string       `db:"customer_phone"`
		TotalPrice    int          `db:"total_price"`
		Status        string       `db:"status"`
		CreatedAt     time.Time    `db:"created_at"`
		UpdatedAt     sql.NullTime `db:"updated_at"`
	}

	TempOrderItem struct {
		ID          int          `db:"id"`
		TempOrderID int          `db:"temp_order_id"`
		ProductID   int          `db:"product_id"`
		ProductName string       `db:"product_name"`
		Price       int          `db:"price"`
		Qty         int          `db:"qty"`
		CreatedAt   time.Time    `db:"created_at"`
	}

	/******************* Subscription *******************/
	Plan struct {
		ID            int          `db:"id"`
		Name          string       `db:"name"`
		DisplayName   string       `db:"display_name"`
		DescriptionEN string       `db:"description_en"`
		DescriptionID string       `db:"description_id"`
		PriceIDR      int          `db:"price_idr"`
		MaxUsers      int          `db:"max_users"`
		IsActive      bool         `db:"is_active"`
		CreatedAt     time.Time    `db:"created_at"`
		UpdatedAt     sql.NullTime `db:"updated_at"`
	}

	Subscription struct {
		ID                 int          `db:"id"`
		ShopID             int          `db:"shop_id"`
		PlanID             int          `db:"plan_id"`
		Status             string       `db:"status"`
		TrialEndsAt        sql.NullTime `db:"trial_ends_at"`
		CurrentPeriodStart time.Time    `db:"current_period_start"`
		CurrentPeriodEnd   time.Time    `db:"current_period_end"`
		CancelledAt        sql.NullTime `db:"cancelled_at"`
		CreatedAt          time.Time    `db:"created_at"`
		UpdatedAt          sql.NullTime `db:"updated_at"`
	}

	Payment struct {
		ID               int          `db:"id"`
		ShopID           int          `db:"shop_id"`
		SubscriptionID   int          `db:"subscription_id"`
		PlanID           int          `db:"plan_id"`
		MidtransOrderID  string       `db:"midtrans_order_id"`
		MidtransTxnID    string       `db:"midtrans_txn_id"`
		AmountIDR        int          `db:"amount_idr"`
		Status           string       `db:"status"`
		SnapToken        string       `db:"snap_token"`
		RedirectURL      string       `db:"redirect_url"`
		PaidAt           sql.NullTime `db:"paid_at"`
		CreatedAt        time.Time    `db:"created_at"`
		UpdatedAt        sql.NullTime `db:"updated_at"`
	}
)
