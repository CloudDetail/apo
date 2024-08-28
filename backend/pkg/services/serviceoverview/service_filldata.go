package serviceoverview

import (
	"fmt"
	"math"
	"strconv"
	"time"

	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) UrlAVG(Urls *[]Endpoint, serviceName string, endTime time.Time, duration string) (*[]Endpoint, error) {
	var AvgErrorRateRes []prom.MetricResult
	//AvgErrorRateRes, err = s.promRepo.QueryPrometheusError(searchTime)
	queryAvgError := prom.QueryEndPointPromql(duration, prom.AvgError, serviceName)
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
			newUrl := Endpoint{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.AvgErrorRate = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var AvgLatencyRes []prom.MetricResult
	//AvgLatencyRes, err = s.promRepo.QueryPrometheusLatency(searchTime)
	queryAvgLatency := prom.QueryEndPointPromql(duration, prom.AvgLatency, serviceName)
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
			newUrl := Endpoint{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.AvgLatency = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var AvgTPSRes []prom.MetricResult
	//AvgTPSRes, err = s.promRepo.QueryPrometheusTPS(searchTime)
	queryAvgTPS := prom.QueryEndPointPromql(duration, prom.AvgTPS, serviceName)
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
					(*Urls)[i].AvgTPM = &value
				}
				break
			}
		}
		if !found {
			newUrl := Endpoint{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.AvgTPM = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	return Urls, err
}

// EndpointsREDMetric 查询Endpoint级别的RED指标结果(包括平均值,日同比变化率,周同比变化率)
func (s *service) EndpointsREDMetric(startTime, endTime time.Time, filter EndpointsFilter) *EndpointsMap {
	var res = &EndpointsMap{
		Endpoints:    []*Endpoint{},
		EndpointsMap: map[string]*Endpoint{},
	}

	var filters []string
	if len(filter.ServiceName) > 0 {
		filters = append(filters, prom.ServiceRegexPQLFilter, prom.RegexContainsValue(filter.ServiceName))
	}
	if len(filter.EndpointName) > 0 {
		filters = append(filters, prom.ContentKeyRegexPQLFilter, prom.RegexContainsValue(filter.EndpointName))
	}
	if len(filter.Namespace) > 0 {
		filters = append(filters, prom.NamespacePQLFilter, filter.Namespace)
	}

	// 填充时间段内的平均RED指标
	s.fillMetric(res, AVG, startTime, endTime, filters)
	// 填充时间段内的RED指标日同比
	s.fillMetric(res, DOD, startTime, endTime, filters)
	// 填充时间段内的RED指标周同比
	s.fillMetric(res, WOW, startTime, endTime, filters)

	return res
}

func (s *service) EndpointsRealtimeREDMetric(filter EndpointsFilter, endpointsMap *EndpointsMap, startTime time.Time, endTime time.Time) {
	var filters []string
	if len(filter.ServiceName) > 0 {
		filters = append(filters, prom.ServiceRegexPQLFilter, prom.RegexContainsValue(filter.ServiceName))
	}
	if len(filter.EndpointName) > 0 {
		filters = append(filters, prom.ContentKeyRegexPQLFilter, prom.RegexContainsValue(filter.EndpointName))
	}
	if len(filter.Namespace) > 0 {
		filters = append(filters, prom.NamespacePQLFilter, filter.Namespace)
	}
	s.fillMetric(endpointsMap, REALTIME, startTime, endTime, filters)
}

func (s *service) fillMetric(res *EndpointsMap, metricGroup string, startTime, endTime time.Time, filters []string) {
	// 装饰器,默认不修改PQL语句,用于AVG或REALTIME两个metricGroup
	var decorator = func(apf prom.AggPQLWithFilters) prom.AggPQLWithFilters {
		return apf
	}

	switch metricGroup {
	case REALTIME:
		// 实时值使用当前时间往前3分钟作为时间间隔
		// 时间单位为microsecond
		startTime = endTime.Add(-3 * time.Minute)
	case DOD:
		decorator = prom.DayOnDay
	case WOW:
		decorator = prom.WeekOnWeek
	}

	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	avgLatency, err := s.promRepo.QueryAggMetricsWithFilter(
		decorator(prom.PQLAvgLatencyWithFilters),
		startTS, endTS,
		prom.EndpointGranularity,
		filters...,
	)
	if err != nil {
		// TODO 输出日志或记录错误到Endpoint中
	}
	res.MergeMetricResults(metricGroup, LATENCY, avgLatency)

	avgErrorRate, err := s.promRepo.QueryAggMetricsWithFilter(
		decorator(prom.PQLAvgErrorRateWithFilters),
		startTS, endTS,
		prom.EndpointGranularity,
		filters...,
	)
	if err != nil {
		// TODO 输出日志或记录错误到Endpoint中
	}
	res.MergeMetricResults(metricGroup, ERROR, avgErrorRate)

	if metricGroup == REALTIME {
		// 目前不计算TPS的实时值
		return
	}
	avgTPS, err := s.promRepo.QueryAggMetricsWithFilter(
		decorator(prom.PQLAvgTPSWithFilters),
		startTS, endTS,
		prom.EndpointGranularity,
		filters...,
	)
	if err != nil {
		// TODO 输出日志或记录错误到Endpoint中
	}

	res.MergeMetricResults(metricGroup, THROUGHPUT, avgTPS)
}

// EndpointsDelaySource 填充延时来源
// 基于输入的Endpoints填充, 会抛弃Endpoints中不存在的记录
func (s *service) EndpointsDelaySource(endpoints *EndpointsMap, startTime, endTime time.Time, filter EndpointsFilter) error {
	var filters []string
	if len(filter.ServiceName) > 0 {
		filters = append(filters, prom.ServiceRegexPQLFilter, prom.RegexContainsValue(filter.ServiceName))
	}
	if len(filter.EndpointName) > 0 {
		filters = append(filters, prom.ServiceRegexPQLFilter, prom.RegexContainsValue(filter.ServiceName))
	}
	if len(filter.Namespace) > 0 {
		filters = append(filters, prom.NamespacePQLFilter, filter.Namespace)
	}

	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	metricResults, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.PQLDepLatencyRadioWithFilters,
		startTS, endTS,
		prom.EndpointGranularity,
		filters...,
	)
	if err != nil {
		return err
	}

	for _, metricResult := range metricResults {
		if len(metricResult.Values) <= 0 {
			continue
		}
		key := metricResult.Metric.SvcName + "_" + metricResult.Metric.ContentKey
		// 所有合并值均只包含最新时间点的结果,直接取metricResult.Values[0]
		value := metricResult.Values[0].Value
		if endpoint, ok := endpoints.EndpointsMap[key]; ok {
			endpoint.DelaySource = &value
		}
	}

	// 因为float64默认初始值为0,即表示外部依赖延时占比为0
	// 符合预期,所以不再初始化未查询到DepLatencyRadio的Endpoint

	return nil
}

func (s *service) EndpointsNamespaceInfo(endpoints *EndpointsMap, startTime, endTime time.Time, filter EndpointsFilter) error {
	var filters []string
	if len(filter.ServiceName) > 0 {
		filters = append(filters, prom.ServiceRegexPQLFilter, prom.RegexContainsValue(filter.ServiceName))
	}
	if len(filter.EndpointName) > 0 {
		filters = append(filters, prom.ServiceRegexPQLFilter, prom.RegexContainsValue(filter.ServiceName))
	}
	if len(filter.Namespace) > 0 {
		filters = append(filters, prom.NamespacePQLFilter, filter.Namespace)
	}

	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	metricResult, err := s.promRepo.QueryAggMetricsWithFilter(
		prom.PQLAvgTPSWithFilters,
		startTS, endTS,
		prom.NSEndpointGranularity,
		filters...,
	)
	if err != nil {
		return err
	}

	for _, metric := range metricResult {
		if len(metric.Values) <= 0 {
			continue
		}
		key := metric.Metric.SvcName + "_" + metric.Metric.ContentKey
		if endpoint, ok := endpoints.EndpointsMap[key]; ok {
			if len(metric.Metric.Namespace) > 0 {
				// 因为查询粒度是 namespace,svc_name,contentKey 所以不用去重
				endpoint.NamespaceList = append(endpoint.NamespaceList, metric.Metric.Namespace)
			}
		}
	}

	return nil
}

func (s *service) UrlDOD(Urls *[]Endpoint, serviceName string, endTime time.Time, duration string) (*[]Endpoint, error) {
	latencyDODquery := prom.QueryEndPointPromql(duration, prom.LatencyDOD, serviceName)
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
			newUrl := Endpoint{
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
	errorDODquery := prom.QueryEndPointPromql(duration, prom.ErrorDOD, serviceName)
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
			newUrl := Endpoint{
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
	tpsDODquery := prom.QueryEndPointPromql(duration, prom.TPSDOD, serviceName)
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
					(*Urls)[i].TPMDayOverDay = &value
				}
				break
			}
		}
		if !found {
			newUrl := Endpoint{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.TPMDayOverDay = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	return Urls, err
}
func (s *service) UrlWOW(Urls *[]Endpoint, serviceName string, endTime time.Time, duration string) (*[]Endpoint, error) {

	var LatencyWoWRes []prom.MetricResult
	//LatencyWoWRes, err = s.promRepo.QueryPrometheusLatencyWeekOver(searchTime)
	latencyWOWquery := prom.QueryEndPointPromql(duration, prom.LatencyWOW, serviceName)
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
			newUrl := Endpoint{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.LatencyWeekOverWeek = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var TPSWoWRes []prom.MetricResult
	//TPSWoWRes, err = s.promRepo.QueryPrometheusTPSWeekOver(searchTime)
	TPSWOWquery := prom.QueryEndPointPromql(duration, prom.TPSWOW, serviceName)
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
					(*Urls)[i].TPMWeekOverWeek = &value
				}
				break
			}
		}
		if !found {
			newUrl := Endpoint{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.TPMWeekOverWeek = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var ErrorWoWRes []prom.MetricResult
	//ErrorWoWRes, err = s.promRepo.QueryPrometheusErrorWeekOver(searchTime)
	errorWoWquery := prom.QueryEndPointPromql(duration, prom.ErrorWOW, serviceName)
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
			newUrl := Endpoint{
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

// EndpointRangeREDChart 查询曲线图
func (s *service) EndpointRangeREDChart(Services *[]serviceDetail, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]serviceDetail, error) {
	var newUrls []Endpoint
	var contentKeys []string
	var stepToStr string

	stepMinutes := float64(step) / float64(time.Minute)
	// 格式化为字符串，保留一位小数
	stepToStr = fmt.Sprintf("%.1fm", stepMinutes)

	// 遍历 services 数组，获取每个 URL 的 ContentKey 并存储到切片中
	for _, service := range *Services {
		for _, Url := range service.Endpoints {
			contentKeys = append(contentKeys, Url.ContentKey)
		}
	}
	//fmt.Printf("contentKeys: %d", len(contentKeys))
	var err error
	var errorDataRes []prom.MetricResult
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
		errorDataQuery := prom.QueryEndPointRangePromql(stepToStr, duration, prom.ErrorData, batch)
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
				newUrl := Endpoint{
					ContentKey:    contentKey,
					SvcName:       serviceName,
					ErrorRateData: result.Values,
				}
				newUrls = append(newUrls, newUrl)
			}
		}
	}
	var LatencyDataRes []prom.MetricResult
	// 分批处理 contentKeys
	for i := 0; i < len(contentKeys); i += batchSize {
		end := i + batchSize
		if end > len(contentKeys) {
			end = len(contentKeys)
		}
		batch := contentKeys[i:end]
		//LatencyDataRes, err = s.promRepo.QueryRangePrometheusLatencyLast30min(searchTime)
		latencyDataQuery := prom.QueryEndPointRangePromql(stepToStr, duration, prom.LatencyData, batch)
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
				newUrl := Endpoint{
					ContentKey:  contentKey,
					SvcName:     serviceName,
					LatencyData: result.Values,
				}
				newUrls = append(newUrls, newUrl)
			}
		}
	}
	var TPSLastDataRes []prom.MetricResult
	// 分批处理 contentKeys
	for i := 0; i < len(contentKeys); i += batchSize {
		end := i + batchSize
		if end > len(contentKeys) {
			end = len(contentKeys)
		}
		batch := contentKeys[i:end]
		//TPSLastDataRes, err = s.promRepo.QueryRangePrometheusTPSLast30min(searchTime)
		TPSDataQuery := prom.QueryEndPointRangePromql(stepToStr, duration, prom.TPSData, batch)
		TPSLastDataRes, err = s.promRepo.QueryRangeData(startTime, endTime, TPSDataQuery, step)
		for _, result := range TPSLastDataRes {
			contentKey := result.Metric.ContentKey
			serviceName := result.Metric.SvcName
			found := false
			for i, Url := range newUrls {
				if Url.ContentKey == contentKey && Url.SvcName == serviceName {
					found = true
					newUrls[i].TPMData = result.Values
					break
				}
			}
			if !found {
				newUrl := Endpoint{
					ContentKey: contentKey,
					SvcName:    serviceName,
					TPMData:    result.Values,
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
				for k, _ := range (*Services)[j].Endpoints {
					if contentKey == (*Services)[j].Endpoints[k].ContentKey {
						(*Services)[j].Endpoints[k].LatencyData = url.LatencyData
						(*Services)[j].Endpoints[k].ErrorRateData = url.ErrorRateData
						(*Services)[j].Endpoints[k].TPMData = url.TPMData
					}
				}
			}
		}
	}
	return Services, err
}

// UrlLatencySource 查询延迟主要依赖
func (s *service) UrlLatencySource(Urls *[]Endpoint, serviceName string, startTime time.Time, endTime time.Time, duration string, step time.Duration) (*[]Endpoint, error) {
	var stepToStr string
	if step >= time.Hour {
		stepToStr = strconv.FormatInt(int64(step/time.Hour), 10) + "h"
	} else if step >= time.Minute {
		stepToStr = strconv.FormatInt(int64(step/time.Minute), 10) + "m"
	} else {
		stepToStr = strconv.FormatInt(int64(step/time.Second), 10) + "s"
	}
	var LatencySourceRes []prom.MetricResult
	//LatencySourceRes, err = s.promRepo.QueryPrometheusLatencyWeekOver(searchTime)
	LatencySourcequery := prom.QueryEndPointPromql(stepToStr, prom.DelaySource, serviceName)
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
func (s *service) UrlAVG1min(Urls *[]Endpoint, serviceName string, endTime time.Time, duration string) (*[]Endpoint, error) {
	var Avg1minErrorRateRes []prom.MetricResult
	//Avg1minErrorRateRes, err = s.promRepo.QueryPrometheusError(searchTime)
	queryAvg1minError := prom.QueryEndPointPromql(duration, prom.Avg1minError, serviceName)
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
			newUrl := Endpoint{
				ContentKey: contentKey,
				SvcName:    serviceName,
			}
			if !math.IsInf(value, 0) { //为无穷大时则不赋值
				newUrl.AvgErrorRate = &value
			}
			*Urls = append(*Urls, newUrl)
		}
	}
	var Avg1minLatencyRes []prom.MetricResult
	//Avg1minLatencyRes, err = s.promRepo.QueryPrometheusLatency(searchTime)
	queryAvg1minLatency := prom.QueryEndPointPromql(duration, prom.Avg1minLatency, serviceName)
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
			newUrl := Endpoint{
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

// EndpointsMap 用于存储相同粒度的多个指标的查询结果,使用MergeMetricResults合并
type EndpointsMap struct {
	// 用于返回Endpoint列表
	Endpoints []*Endpoint
	// EndpointsMap 用于通过ContentKey和SvcName快速查询对应的Endpoint
	EndpointsMap map[string]*Endpoint
}

const (
	// metricGroup
	REALTIME = "realtime" // endpoint时刻瞬时值
	AVG      = "avg"      // start~endpoint之间的平均值
	DOD      = "dod"      // start~endpoint时段和昨日日同比
	WOW      = "wow"      // start~endpoint时段和上周周同比

	// metricName
	LATENCY    = "latency"
	ERROR      = "error"
	THROUGHPUT = "throughput"
)

// MergeMetricResults 支持合并下面几种类型的metric
// 实时值 (realtime): 延迟 (latency), 错误率 (error)
// 平均值 (avg) : 延迟 (latency), 错误率 (error), 吞吐量 (throughput)
// 日同比 (dod) : 延迟 (latency), 错误率 (error), 吞吐量 (throughput)
// 周同比 (wow) : 延迟 (latency), 错误率 (error), 吞吐量 (throughput)
func (m *EndpointsMap) MergeMetricResults(metricGroup, metricName string, metricResults []prom.MetricResult) {
	for _, metricResult := range metricResults {
		if len(metricResult.Values) <= 0 {
			continue
		}
		key := metricResult.Metric.SvcName + "_" + metricResult.Metric.ContentKey
		// 所有合并值均只包含最新时间点的结果,直接取metricResult.Values[0]
		value := metricResult.Values[0].Value
		endpoint, find := m.EndpointsMap[key]
		if !find && metricName == LATENCY {
			// 由Latency查询结果添加新的endpoint,如果latency指标无结果, 没有添加其他指标的必要
			endpoint = &Endpoint{
				ContentKey: metricResult.Metric.ContentKey,
				SvcName:    metricResult.Metric.SvcName,
			}
			m.Endpoints = append(m.Endpoints, endpoint)
			m.EndpointsMap[key] = endpoint
		} else if !find {
			continue
		}
		if math.IsInf(value, 0) {
			continue
		}
		SetValue(endpoint, metricGroup, metricName, value)
	}
}

func SetValue(e *Endpoint, metricGroup, metricName string, value float64) bool {
	switch metricGroup {
	case REALTIME:
		switch metricName {
		case LATENCY:
			micros := value / 1e3
			e.Avg1minLatency = &micros
		case ERROR:
			e.Avg1minErrorRate = &value
		}
	case AVG:
		switch metricName {
		case LATENCY:
			micros := value / 1e3
			e.AvgLatency = &micros
		case ERROR:
			e.AvgErrorRate = &value
		case THROUGHPUT:
			tpm := value * 60
			e.AvgTPM = &tpm
		}
	case DOD:
		radio := (value - 1) * 100
		switch metricName {
		case LATENCY:
			e.LatencyDayOverDay = &radio
		case ERROR:
			e.ErrorRateDayOverDay = &radio
		case THROUGHPUT:
			e.TPMDayOverDay = &radio
		}
	case WOW:
		radio := (value - 1) * 100
		switch metricName {
		case LATENCY:
			e.LatencyWeekOverWeek = &radio
		case ERROR:
			e.ErrorRateWeekOverWeek = &radio
		case THROUGHPUT:
			e.TPMWeekOverWeek = &radio
		}
	}

	return true
}
