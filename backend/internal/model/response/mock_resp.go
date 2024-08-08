package response

type CreateResponse struct {
	Id uint `json:"id"` // 主键ID
}

type ListResponse struct {
	List       []*ListData `json:"list"`
	Pagination *Pagination `json:"pagination"`
}

type ListData struct {
	Id   int    `json:"id"`   // ID
	Name string `json:"name"` // 用户名
}

type Pagination struct {
	Total        int64 `json:"total"`          // 总记录数
	CurrentPage  int   `json:"current_page"`   // 当前页码
	PerPageCount int   `json:"per_page_count"` // 每页条数
}

type DetailResponse struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type DeleteResponse struct {
	Id uint `json:"id"` // 主键ID
}
