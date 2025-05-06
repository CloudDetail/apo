package response

import "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"


type GetAlertSlienceConfigResponse struct {
	Slience *slienceconfig.AlertSlienceConfig
}

type ListAlertSlienceConfigResponse struct {
	Sliences []slienceconfig.AlertSlienceConfig
}
