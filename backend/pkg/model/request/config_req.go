package request

type SetTTLRequest struct {
	DataType string `json:"dataType" binding:"required,oneof=logs trace k8s other topology"` // 类型（日志、trace、Kubernetes、其他）
	Day      int    `json:"day" binding:"required"`                                          // 保存数据周期天数
}

type SetSingleTTLRequest struct {
	Name string `json:"name" binding:"required"` // 表名
	Day  int    `json:"day" binding:"required"`  // 保存数据周期天数
}
