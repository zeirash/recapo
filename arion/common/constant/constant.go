package constant

const (
	// User role constants
	RoleSystem = "system"
	RoleOwner  = "owner"
	RoleAdmin  = "admin"

	// Github label constants
	GithubLabelUser = "user"
	GithubLabelBug  = "bug"
	GithubLabelEnhancement = "enhancement"

	// Subscription status constants
	SubStatusTrialing  = "trialing"
	SubStatusActive    = "active"
	SubStatusPastDue   = "past_due"
	SubStatusCancelled = "cancelled"
	SubStatusExpired   = "expired"

	// Payment status constants
	PaymentStatusPending    = "pending"
	PaymentStatusSettlement = "settlement"
	PaymentStatusCapture    = "capture"
	PaymentStatusDeny       = "deny"
	PaymentStatusCancel     = "cancel"
	PaymentStatusExpire     = "expire"
	PaymentStatusFailure    = "failure"

	// Order status constants
	FilterStatusAll       = "all"

	OrderStatusCreated    = "created"
	OrderStatusInProgress = "in_progress"
	OrderStatusInDelivery = "in_delivery"
	OrderStatusDone       = "done"
	OrderStatusCancelled  = "cancelled"

	TempOrderStatusPending   = "pending"
	TempOrderStatusAccepted  = "accepted"
	TempOrderStatusRejected  = "rejected"

	OrderPaymentStatusOutstanding = "outstanding"
	OrderPaymentStatusPaid        = "paid"
)
