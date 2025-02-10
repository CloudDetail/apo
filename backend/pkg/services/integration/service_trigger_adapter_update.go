package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

const adapterUpdateAPI = "/trace/api/update"

var adapterServiceURL = "apo-apm-adapter-svc:8079"

func (s *service) TriggerAdapterUpdate(req *integration.TriggerAdapterUpdateRequest) {
	traceAPI, err := s.dbRepo.GetLatestTraceAPIs(req.LastUpdateTS)
	if err != nil {
		log.Println("get latest trace api error: ", err)
	}

	if traceAPI == nil {
		return
	}

	apiData, _ := json.Marshal(traceAPI)
	if apiData == nil {
		return
	}

	resp, err := http.Post(fmt.Sprintf("http://%s%s", adapterServiceURL, adapterUpdateAPI),
		"application/json", bytes.NewBuffer(apiData))

	if err != nil {
		log.Println("trigger adapter update error: ", err)
		return
	}
	resp.Body.Close()
}
