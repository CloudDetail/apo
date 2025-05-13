# 开发规范
> 项目由以下几个模块组成，通过Controller/Service/Reporitory的模式进行开发
```
├── api -- Controller层，声明Handler，并实现请求的解析和响应 [1]
├── model
│   ├── request -- HTTP请求实体   [2]
│   └── response -- HTTP返回实体  [2]
├── services Service层，定义相关接口和实现 [3]
├── repository -- 相关操作 和 中间结果声明 [3]
│   ├── clickhouse
│   ├── database
│   └── promethues
├── code -- ErrorCode声明 [4]
└── router
    └── router_api.go -- 定义接口 与 api Handler的绑定 [5]
```

## 命名规范
### Controller
路径 - `api/<module>/`
接口声明 - 在`api/<module>/`下创建handler.go，将所有相关接口都定义在Handler接口中
接口实现 - 由`handlergen`动态生成func_xxx.go后，实现相关接口
```golang
/api/mock
├── func_create.go
├── func_delete.go
├── func_detail.go
├── func_list.go    // func_<action>.go
└── handler.go

type Handler interface {
	// Create 创建/编辑xx
	// @Tags API.mock
	// @Router /api/mock [post]
	Create() core.HandlerFunc

	// List xx列表
	// @Tags API.mock
	// @Router /api/mock [get]
	List() core.HandlerFunc

	// Detail xx详情
	// @Tags API.mock
	// @Router /api/mock/{id} [get]
	Detail() core.HandlerFunc

	// Delete 删除xx
	// @Tags API.mock
	// @Router /api/mock/{id} [delete]
	Delete() core.HandlerFunc
}
```

### Vo
路径 - `model/request` 和 `model/response`
生成方式 - 由`handlergen`动态生成后，迁移到该包

### Service
路径 `services/<module>/`
接口声明 - 在`services/<module>/`下创建service.go，并将所有接口声明在Service中
接口实现 - 在`service_<action>.go`中实现相关业务接口
```go
/services/mock
├── service_create.go
├── service_delete.go
├── service_detail.go
├── service_pagelist.go // service_<action>.go
└── service.go

type Service interface {
	Create(req *request.CreateRequest) (resp *response.CreateResponse, err error)
	PageList(req *request.ListRequest) (resp *response.ListResponse, err error)
	Detail(req *request.DetailRequest) (info *response.DetailResponse, err error)
	Delete(req *request.DeleteRequest) error
}
```

### Dao
路径 - `repository/<db>`
接口声明 - 在`repository/<db>/`下的dao.go，并将所有接口声明在Repo中
接口实现 - 在`dao_<table>.go`中实现DB CRUD接口
```golang
repository/database
├── dao.go
└── dao_mock.go // dao_<table>.go

type Repo interface {
	CreateMock(model *Mock) (id uint, err error)
	GetThreshold(id uint) (model *Mock, err error)
	ListMocksByCondition(req *request.ListRequest) (r []*response.ListData, count int64, err error)
	UpdateMockById(id uint, m map[string]interface{}) error
	DeleteMockById(id uint) error
}
```

## 开发流程
* api包声明Handler，即Controller
* model/request和 model/response 定义入参、出参信息
* services包声明并实现Service，如果有repository操作，则在repoitory包下将相关业务行为放到对应组件下，类似于Dao操作
* code包 定义ErrorCode和对应的中英文描述信息
* router_api.go 进行接口绑定

## 1 Controller开发
### 1.1 定义Handler
> 在pkg/api/下创建所需实体的handler.go，具体Demo见intrnal/pkg/api
> eg. 先创建mock文件夹
> 在mock文件夹下创建 handler.go
> 如果已有相关的handler.go，则在Handler种新增新的API接口

此处主要负责定义接口名，即Method
注释部分用于Swagger模板生成使用，将接口信息、描述、分组进行定义
```go
package mock

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"go.uber.org/zap"
)

type Handler interface {
	// Method 方法功能说明
	// @Tags API.mock
	// @Router /api/xxx [get]
	Method() core.HandlerFunc
}

type handler struct {
	logger *zap.Logger
}
```

### 1.2 模板生成Handler实现
```bash
# 构建二进制
Windows - go build -o handlergen.exe ./cmd/handlergen
Linux   - go build -o handlergen ./cmd/handlergen

# 指定需要新生成的handler包
Windows - .\handlergen.exe -handler=mock
Linux   - ./handlergen -handler=mock

# 最终会生成func_xx.go，包含methodRequest、methodResponse实体和空白实现接口。
└── file :  ./pkg/api/mock/func_method.go
```

```go
type methodRequest struct {
    // 补充请求参数
}

type methodResponse struct {
    // 补充返回参数
}

// List xx列表
// @Summary xx列表
// @Description xx列表
// @Tags API.mock
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body methodRequest true "请求信息"
// @Success 200 {object} listResponse
// @Failure 400 {object} code.Failure
// @Router /api/xxx [get]
func (h *handler) Method() core.HandlerFunc {
    return func(c core.Context) {
    }
}
```

### 1.3 参数绑定和校验
> 由于请求参数有5种类型 - Query、PathUrl、PostForm、JSON、Body，下面分别对各种类型进行定义说明

#### 1.3.1 Query
> GET /api/mock?name=xxx&id=xxx
```go
type methodRequest struct {
    Name string `form:"name" binding:"required"` // 名称
    ID   int    `form:"id" binding:"required"`   // ID
}

// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param name query string true "名称"
// @Param id query int true "ID"
// @Router /api/mock [get]
func (h *handler) Method() core.HandlerFunc {
    return func(c core.Context) {
        req := new(methodRequest)
        if err := c.ShouldBindQuery(req); err != nil {
        }
    }
}
```

#### 1.3.2 PathUrl
> GET /api/{name}/{id}
```go
type methodRequest struct {
    Name string `uri:"name" binding:"required"` // 名称
    ID   int    `uri:"id" binding:"required"`   // ID
}

// @Accept json
// @Produce json
// @Param name path string true "名称"
// @Param id path int true "ID"
// @Router /api/{name}/{id} [get]
func (h *handler) Method() core.HandlerFunc {
    return func(c core.Context) {
        req := new(methodRequest)
        if err := c.ShouldBindURI(req); err != nil {
        }
    }
}
```

#### 1.3.3 PostForm
> POST /api/mock
```go
type methodRequest struct {
    Name string `form:"name" binding:"required"` // 名称
    ID   int    `form:"id" binding:"required"`   // ID
}

// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param name formData string true "名称"
// @Param id formData int true "ID"
// @Router /api/mock [post]
func (h *handler) Method() core.HandlerFunc {
    return func(c core.Context) {
        req := new(methodRequest)
        if err := c.ShouldBindPostForm(req); err != nil {
        }
    }
}
```

#### 1.3.4 json
> POST /api/mock

```go
type methodRequest struct {
    Name string `json:"name" binding:"required"` // 名称
    ID   int    `json:"id" binding:"required"`   // ID
}

// @Accept json
// @Produce json
// @Param Request body methodRequest true "请求信息"
// @Router /api/mock [post]
func (h *handler) Method() core.HandlerFunc {
    return func(c core.Context) {
        req := new(methodRequest)
        if err := c.ShouldBindJSON(req); err != nil {
        }
    }
}
```

### 1.4 迁移request和response实体
将对应的Request 和 Response对象迁移到 model/request和 model/response中
Controller只负责请求解析 和 Service结果返回，参数需与Service层共用

### 1.5 错误处理
Handler解析完请求后，进行必填等校验，如果失败则直接返回 400

后续业务通过直接调用 Service，如果Service返回错误，则需返回该业务相关的ErrorCode

## 2 Service 声明和实现
在services包下 将相同的接口名复制到Service下，参数为对应的Request，返回则是对应的Response和error

考虑到会有DB、ClickHouse等操作，可以将对应的业务迁移到repository包下，将业务分层处理

## 3 声明code
> 在code.go文件中声明业务相关的Code
> 分为2类ErrorCode，通用型(AXXXX) 和 业务模块型(BXXYY)
> 通用型为系统级错误，eg. 参数缺失、DB连接失败等
> 业务模块型为每个接口一个错误编号，基于模块ID + 自增业务ID
```go
UserCreateError = "B0201"
UserUpdateError = "B0202"
UserDeleteError = "B0203"
...
```

再在en.go和zh-cn.go声明错误说明
```
# en.go
UserCreateError: "Failed to create user",
UserUpdateError: "Failed to update user",
UserDeleteError: "Failed to delete user",

# zh-cn.go
UserCreateError: "创建用户失败",
UserUpdateError: "更新用户失败",
UserDeleteError: "删除用户失败",
```

## 4 返回错误code
> 返回业务报错
```go
resp, err := h.userService.Create(c, req)
if err != nil {
    c.AbortWithError(
        http.StatusBadRequest,
        code.UserCreateError,
        code.Text(code.UserCreateError)).WithError(err),
    )
    return
}
```

## 5 绑定对外接口
> router_api.go中，绑定接口与实现
```go
api := r.mux.Group("/api")
{
    mockHandler := mock.New(r.logger)
    api.POST("/mock", mockHandler.Create())
    api.GET("/mock", mockHandler.List())
    api.GET("/mock/:id", mockHandler.Detail())
    api.DELETE("/mock/:id", mockHandler.Delete())
}
```

# 启动服务
## 构建swagger Doc
每次接口声明有变更 或 接口有新增都需进行Swagger Doc构建
```
go install github.com/swaggo/swag/cmd/swag@latest
swag init
```

## 构建二进制
> go build -v -o apo

## 访问Swagger，并测试接口
> ./apo

访问 http://localhost:19999/swagger/index.html