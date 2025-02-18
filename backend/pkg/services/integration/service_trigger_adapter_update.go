package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func init() {
	if value, find := os.LookupEnv("ADAPTER_SERVICE_ADDRESS"); find {
		if !strings.HasPrefix(value, "http://") {
			value = "http://" + value
		}
		adapterServiceAddress = value
	}
}

const adapterUpdateAPI = "/trace/api/update"

var adapterServiceAddress = "http://apo-apm-adapter-svc:8079"

func (s *service) TriggerAdapterUpdate(req *integration.TriggerAdapterUpdateRequest) {
	traceAPI, err := s.dbRepo.GetLatestTraceAPIs(req.LastUpdateTS)
	if err != nil {
		log.Println("get latest trace api error: ", err)
	}

	if traceAPI == nil {
		return
	}

	apiData, err := json.Marshal(traceAPI)
	if err != nil {
		return
	}

	resp, err := http.Post(fmt.Sprintf("%s%s", adapterServiceAddress, adapterUpdateAPI),
		"application/json", bytes.NewBuffer(apiData))
	if err != nil {
		log.Println("trigger adapter update error: ", err)
		return
	}
	resp.Body.Close()
}
