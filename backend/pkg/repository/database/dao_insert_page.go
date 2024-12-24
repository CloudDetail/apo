package database

// InsertPage saves embedded interface.
type InsertPage struct {
	PageID     int    `gorm:"column:page_id;primary_key;auto_increment" json:"pageId"`
	Url        string `gorm:"column:url" json:"url"`
	Type       string `gorm:"column:type" json:"type"` // For now it's grafana or jaeger.
	MenuItemID int    `gorm:"column:menu_item_id;uniqueIndex" json:"-"`

	MenuItemKey string `gorm:"-" json:"-"`
}

func (t *InsertPage) TableName() string {
	return "insert_page"
}

func (repo *daoRepo) GetItemInsertPage(items *[]MenuItem) error {
	if len(*items) == 0 {
		return nil
	}

	itemIDs := make([]int, 0, len(*items))
	for _, item := range *items {
		itemIDs = append(itemIDs, item.ItemID)
	}

	var pages []InsertPage
	if err := repo.db.Table("insert_page").
		Select("page_id, url, type").
		Where("menu_item_id IN ?", itemIDs).
		Find(&pages).Error; err != nil {
		return err
	}

	pageMap := make(map[int]InsertPage)
	for _, page := range pages {
		pageMap[page.MenuItemID] = page
	}

	for i := range *items {
		if page, ok := pageMap[(*items)[i].ItemID]; ok {
			(*items)[i].Page = &page
		}
	}

	return nil
}
