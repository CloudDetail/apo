package profile

type Tenant struct {
	TenantID int `gorm:"column:tenant_id;primary_key" json:"tenantId"`

	Name string `gorm:"column:name" json:"name"`
	// TODO More detail fields
	Info string `gorm:"column:info" json:"info"`
}
