package model

// RoleFilter Three fields can not use at the same time.
type RoleFilter struct {
	Names []string
	Name  string
	IDs   []int
	ID    int
}
