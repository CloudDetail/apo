package database

// InsertPage saves embedded interface.
type InsertPage struct {
	PageID int    `gorm:"column:page_id;primary_key;auto_increment" json:"pageId"`
	Url    string `gorm:"column:url" json:"url"`
	Type   string `gorm:"column:type" json:"type"` // For now it's grafana or jaeger.
}

func (t *InsertPage) TableName() string {
	return "insert_page"
}

func (repo *daoRepo) GetItemInsertPage(items *[]MenuItem) error {
	if len(*items) == 0 {
		return nil
	}

	insertPageIDs := make([]int, 0, len(*items))
	for _, item := range *items {
		insertPageIDs = append(insertPageIDs, item.InsertPageID)
	}

	var pages []InsertPage
	if err := repo.db.Table("insert_page").
		Select("page_id, url, type").
		Where("page_id IN ?", insertPageIDs).
		Find(&pages).Error; err != nil {
		return err
	}

	pageMap := make(map[int]InsertPage)
	for _, page := range pages {
		pageMap[page.PageID] = page
	}

	for i := range *items {
		if page, ok := pageMap[(*items)[i].InsertPageID]; ok {
			(*items)[i].Page = &page
		}
	}

	return nil
}
