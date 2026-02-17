package constant

// User role constants
const (
	RoleSystem = "system"
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
)

// Order status constants
const (
	OrderStatusCreated    = "created"
	OrderStatusInProgress = "in_progress"
	OrderStatusInDelivery = "in_delivery"
	OrderStatusDone       = "done"
	OrderStatusCancelled  = "cancelled"

	TempOrderStatusPending   = "pending"
	TempOrderStatusAccepted  = "accepted"
	TempOrderStatusRejected  = "rejected"
)
