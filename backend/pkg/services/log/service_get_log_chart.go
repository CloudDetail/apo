package log

import (
	"errors"
	"sort"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

const SecondToMirco = 1000000

func (s *service) getChart(req *request.LogQueryRequest) (*response.LogChartResponse, error) {
	rows, interval, err := s.chRepo.GetLogChart(req)
	if err != nil {
		return nil, err
	}
	charts := make([]*response.LogHistogram, 0, len(rows))

	for _, row := range rows {
		chart := response.LogHistogram{}
		if row["count"] != nil {
			switch row["count"].(type) {
			case uint64:
				chart.Count = row["count"].(uint64)
			}
		}
		if row["timeline"] != nil {
			switch row["timeline"].(type) {
			case time.Time:
				chart.From = row["timeline"].(time.Time).Unix()
			case *time.Time:
				chart.From = row["timeline"].(*time.Time).Unix()
			}
		}
		charts = append(charts, &chart)
	}
	res := &response.LogChartResponse{}
	chartMap := make(map[int64]*response.LogHistogram)
	// get key info
	var firstFrom int64
	var latestFrom int64
	for i, chart := range charts {
		chartMap[chart.From] = chart
		res.Count += chart.Count
		if i == 0 {
			firstFrom = chart.From
		}
		latestFrom = chart.From
	}
	// fill charts
	st, et := req.StartTime/1000000, req.EndTime/1000000
	if (firstFrom < st-interval || firstFrom > et+interval) || (latestFrom < st-interval || latestFrom > et+interval) {
		return nil, errors.New("invalid time range")
	}
	// fill head
	if st+interval < firstFrom {
		// 说明有很多数据需要填充
		fillNum := (firstFrom - st) / interval
		for i := int64(0); i < (fillNum); i++ {
			from := firstFrom - interval*(i+1)
			if from < st {
				from = st
			}
			if _, ok := chartMap[from]; !ok {
				chartMap[from] = &response.LogHistogram{
					Count: 0,
					From:  from,
					To:    firstFrom - interval*i,
				}
			}
		}
	}
	// fill tail
	if et-interval > latestFrom {
		// 说明有很多数据需要填充
		fillNum := (et - latestFrom) / interval
		for i := int64(0); i < (fillNum); i++ {
			// to := latestFrom + interval*(i+2)
			from := latestFrom + interval*(i+1)
			// if to > st {
			// 	to = st
			// }
			if _, ok := chartMap[from]; !ok {
				chartMap[from] = &response.LogHistogram{
					Count: 0,
					From:  from,
					To:    firstFrom - interval*i,
				}
			}
		}
	}
	for i := firstFrom; i < latestFrom; i += interval {
		if _, ok := chartMap[i]; !ok {
			chartMap[i] = &response.LogHistogram{
				Count: 0,
				From:  i,
				To:    i + interval,
			}
		}
	}
	fillCharts := make([]*response.LogHistogram, 0)
	for _, chart := range chartMap {
		fillCharts = append(fillCharts, chart)
	}
	sort.Slice(fillCharts, func(i int, j int) bool {
		return fillCharts[i].From < fillCharts[j].From
	})
	l := len(fillCharts)
	if l == 1 {
		fillCharts[0].From = st * SecondToMirco
		fillCharts[0].To = et * SecondToMirco
	} else if l > 1 {
		for i := range fillCharts {
			if i == 0 {
				fillCharts[0].From = st * SecondToMirco
				fillCharts[0].To = fillCharts[1].From * SecondToMirco
			} else if i == l-1 {
				fillCharts[i].To = et * SecondToMirco
			} else {
				fillCharts[i].To = fillCharts[i+1].From * SecondToMirco
			}
		}
	}
	res.Histograms = fillCharts
	return res, nil
}

func (s *service) GetLogChart(req *request.LogQueryRequest) (*response.LogChartResponse, error) {
	return s.getChart(req)
}
