package apierr

// Error key constants — used in errors.New() and for programmatic comparisons.
// Each key maps to a human-readable message in common/i18n/messages/en.json.
const (
	// Validation
	ErrEmailRequired        = "err_email_required"
	ErrPasswordRequired     = "err_password_required"
	ErrEmailInvalid         = "err_email_invalid"
	ErrNameRequired         = "err_name_required"
	ErrPhoneRequired        = "err_phone_required"
	ErrAddressRequired      = "err_address_required"
	ErrCustomerIDRequired   = "err_customer_id_required"
	ErrPriceInvalid         = "err_price_invalid"
	ErrImageTooLarge        = "err_image_too_large"
	ErrImageFieldRequired   = "err_image_field_required"
	ErrImageURLRequired     = "err_image_url_required"
	ErrOrderIDRequired      = "err_order_id_required"
	ErrOrderItemIDRequired  = "err_order_item_id_required"
	ErrProductIDRequired    = "err_product_id_required"
	ErrQtyRequired          = "err_qty_required"
	ErrTempOrderIDRequired  = "err_temp_order_id_required"
	ErrShareTokenRequired   = "err_share_token_required"
	ErrPlanIDRequired       = "err_plan_id_required"
	ErrRefreshTokenRequired = "err_refresh_token_required"
	ErrCustomerNameRequired  = "err_customer_name_required"
	ErrCustomerPhoneRequired = "err_customer_phone_required"
	ErrOrderItemsRequired   = "err_order_items_required"

	// Auth / Middleware
	ErrInvalidTokenFormat  = "err_invalid_token_format"
	ErrNotAuthorized       = "err_not_authorized"
	ErrNoSystemAccess      = "err_no_system_access"
	ErrMissingShopContext  = "err_missing_shop_context"
	ErrSubscriptionRequired = "err_subscription_required"

	// Image
	ErrUnsupportedImageType = "err_unsupported_image_type"
	ErrInvalidImageURL      = "err_invalid_image_url"

	// Business logic
	ErrUserNotFound          = "err_user_not_found"
	ErrUserAlreadyExists     = "err_user_already_exists"
	ErrUserNotExist          = "err_user_not_exist"
	ErrPasswordIncorrect     = "err_password_incorrect"
	ErrInvalidRefreshToken   = "err_invalid_refresh_token"
	ErrCustomerNotFound      = "err_customer_not_found"
	ErrCustomerPhoneExists   = "err_customer_phone_exists"
	ErrActiveOrderExists     = "err_active_order_exists"
	ErrProductNotFound       = "err_product_not_found"
	ErrProductNameExists     = "err_product_name_exists"
	ErrImageNotFound         = "err_image_not_found"
	ErrOrderNotFound         = "err_order_not_found"
	ErrOrderItemNotFound     = "err_order_item_not_found"
	ErrShopNotFound          = "err_shop_not_found"
	ErrTempOrderNotFound     = "err_temp_order_not_found"
	ErrPlanNotFound          = "err_plan_not_found"
	ErrSubscriptionNotFound  = "err_subscription_not_found"
	ErrSubscriptionNotActive = "err_subscription_not_active"
	ErrPaymentNotFound       = "err_payment_not_found"
	ErrNoActivePlans         = "err_no_active_plans"
	ErrInvalidSignature      = "err_invalid_signature"
)
