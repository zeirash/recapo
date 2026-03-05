package constant

// User role constants
const (
	RoleSystem = "system"
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
)

// Subscription status constants
const (
	SubStatusTrialing  = "trialing"
	SubStatusActive    = "active"
	SubStatusPastDue   = "past_due"
	SubStatusCancelled = "cancelled"
	SubStatusExpired   = "expired"
)

// Payment status constants
const (
	PaymentStatusPending    = "pending"
	PaymentStatusSettlement = "settlement"
	PaymentStatusCapture    = "capture"
	PaymentStatusDeny       = "deny"
	PaymentStatusCancel     = "cancel"
	PaymentStatusExpire     = "expire"
	PaymentStatusFailure    = "failure"
)

// Order status constants
const (
	FilterStatusAll       = "all"

	OrderStatusCreated    = "created"
	OrderStatusInProgress = "in_progress"
	OrderStatusInDelivery = "in_delivery"
	OrderStatusDone       = "done"
	OrderStatusCancelled  = "cancelled"

	TempOrderStatusPending   = "pending"
	TempOrderStatusAccepted  = "accepted"
	TempOrderStatusRejected  = "rejected"
)
