package serviceoverview

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) UrlAVG(Urls *[]Url, serviceName string, endTime time.Time, duration string) (*[]Url, error) {
	var AvgErrorRateRes []prometheus.MetricResult
	//AvgErrorRateRes, err = s.promRepo.QueryPrometheusError(searchTime)
	queryAvgError := prometheus.QueryEndPointPromql(duration, prometheus.AvgError, serviceName)
	AvgErrorRateRes, err := s.promRepo.QueryErrorRateData(endTime, queryAvgError)
	for _, result := range AvgErrorRateRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Urls)[i].AvgErrorRate = &value
				}
				break
			}
		}
		if !found {
			newUrl := Url{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.AvgErrorRate = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var AvgLatencyRes []prometheus.MetricResult
	//AvgLatencyRes, err = s.promRepo.QueryPrometheusLatency(searchTime)
	queryAvgLatency := prometheus.QueryEndPointPromql(duration, prometheus.AvgLatency, serviceName)
	AvgLatencyRes, err = s.promRepo.QueryLatencyData(endTime, queryAvgLatency)
	for _, result := range AvgLatencyRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Urls)[i].AvgLatency = &value
				}
				break
			}
		}
		if !found {
			newUrl := Url{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.AvgLatency = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var AvgTPSRes []prometheus.MetricResult
	//AvgTPSRes, err = s.promRepo.QueryPrometheusTPS(searchTime)
	queryAvgTPS := prometheus.QueryEndPointPromql(duration, prometheus.AvgTPS, serviceName)
	AvgTPSRes, err = s.promRepo.QueryData(endTime, queryAvgTPS)
	for _, result := range AvgTPSRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Urls)[i].AvgTPS = &value
				}
				break
			}
		}
		if !found {
			newUrl := Url{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.AvgTPS = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	return Urls, err
}

func (s *service) UrlDOD(Urls *[]Url, serviceName string, endTime time.Time, duration string) (*[]Url, error) {
	latencyDODquery := prometheus.QueryEndPointPromql(duration, prometheus.LatencyDOD, serviceName)
	latencyDoDres, err := s.promRepo.QueryData(endTime, latencyDODquery)
	for _, result := range latencyDoDres {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Urls)[i].LatencyDayOverDay = &value
				}
				break
			}
		}
		if !found {
			newUrl := Url{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.LatencyDayOverDay = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	//errorDoDres, err := s.promRepo.QueryPrometheusErrorDayOver(searchTime)
	errorDODquery := prometheus.QueryEndPointPromql(duration, prometheus.ErrorDOD, serviceName)
	errorDoDres, err := s.promRepo.QueryData(endTime, errorDODquery)
	// 更新wrongUrls中的内容
	for _, result := range errorDoDres {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时,错误率赋值为MaxFloat64
					(*Urls)[i].ErrorRateDayOverDay = &value
				} else {
					var value float64
					value = RES_MAX_VALUE
					pointer := &value
					(*Urls)[i].ErrorRateDayOverDay = pointer
				}
				break
			}
		}
		if !found {
			newUrl := Url{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.ErrorRateDayOverDay = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	//tpsResults, err := s.promRepo.QueryPrometheusTPSDayOver(searchTime)
	tpsDODquery := prometheus.QueryEndPointPromql(duration, prometheus.TPSDOD, serviceName)
	tpsResults, err := s.promRepo.QueryData(endTime, tpsDODquery)
	for _, result := range tpsResults {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Urls)[i].TPSDayOverDay = &value
				}
				break
			}
		}
		if !found {
			newUrl := Url{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.TPSDayOverDay = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	return Urls, err
}
func (s *service) UrlWOW(Urls *[]Url, serviceName string, endTime time.Time, duration string) (*[]Url, error) {

	var LatencyWoWRes []prometheus.MetricResult
	//LatencyWoWRes, err = s.promRepo.QueryPrometheusLatencyWeekOver(searchTime)
	latencyWOWquery := prometheus.QueryEndPointPromql(duration, prometheus.LatencyWOW, serviceName)
	LatencyWoWRes, err := s.promRepo.QueryData(endTime, latencyWOWquery)
	for _, result := range LatencyWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Urls)[i].LatencyWeekOverWeek = &value
				}
				break
			}
		}
		if !found {
			newUrl := Url{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.LatencyWeekOverWeek = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var TPSWoWRes []prometheus.MetricResult
	//TPSWoWRes, err = s.promRepo.QueryPrometheusTPSWeekOver(searchTime)
	TPSWOWquery := prometheus.QueryEndPointPromql(duration, prometheus.TPSWOW, serviceName)
	TPSWoWRes, err = s.promRepo.QueryData(endTime, TPSWOWquery)
	for _, result := range TPSWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Urls)[i].TPSWeekOverWeek = &value
				}
				break
			}
		}
		if !found {
			newUrl := Url{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.TPSWeekOverWeek = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var ErrorWoWRes []prometheus.MetricResult
	//ErrorWoWRes, err = s.promRepo.QueryPrometheusErrorWeekOver(searchTime)
	errorWoWquery := prometheus.QueryEndPointPromql(duration, prometheus.ErrorWOW, serviceName)
	ErrorWoWRes, err = s.promRepo.QueryData(endTime, errorWoWquery)
	for _, result := range ErrorWoWRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时,错误率赋值为MaxFloat64
					(*Urls)[i].ErrorRateWeekOverWeek = &value
				} else {
					var value float64
					value = RES_MAX_VALUE
					pointer := &value
					(*Urls)[i].ErrorRateWeekOverWeek = pointer
				}
				break
			}
		}
		if !found {
			newUrl := Url{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.ErrorRateWeekOverWeek = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	return Urls, err
}

// UrlRangeData 查询曲线图
func (s *service) UrlRangeData(Services *[]serviceDetail, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]serviceDetail, error) {
	var newUrls []Url
	var contentKeys []string
	var stepToStr string

	stepMinutes := float64(step) / float64(time.Minute)
	// 格式化为字符串，保留一位小数
	stepToStr = fmt.Sprintf("%.1fm", stepMinutes)

	// 遍历 services 数组，获取每个 URL 的 ContentKey 并存储到切片中
	for _, service := range *Services {
		for _, Url := range service.Urls {
			contentKeys = append(contentKeys, Url.ContentKey)
		}
	}
	//fmt.Printf("contentKeys: %d", len(contentKeys))
	var err error
	var errorDataRes []prometheus.MetricResult
	//每300个url查询一次
	batchSize := 300
	// 分批处理 contentKeys
	for i := 0; i < len(contentKeys); i += batchSize {
		end := i + batchSize
		if end > len(contentKeys) {
			end = len(contentKeys)
		}
		batch := contentKeys[i:end]
		//errorDataRes, err = s.promRepo.QueryRangePrometheusErrorLast30min(searchTime)
		errorDataQuery := prometheus.QueryEndPointRangePromql(stepToStr, duration, prometheus.ErrorData, batch)
		errorDataRes, err = s.promRepo.QueryRangeErrorData(startTime, endTime, errorDataQuery, step)
		for _, result := range errorDataRes {
			contentKey := result.Metric.ContentKey
			serviceName := result.Metric.SvcName
			found := false

			for i, Url := range newUrls {
				if Url.ContentKey == contentKey && Url.SvcName == serviceName {
					found = true
					newUrls[i].ErrorRateData = result.Values
					break
				}
			}
			if !found {
				newUrl := Url{
					ContentKey:    contentKey,
					SvcName:       serviceName,
					ErrorRateData: result.Values,
				}
				newUrls = append(newUrls, newUrl)
			}
		}
	}
	var LatencyDataRes []prometheus.MetricResult
	// 分批处理 contentKeys
	for i := 0; i < len(contentKeys); i += batchSize {
		end := i + batchSize
		if end > len(contentKeys) {
			end = len(contentKeys)
		}
		batch := contentKeys[i:end]
		//LatencyDataRes, err = s.promRepo.QueryRangePrometheusLatencyLast30min(searchTime)
		latencyDataQuery := prometheus.QueryEndPointRangePromql(stepToStr, duration, prometheus.LatencyData, batch)
		LatencyDataRes, err = s.promRepo.QueryRangeLatencyData(startTime, endTime, latencyDataQuery, step)
		for _, result := range LatencyDataRes {
			contentKey := result.Metric.ContentKey
			serviceName := result.Metric.SvcName
			found := false
			for i, Url := range newUrls {
				if Url.ContentKey == contentKey && Url.SvcName == serviceName {
					found = true
					newUrls[i].LatencyData = result.Values
					break
				}
			}
			if !found {
				newUrl := Url{
					ContentKey:  contentKey,
					SvcName:     serviceName,
					LatencyData: result.Values,
				}
				newUrls = append(newUrls, newUrl)
			}
		}
	}
	var TPSLastDataRes []prometheus.MetricResult
	// 分批处理 contentKeys
	for i := 0; i < len(contentKeys); i += batchSize {
		end := i + batchSize
		if end > len(contentKeys) {
			end = len(contentKeys)
		}
		batch := contentKeys[i:end]
		//TPSLastDataRes, err = s.promRepo.QueryRangePrometheusTPSLast30min(searchTime)
		TPSDataQuery := prometheus.QueryEndPointRangePromql(stepToStr, duration, prometheus.TPSData, batch)
		TPSLastDataRes, err = s.promRepo.QueryRangeData(startTime, endTime, TPSDataQuery, step)
		for _, result := range TPSLastDataRes {
			contentKey := result.Metric.ContentKey
			serviceName := result.Metric.SvcName
			found := false
			for i, Url := range newUrls {
				if Url.ContentKey == contentKey && Url.SvcName == serviceName {
					found = true
					newUrls[i].TPSData = result.Values
					break
				}
			}
			if !found {
				newUrl := Url{
					ContentKey: contentKey,
					SvcName:    serviceName,
					TPSData:    result.Values,
				}
				newUrls = append(newUrls, newUrl)
			}
		}

	}

	for _, url := range newUrls {
		serviceName := url.SvcName
		contentKey := url.ContentKey
		for j, _ := range *Services {
			if (*Services)[j].ServiceName == serviceName {
				for k, _ := range (*Services)[j].Urls {
					if contentKey == (*Services)[j].Urls[k].ContentKey {
						(*Services)[j].Urls[k].LatencyData = url.LatencyData
						(*Services)[j].Urls[k].ErrorRateData = url.ErrorRateData
						(*Services)[j].Urls[k].TPSData = url.TPSData
					}
				}
			}
		}
	}
	return Services, err
}

// UrlLatencySource 查询延迟主要依赖
func (s *service) UrlLatencySource(Urls *[]Url, serviceName string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]Url, error) {
	var stepToStr string
	if step >= time.Hour {
		stepToStr = strconv.FormatInt(int64(step/time.Hour), 10) + "h"
	} else if step >= time.Minute {
		stepToStr = strconv.FormatInt(int64(step/time.Minute), 10) + "m"
	} else {
		stepToStr = strconv.FormatInt(int64(step/time.Second), 10) + "s"
	}
	var LatencySourceRes []prometheus.MetricResult
	//LatencySourceRes, err = s.promRepo.QueryPrometheusLatencyWeekOver(searchTime)
	LatencySourcequery := prometheus.QueryEndPointPromql(stepToStr, prometheus.DelaySource, serviceName)
	LatencySourceRes, err := s.promRepo.QueryRangeData(startTime, endTime, LatencySourcequery, step)
	for _, result := range LatencySourceRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Urls)[i].DelaySource = &value
				}
				break
			}
		}

	}

	return Urls, err
}

// UrlAVG1min 查询最近一分钟之内的平均值
func (s *service) UrlAVG1min(Urls *[]Url, serviceName string, endTime time.Time, duration string) (*[]Url, error) {
	var Avg1minErrorRateRes []prometheus.MetricResult
	//Avg1minErrorRateRes, err = s.promRepo.QueryPrometheusError(searchTime)
	queryAvg1minError := prometheus.QueryEndPointPromql(duration, prometheus.Avg1minError, serviceName)
	Avg1minErrorRateRes, err := s.promRepo.QueryErrorRateData(endTime, queryAvg1minError)
	//log.Printf("%v", Avg1minErrorRateRes)
	for _, result := range Avg1minErrorRateRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Urls)[i].Avg1minErrorRate = &value
				}
				break
			}
		}
		if !found {
			newUrl := Url{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.AvgErrorRate = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var Avg1minLatencyRes []prometheus.MetricResult
	//Avg1minLatencyRes, err = s.promRepo.QueryPrometheusLatency(searchTime)
	queryAvg1minLatency := prometheus.QueryEndPointPromql(duration, prometheus.Avg1minLatency, serviceName)
	Avg1minLatencyRes, err = s.promRepo.QueryLatencyData(endTime, queryAvg1minLatency)
	for _, result := range Avg1minLatencyRes {
		contentKey := result.Metric.ContentKey
		serviceName := result.Metric.SvcName
		found := false
		value := result.Values[0].Value
		for i, Url := range *Urls {
			if Url.ContentKey == contentKey && Url.SvcName == serviceName {
				found = true
				if !math.IsInf(value, 0) { //为无穷大时则不赋值
					(*Urls)[i].Avg1minLatency = &value
				}
				break
			}
		}
		if !found {
			newUrl := Url{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.AvgLatency = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	return Urls, err
}
