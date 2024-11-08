package jaeger

import (
	"fmt"
	"github.com/CloudDetail/apo/backend/config"
	"io"
	"net/http"
	"net/http/httputil"
)

const trace_path = "/jaeger/api/traces/"

func (j *jaegerRepo) GetSingleTrace(traceId string) (string, error) {
	jaegerConf := config.Get().Jaeger
	url := jaegerConf.Address + trace_path + traceId

	req, err := http.NewRequest("GET", url, nil)
	res, _ := httputil.DumpRequest(req, true)
	fmt.Println(res)
	resp, err := j.cli.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	info, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(info), nil
}
