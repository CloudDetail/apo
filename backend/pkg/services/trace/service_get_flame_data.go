package trace

import (
	"encoding/json"
	"fmt"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"log"
)

type FlameGraph struct {
	Version     int         `json:"version"`
	Flamebearer Flamebearer `json:"flamebearer"`
	Metadata    struct {
		Name       string `json:"name"`
		Format     string `json:"format"`
		SampleRate uint32 `json:"sampleRate"`
		SpyName    string `json:"spyName"`
		Units      string `json:"units"`
	} `json:"metadata"`
}

type Flamebearer struct {
	Names    []string `json:"names"`
	Levels   [][]int  `json:"levels"`
	NumTicks int      `json:"numTicks"`
	MaxSelf  int      `json:"maxSelf"`
}

func (s *service) GetFlameGraphData(req *request.GetFlameDataRequest) (*response.GetFlameDataResponse, error) {
	flameData, err := s.chRepo.GetFlameGraphData(req.StartTime, req.EndTime, req.PID, req.TID, req.SampleType)
	if err != nil {
		return nil, err
	}
	for i := range *flameData {
		var flamebearer Flamebearer
		err = json.Unmarshal([]byte((*flameData)[i].FlameBearer), &flamebearer)
		if err != nil {
			log.Printf("Get flame graph data unmarshal err: %s", err)
			continue
		}
		var flameGraph FlameGraph
		flameGraph.Flamebearer = flamebearer
		flameGraph.Metadata.SampleRate = (*flameData)[i].SampleRate
		flameGraph.Version = 1
		flameGraph.Metadata.Format = "single"
		flameGraph.Metadata.Units = "samples"
		name := fmt.Sprintf("%d-%d", (*flameData)[i].PID, (*flameData)[i].TID)
		flameGraph.Metadata.Name = name
		bearer, err := json.Marshal(flameGraph)
		if err != nil {
			log.Printf("Get flame graph data marshal err: %s", err)
			continue
		}
		(*flameData)[i].FlameBearer = string(bearer)
	}

	return (*response.GetFlameDataResponse)(flameData), nil
}
