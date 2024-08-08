package serviceoverview

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
)

type getServiceMoreUrlListRequest struct {
	StartTime   int64  `form:"startTime" binding:"required"`                 // 查询开始时间
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Step        int64  `form:"step" binding:"required"`                      // 步长
	ServiceName string `form:"serviceName" binding:"required"`               // 应用名
	SortRule    int    `form:"sortRule" binding:"required"`                  //排序逻辑
}

// GetServiceMoreUrlList 获取服务的更多url列表
// @Summary 获取服务的更多url列表
// @Description 获取服务的更多url列表
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param step query int64 true "步长"
// @Param serviceName query string true "应用名称"
// @Param sortRule query int true "排序逻辑"
// @Success 200 {object} []response.ServiceDetail
// @Failure 400 {object} code.Failure
// @Router /api/service/moreUrl [get]
func (h *handler) GetServiceMoreUrlList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(getServiceMoreUrlListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		var startTime time.Time
		var endTime time.Time
		req.StartTime = req.StartTime / 1000000 //接收的微秒级别的startTime和endTime需要先转成秒级别
		req.EndTime = req.EndTime / 1000000     //接收的微秒级别的startTime和endTime需要先转成秒级别
		startTime = time.Unix(req.StartTime, 0)
		endTime = time.Unix(req.EndTime, 0)
		step := time.Duration(req.Step * 1000)
		//step := time.Minute
		serviceName := req.ServiceName
		sortRule := serviceoverview.SortType(req.SortRule)
		var res []response.ServiceDetail
		data, err := h.serviceoverview.GetServiceMoreUrl(startTime, endTime, step, serviceName, sortRule)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceMoreUrlListError,
				code.Text(code.GetServiceMoreUrlListError)).WithError(err),
			)
			return
		}
		if data != nil {
			res = data
		} else {
			res = []response.ServiceDetail{} // 确保返回一个空数组
		}

		c.Payload(res)
	}
}
