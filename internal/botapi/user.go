package botapi

type UserRole int

const (
	RegularUser UserRole = iota
	PromotedUser
	AdminUser
)
