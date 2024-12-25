package database

// InsertPage saves embedded interface.
type InsertPage struct {
	PageID int    `gorm:"column:page_id;primary_key;auto_increment" json:"pageId"`
	Url    string `gorm:"column:url" json:"url"`
	Type   string `gorm:"column:type" json:"type"` // For now it's grafana or jaeger.
	Custom bool   `gorm:"column:custom" json:"-"`
}

// RouterInsertPage maps router to inserted page.
type RouterInsertPage struct {
	ID       int `gorm:"column:id"`
	RouterID int `gorm:"column:router_id;uniqueIndex"`
	PageID   int `gorm:"column:page_id;"`
}

func (RouterInsertPage) TableName() string {
	return "router_insert_page"
}

func (t *InsertPage) TableName() string {
	return "insert_page"
}

func (repo *daoRepo) GetRouterInsertedPage(routers []*Router) error {
	if len(routers) == 0 {
		return nil
	}

	for i := range routers {
		var page InsertPage
		err := repo.db.Table("router_insert_page").
			Select("insert_page.*").
			Joins("JOIN insert_page ON router_insert_page.page_id = insert_page.page_id").
			Where("router_insert_page.router_id = ?", routers[i].RouterID).
			Find(&page).Error
		if err != nil {
			return err
		}
		if page.PageID != 0 {
			routers[i].Page = &page
		}
	}

	return nil
}
