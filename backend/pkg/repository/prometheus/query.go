package prometheus

import (
	"fmt"
	"strings"
)

type QueryType int

// atodo 将正则表达式转换
const (
	AvgError QueryType = iota //平均错误率
	ErrorDOD
	ErrorWOW
	ErrorData
	AvgLatency
	LatencyDOD
	LatencyWOW
	LatencyData
	AvgTPS
	TPSDOD
	TPSWOW
	TPSData
	DelaySource
	AvgLog
	LogDOD
	LogWOW
	ServiceAvgLog
	ServiceInstancePod
	ServiceInstanceContainer
	ServiceInstancePid
	AvgDependencyLatency // 平均外部依赖耗时
	Avg1minError
	Avg1minLatency
)

/*
*
查询语句
*/

const (
	FillNodeName = `avg(
  increase(
    kindling_span_trace_duration_nanoseconds_count{
      svc_name=~"%s",
      content_key=~"%s",
    }
  )

) by(content_key,svc_name,pod,node_name,pid,container_id) 
`
)

func QueryNodeName(serviceName string, contentKey string) string {
	contentKey = EscapeRegexp(contentKey)
	return fmt.Sprintf(FillNodeName, serviceName, contentKey)
}

const (
	AVG_1MIN_ERROR_BY_SERVICE = `
		sum by(svc_name, content_key) 
(
			increase(kindling_span_trace_duration_nanoseconds_count{is_error="true", svc_name=~"%s"}[3m]
		)) / sum by(svc_name, content_key) (
			increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s"}[3m]
		))
	`
	AVG_1MIN_ERROR = `
		sum by(svc_name, content_key) (
			increase(kindling_span_trace_duration_nanoseconds_count{is_error="true"}[3m]
		)) / sum by(svc_name, content_key) (
			increase(kindling_span_trace_duration_nanoseconds_count[3m]
		))
	`
	AVG_1MIN_LATENCY_BY_SERVICE = `
sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*",svc_name=~"%s"}[3m])
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[3m]
  )
) by(content_key,svc_name)`
	AVG_1MIN_LATENCY = `sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*"}[3m])
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[3m]
  )
) by(content_key,svc_name)`
	AVG_ERROR_BY_PID = `
		sum by(svc_name, content_key, node_name, pid,node_ip) (
			increase(kindling_span_trace_duration_nanoseconds_count{is_error="true", svc_name=~"%s", content_key=~"%s", pod="", container_id=""}[%s]
		)) / sum by(svc_name, content_key, node_name, pid,node_ip) (
			increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s", content_key=~"%s", pod="", container_id=""}[%s]
		))
	`
	AVG_ERROR_BY_CONTAINERID = `
		sum by(svc_name, content_key, node_name, container_id,pid, node_ip) (
			increase(kindling_span_trace_duration_nanoseconds_count{is_error="true", svc_name=~"%s", content_key=~"%s", pod="", container_id=~".+"}[%s]
		)) / sum by(svc_name, content_key, node_name, container_id,pid, node_ip) (
			increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s", content_key=~"%s", pod="", container_id=~".+"}[%s]
		))
	`
	AVG_ERROR_BY_POD = `
		sum by(svc_name, content_key, pod, container_id, node_name, namespace,pid, node_ip) (
			increase(kindling_span_trace_duration_nanoseconds_count{is_error="true", svc_name=~"%s", content_key=~"%s", pod=~".+"}[%s]
		)) / sum by(svc_name, content_key, pod, container_id, node_name, namespace,pid, node_ip) (
			increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s", content_key=~"%s", pod=~".+"}[%s]
		))
	`

	AVG_ERROR_BY_SERVICE = `
		sum by(svc_name, content_key) 
(
			increase(kindling_span_trace_duration_nanoseconds_count{is_error="true", svc_name=~"%s"}[%s]
		)) / sum by(svc_name, content_key) (
			increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s"}[%s]
		))
	`

	AVG_ERROR = `
		sum by(svc_name, content_key) (
			increase(kindling_span_trace_duration_nanoseconds_count{is_error="true"}[%s]
		)) / sum by(svc_name, content_key) (
			increase(kindling_span_trace_duration_nanoseconds_count[%s]
		))
	`
	ERROR_DOD_BY_PID = `
((
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~""}[%s]
    )
  ) by(content_key, svc_name,pid,node_name,node_ip)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~""}[%s])
  ) by(content_key, svc_name,pid,node_name,node_ip)
)
  /
(
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~""}[%s] offset 24h
    )
  ) by(content_key, svc_name,pid,node_name,node_ip)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~""}[%s] offset 24h)
  ) by(content_key, svc_name,pid,node_name,node_ip)
)-1)*100`

	ERROR_DOD_BY_CONTAINERID = `
((
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~".+"}[%s]
    )
  ) by(content_key, svc_name,container_id,node_name)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~".+"}[%s])
  ) by(content_key, svc_name,container_id,node_name)
)
  /
(
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~".+"}[%s] offset 24h
    )
  ) by(content_key, svc_name,container_id,node_name)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~".+"}[%s] offset 24h)
  ) by(content_key, svc_name,container_id,node_name)
)-1)*100`
	ERROR_DOD_BY_POD = `
((
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~".+"}[%s]
    )
  ) by(content_key, svc_name,pod,container_id, node_name, namespace, node_ip)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~".+"}[%s])
  ) by(content_key, svc_name,pod,container_id, node_name, namespace, node_ip)
)
  /
(
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~".+"}[%s] offset 24h
    )
  ) by(content_key, svc_name,pod,container_id, node_name, namespace, node_ip)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~".+"}[%s] offset 24h)
  ) by(content_key, svc_name,pod,container_id, node_name, namespace, node_ip)
)-1)*100`
	ERROR_DOD_BY_SERVICE = ` 

((
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s"}[%s]
    )
  ) by(svc_name,content_key)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s"}[%s])
  ) by(svc_name,content_key)
)
  /
(
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s"}[%s] offset 24h
    )
  ) by(svc_name,content_key)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s"}[%s] offset 24h)
  ) by(svc_name,content_key)
)-1)*100`

	ERROR_DOD = `
((
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true"}[%s]
    )
  ) by(svc_name,content_key)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count[%s])
  ) by(svc_name,content_key)
)
  /
(
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true"}[%s] offset 24h
    )
  ) by(svc_name,content_key)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count[%s] offset 24h)
  ) by(svc_name,content_key)
)-1)*100 `

	ERROR_WOW_BY_PID = `((
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~""}[%s]
    )
  ) by(content_key, svc_name,pid,node_name,node_ip)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~""}[%s])
  ) by(content_key, svc_name,pid,node_name,node_ip)
)
  /
(
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~""}[%s] offset 7d
    )
  ) by(content_key, svc_name,pid,node_name,node_ip)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~""}[%s] offset 7d)
  ) by(content_key, svc_name,pid,node_name,node_ip)
)-1)*100`
	ERROR_WOW_BY_CONTAINERID = `((
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~".+"}[%s]
    )
  ) by(content_key, svc_name,container_id,node_name,node_ip)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~".+"}[%s])
  ) by(content_key, svc_name,container_id,node_name,node_ip)
)
  /
(
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~".+"}[%s] offset 7d
    )
  ) by(content_key, svc_name,container_id,node_name,node_ip)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~"",container_id=~".+"}[%s] offset 7d)
  ) by(content_key, svc_name,container_id,node_name,node_ip)
)-1)*100`
	ERROR_WOW_BY_POD = `((
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~".+"}[%s]
    )
  ) by(content_key, svc_name,pod,container_id, node_name, namespace, node_ip)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~".+"}[%s])
  ) by(content_key, svc_name,pod,container_id, node_name, namespace, node_ip)
)
  /
(
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s",
				content_key=~"%s",pod=~".+"}[%s] offset 7d
    )
  ) by(content_key, svc_name,pod,container_id, node_name, namespace, node_ip)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",
				content_key=~"%s",pod=~".+"}[%s] offset 7d)
  ) by(content_key, svc_name,pod,container_id, node_name, namespace, node_ip)
)-1)*100`
	ERROR_WOW_BY_SERVICE = `

((
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s"}[%s]
    )
  ) by(svc_name,content_key)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s"}[%s])
  ) by(svc_name,content_key)
)
  /
(
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true",svc_name=~"%s"}[%s] offset 7d
    )
  ) by(svc_name,content_key)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s"}[%s] offset 7d)
  ) by(svc_name,content_key)
)-1)*100`
	ERROR_WOW = `
((
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true"}[%s]
    )
  ) by(svc_name,content_key)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count[%s])
  ) by(svc_name,content_key)
)
  /
(
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{is_error="true"}[%s] offset 7d
    )
  ) by(svc_name,content_key)
    /
  sum(
    increase(kindling_span_trace_duration_nanoseconds_count[%s] offset 7d)
  ) by(svc_name,content_key)
)-1)*100  `
	AVG_LATENCY_BY_PID = `
sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s])
) by(content_key, svc_name,pid,node_name,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s]
  )
) by(content_key, svc_name,pid,node_name,node_ip)`
	AVG_LATENCY_BY_CONTAINERID = `
sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s])
) by(content_key, svc_name,container_id,node_name,pid,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s]
  )
) by(content_key, svc_name,container_id,node_name,pid,node_ip)`
	AVG_LATENCY_BY_POD = `
sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s])
) by(content_key, svc_name,pod,container_id, namespace,pid,node_name,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s]
  )
) by(content_key, svc_name,pod,container_id, namespace,pid,node_name,node_ip)`
	AVG_LATENCY_BY_SERVICE = `
sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*",svc_name=~"%s"}[%s])
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s]
  )
) by(content_key,svc_name)`

	AVG_LATENCY = `sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*"}[%s])
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s]
  )
) by(content_key,svc_name)`

	LATENCY_DOD_BY_PID = `

((sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s])
) by(content_key, svc_name,pid,node_name,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s]
  )
) by(content_key, svc_name,pid,node_name,node_ip))/(sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s]offset 24h)
) by(content_key, svc_name,pid,node_name,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s] offset 24h
  )
) by(content_key, svc_name,pid,node_name,node_ip))-1)*100`
	LATENCY_DOD_BY_CONTAINERID = `
((sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s])
) by(content_key, svc_name,container_id,node_name,pid,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s]
  )
) by(content_key, svc_name,container_id,node_name,pid,node_ip))/(sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s]offset 24h)
) by(content_key, svc_name,container_id,node_name,pid,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s] offset 24h
  )
) by(content_key, svc_name,container_id,node_name,pid,node_ip))-1)*100`
	LATENCY_DOD_BY_POD = `

((sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s])
) by(content_key, svc_name,pod,container_id, namespace,pid, node_name,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s]
  )
) by(content_key, svc_name,pod,container_id, namespace,pid, node_name,node_ip))/(sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s]offset 24h)
) by(content_key, svc_name,pod,container_id, namespace,pid, node_name,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s] offset 24h
  )
) by(content_key, svc_name,pod,container_id, namespace,pid, node_name,node_ip))-1)*100`
	LATENCY_DOD_BY_SERVICE = `
((sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*",svc_name=~"%s"}[%s])
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s]
  )
) by(content_key,svc_name))/(sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*",svc_name=~"%s"}[%s]offset 24h)
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s] offset 24h
  )
) by(content_key,svc_name))-1)*100`
	LATENCY_DOD = `

((sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*"}[%s])
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s]
  )
) by(content_key,svc_name))/(sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*"}[%s]offset 24h)
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s] offset 24h
  )
) by(content_key,svc_name))-1)*100`

	LATENCY_WOW_BY_PID = `((sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s])
) by(content_key, svc_name,pid,node_name,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s]
  )
) by(content_key, svc_name,pid,node_name,node_ip))/(sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s]offset 7d)
) by(content_key, svc_name,pid,node_name,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s] offset 7d
  )
) by(content_key, svc_name,pid,node_name,node_ip))-1)*100`
	LATENCY_WOW_BY_CONTAINERID = `
((sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s])
) by(content_key, svc_name,container_id,node_name,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s]
  )
) by(content_key, svc_name,container_id,node_name,node_ip))/(sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s]offset 7d)
) by(content_key, svc_name,container_id,node_name,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s] offset 7d
  )
) by(content_key, svc_name,container_id,node_name,node_ip))-1)*100`
	LATENCY_WOW_BY_POD = `((sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s])
) by(content_key, svc_name,pod,container_id, namespace,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s]
  )
) by(content_key, svc_name,pod,container_id, namespace,node_ip))/(sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s]offset 7d)
) by(content_key, svc_name,pod,container_id, namespace,node_ip)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s] offset 7d
  )
) by(content_key, svc_name,pod,container_id, namespace,node_ip))-1)*100`
	LATENCY_WOW_BY_SERVICE = `((sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*",svc_name=~"%s"}[%s])
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s]
  )
) by(content_key,svc_name))/(sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*",svc_name=~"%s"}[%s]offset 7d)
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s] offset 7d
  )
) by(content_key,svc_name))-1)*100`

	LATENCY_WOW = `((sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*"}[%s])
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s]
  )
) by(content_key,svc_name))/(sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*"}[%s]offset 7d)
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s] offset 7d
  )
) by(content_key,svc_name))-1)*100`
	AVG_TPS_BY_PID         = `(sum by (content_key, svc_name,pid,node_name,node_ip) (increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s])))/%s`
	AVG_TPS_BY_CONTAINERID = `(sum by (content_key, svc_name,container_id,node_name,pid, node_ip) (increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s])))/%s`
	AVG_TPS_BY_POD         = `(sum by (content_key, svc_name,pod,container_id, node_name, namespace,pid, node_ip) (increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s])))/%s`
	AVG_TPS_BY_SERVICE     = `(sum by (content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s])))/%s`
	AVG_TPS                = `(sum by (content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s])))/%s`
	TPS_DOD_BY_PID         = `
sum by (content_key, svc_name,pid,node_name,node_ip)(
  (
    (
      sum by (content_key, svc_name,pid,node_name,node_ip)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s])
      )
      -
      sum by (content_key, svc_name,pid,node_name,node_ip)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s] offset 24h)
      )
    )
    /
    sum by (content_key, svc_name,pid,node_name,node_ip)(
      increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s] offset 24h)
    )
  ) * 100
)`
	TPS_DOD_BY_CONTAINERID = `

sum by (content_key, svc_name,container_id,node_name)(
  (
    (
      sum by (content_key, svc_name,container_id,node_name)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s])
      )
      -
      sum by (content_key, svc_name,container_id,node_name)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s] offset 24h)
      )
    )
    /
    sum by (content_key, svc_name,container_id,node_name)(
      increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s] offset 24h)
    )
  ) * 100
)`
	TPS_DOD_BY_POD = `
sum by (content_key, svc_name,pod,container_id, node_name, namespace, node_ip)(
  (
    (
      sum by (content_key, svc_name,pod,container_id, node_name, namespace, node_ip)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s])
      )
      -
      sum by (content_key, svc_name,pod,container_id, node_name, namespace, node_ip)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s] offset 24h)
      )
    )
    /
    sum by (content_key, svc_name,pod,container_id, node_name, namespace, node_ip)(
      increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s] offset 24h)
    )
  ) * 100
)`
	TPS_DOD_BY_SERVICE = `
sum by (content_key, svc_name)(
  (
    (
      sum by (content_key, svc_name)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s])
      )
      -
      sum by (content_key, svc_name)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s] offset 24h)
      )
    )
    /
    sum by (content_key, svc_name)(
      increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s] offset 24h)
    )
  ) * 100
)`

	TPS_DOD = `

sum by (content_key, svc_name)(
  (
    (
      sum by (content_key, svc_name)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s])
      )
      -
      sum by (content_key, svc_name)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s] offset 24h)
      )
    )
    /
    sum by (content_key, svc_name)(
      increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s] offset 24h)
    )
  ) * 100
)
`
	TPS_WOW_BY_PID = `
sum by (content_key, svc_name,pid,node_name,node_ip)(
  (
    (
      sum by (content_key, svc_name,pid,node_name,node_ip)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s])
      )
      -
      sum by (content_key, svc_name,pid,node_name,node_ip)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s] offset 7d)
      )
    )
    /
    sum by (content_key, svc_name,pid,node_name,node_ip)(
      increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s] offset 7d)
    )
  ) * 100
)`
	TPS_WOW_BY_CONTAINERID = `
sum by (content_key, svc_name,container_id,node_name)(
  (
    (
      sum by (content_key, svc_name,container_id,node_name, node_ip)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s])
      )
      -
      sum by (content_key, svc_name,container_id,node_name, node_ip)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s] offset 7d)
      )
    )
    /
    sum by (content_key, svc_name,container_id,node_name, node_ip)(
      increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s] offset 7d)
    )
  ) * 100
)`
	TPS_WOW_BY_POD = `
sum by (content_key, svc_name,pod,container_id, node_name, namespace, node_ip)(
  (
    (
      sum by (content_key, svc_name,pod,container_id, node_name, namespace, node_ip)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s])
      )
      -
      sum by (content_key, svc_name,pod,container_id, node_name, node_ip)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s] offset 7d)
      )
    )
    /
    sum by (content_key, svc_name,pod,container_id, node_name, namespace, node_ip)(
      increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s] offset 7d)
    )
  ) * 100
)`
	TPS_WOW_BY_SERVICE = `
sum by (content_key, svc_name)(
  (
    (
      sum by (content_key, svc_name)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s])
      )
      -
      sum by (content_key, svc_name)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s] offset 7d)
      )
    )
    /
    sum by (content_key, svc_name)(
      increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s] offset 7d)
    )
  ) * 100
)`

	TPS_WOW = `
sum by (content_key, svc_name)(
  (
    (
      sum by (content_key, svc_name)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s])
      )
      -
      sum by (content_key, svc_name)(
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s] offset 7d)
      )
    )
    /
    sum by (content_key, svc_name)(
      increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s] offset 7d)
    )
  ) * 100
)`

	DELAY_SOURCE_BY_SERVICE = `


(
  (
    sum by(content_key, svc_name)(
      increase(
        kindling_profiling_epoll_duration_nanoseconds_sum{
          content_key=~".*",svc_name=~"%s"
        }[%s]
      )
    )
    /
    sum by(content_key, svc_name)(
      increase(
        kindling_profiling_epoll_duration_nanoseconds_count{
         content_key=~".*",svc_name=~"%s"
        }[%s]
      )
    )
  )
  +
  (
    sum by(content_key, svc_name)(
      increase(
        kindling_profiling_net_duration_nanoseconds_sum{
          content_key=~".*",svc_name=~"%s"
        }[%s]
      )
    )
    /
    sum by(content_key, svc_name)(
      increase(
        kindling_profiling_net_duration_nanoseconds_count{
         content_key=~".*",svc_name=~"%s"
        }[%s]
      )
    )
  )
)
/
(
  sum(
    increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*",svc_name=~"%s"}[%s])
  ) by(content_key, svc_name)
  /
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{content_key=~".*",svc_name=~"%s"}[%s]
    )
  ) by(content_key, svc_name)
)`

	DELAY_SOURCE = `(
  (
    sum by(content_key, svc_name)(
      increase(
        kindling_profiling_epoll_duration_nanoseconds_sum{
          content_key=~".*"
        }[%s]
      )
    )
    /
    sum by(content_key, svc_name)(
      increase(
        kindling_profiling_epoll_duration_nanoseconds_count{
         content_key=~".*"
        }[%s]
      )
    )
  )
  +
  (
    sum by(content_key, svc_name)(
      increase(
        kindling_profiling_net_duration_nanoseconds_sum{
          content_key=~".*"
        }[%s]
      )
    )
    /
    sum by(content_key, svc_name)(
      increase(
        kindling_profiling_net_duration_nanoseconds_count{
         content_key=~".*"
        }[%s]
      )
    )
  )
)
/
(
  sum(
    increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~".*"}[%s])
  ) by(content_key, svc_name)
  /
  sum(
    increase(
      kindling_span_trace_duration_nanoseconds_count{content_key=~".*"}[%s]
    )
  ) by(content_key, svc_name)
)`

	TPS_DATA = `
    (sum by (content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s"}[%s])))/%s
`
	LATENCY_DATA = `
sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s"}[%s])
) by(content_key,svc_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s"}[%s]
  )
) by(content_key,svc_name)`

	ERROR_DATA = `
		sum by(svc_name, content_key) (
			increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s", is_error="true"}[%s]
		)) / sum by(svc_name, content_key) (
			increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s"}[%s]
		))`
	TPS_DATA_BY_PID = `
    (sum by (content_key, svc_name,pid,node_name,node_ip) (increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s])))/%s
`
	LATENCY_DATA_BY_PID = `
sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s])
) by(content_key, svc_name,pid,node_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s]
  )
) by(content_key, svc_name,pid,node_name)`

	ERROR_DATA_BY_PID = `avg by (content_key, svc_name,pid,node_name,node_ip) (
        (increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s", is_error="true",pod=~"",container_id=~""}[%s]) or 0)
        /
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~""}[%s])
    )`
	TPS_DATA_BY_CONTAINERID = `
    (sum by (content_key, svc_name,container_id,node_name) (increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s])))/%s
`
	LATENCY_DATA_BY_CONTAINERID = `
sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s])
) by(content_key, svc_name,container_id,node_name)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s]
  )
) by(content_key, svc_name,container_id,node_name)`

	ERROR_DATA_BY_CONTAINERID = `avg by (content_key, svc_name,container_id,node_name,node_ip) (
       ( increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s", is_error="true",pod=~"",container_id=~".+"}[%s]) or 0)
        /
        increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~"",container_id=~".+"}[%s])
    )`
	TPS_DATA_BY_POD = `
    (sum by (content_key, svc_name,pod, node_name, namespace, node_ip) (increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s])))/%s
`
	LATENCY_DATA_BY_POD = `
sum(
  increase(kindling_span_trace_duration_nanoseconds_sum{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s])
) by(content_key, svc_name,pod,node_name, namespace)
  /
sum(
  increase(
    kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s]
  )
) by(content_key, svc_name,pod,node_name, namespace)`

	ERROR_DATA_BY_POD = `
(
  (
    sum(increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s", is_error="true",pod=~".+"}[%s])) 
    by (content_key, svc_name, pod, node_name, namespace, node_ip)
    or 0
  ) 
  / 
  sum(increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s])) 
  by (content_key, svc_name, pod, node_name, namespace, node_ip)
) 
or 
(
  sum(increase(kindling_span_trace_duration_nanoseconds_count{content_key=~"%s",svc_name=~"%s",pod=~".+"}[%s])) 
  by (content_key, svc_name, pod, node_name, namespace, node_ip) * 0
)`
)

func QueryEndPointPromql(duration string, queryType QueryType, serviceNames string) string {

	switch queryType {
	//突变排序的1m平均指标数据
	case Avg1minError:
		if serviceNames != "" {
			return fmt.Sprintf(AVG_1MIN_ERROR_BY_SERVICE, serviceNames, serviceNames)
		} else {
			return fmt.Sprintf(AVG_1MIN_ERROR)
		}
		//突变排序的1m平均指标数据
	case Avg1minLatency:
		if serviceNames != "" {
			return fmt.Sprintf(AVG_1MIN_LATENCY_BY_SERVICE, serviceNames, serviceNames)
		} else {
			return fmt.Sprintf(AVG_1MIN_LATENCY)
		}
	case AvgError:
		if serviceNames != "" {
			return fmt.Sprintf(AVG_ERROR_BY_SERVICE, serviceNames, duration, serviceNames, duration)
		} else {
			return fmt.Sprintf(AVG_ERROR,
				duration,
				duration)
		}
	case ErrorDOD:
		if serviceNames != "" {
			return fmt.Sprintf(ERROR_DOD_BY_SERVICE,
				serviceNames, duration, serviceNames, duration, serviceNames, duration, serviceNames, duration)
		} else {
			return fmt.Sprintf(ERROR_DOD,
				duration, duration, duration, duration)
		}
	case ErrorWOW:
		if serviceNames != "" {
			return fmt.Sprintf(ERROR_WOW_BY_SERVICE,
				serviceNames, duration, serviceNames, duration, serviceNames, duration, serviceNames, duration)
		} else {
			return fmt.Sprintf(ERROR_WOW,
				duration, duration, duration, duration)
		}

	case AvgLatency:
		if serviceNames != "" {
			return fmt.Sprintf(AVG_LATENCY_BY_SERVICE,
				serviceNames, duration, serviceNames, duration)
		} else {
			return fmt.Sprintf(AVG_LATENCY,
				duration, duration)
		}
	case LatencyDOD:
		if serviceNames != "" {
			return fmt.Sprintf(LATENCY_DOD_BY_SERVICE, serviceNames, duration, serviceNames, duration, serviceNames, duration, serviceNames, duration)
		} else {
			return fmt.Sprintf(LATENCY_DOD, duration, duration, duration, duration)
		}

	case LatencyWOW:
		if serviceNames != "" {
			return fmt.Sprintf(LATENCY_WOW_BY_SERVICE, serviceNames, duration, serviceNames, duration, serviceNames, duration, serviceNames, duration)
		} else {
			return fmt.Sprintf(LATENCY_WOW, duration, duration, duration, duration)
		}
	case AvgTPS:
		trimmedDuration := strings.TrimSuffix(duration, "m")
		if serviceNames != "" {
			return fmt.Sprintf(AVG_TPS_BY_SERVICE, serviceNames, duration, trimmedDuration)
		} else {
			return fmt.Sprintf(AVG_TPS, duration, trimmedDuration)
		}

	case TPSDOD:
		if serviceNames != "" {
			return fmt.Sprintf(TPS_DOD_BY_SERVICE, serviceNames, duration, serviceNames, duration, serviceNames, duration)
		} else {
			return fmt.Sprintf(TPS_DOD, duration, duration, duration)
		}

	case TPSWOW:
		if serviceNames != "" {
			return fmt.Sprintf(TPS_WOW_BY_SERVICE, serviceNames, duration, serviceNames, duration, serviceNames, duration)
		} else {
			return fmt.Sprintf(TPS_WOW, duration, duration, duration)
		}
	case DelaySource:
		if serviceNames != "" {
			return fmt.Sprintf(DELAY_SOURCE_BY_SERVICE, serviceNames, duration, serviceNames, duration, serviceNames, duration, serviceNames, duration, serviceNames, duration, serviceNames, duration)
		} else {
			return fmt.Sprintf(DELAY_SOURCE, duration, duration, duration, duration, duration, duration)
		}
	}

	return ""
}

func QueryEndPointRangePromql(step string, duration string, queryType QueryType, contentKeys []string) string {
	escapedKeys := make([]string, len(contentKeys))
	for i, key := range contentKeys {
		escapedKeys[i] = EscapeRegexp(key)
	}

	// 使用 strings.Join 生成正则表达式模式
	regexPattern := strings.Join(escapedKeys, "|")
	switch queryType {
	case TPSData:
		trimmedDuration := strings.TrimSuffix(step, "m")
		return fmt.Sprintf(TPS_DATA, regexPattern, step, trimmedDuration)
	case LatencyData:
		return fmt.Sprintf(LATENCY_DATA, regexPattern, step, regexPattern, step)
	case ErrorData:
		return fmt.Sprintf(ERROR_DATA, regexPattern, step, regexPattern, step)
	default:
		return ""
	}

}
func QueryPodPromql(duration string, queryType QueryType, serviceName string, contentKey string) string {
	contentKey = EscapeRegexp(contentKey)
	switch queryType {
	case AvgError:
		return fmt.Sprintf(AVG_ERROR_BY_POD, serviceName, contentKey, duration, serviceName, contentKey, duration)
	case ErrorDOD:
		return fmt.Sprintf(ERROR_DOD_BY_POD, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration)
	case ErrorWOW:
		return fmt.Sprintf(ERROR_WOW_BY_POD, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration)
	case AvgLatency:
		return fmt.Sprintf(AVG_LATENCY_BY_POD, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case LatencyDOD:
		return fmt.Sprintf(LATENCY_DOD_BY_POD, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case LatencyWOW:
		return fmt.Sprintf(LATENCY_WOW_BY_POD, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case AvgTPS:
		trimmedDuration := strings.TrimSuffix(duration, "m")
		return fmt.Sprintf(AVG_TPS_BY_POD, contentKey, serviceName, duration, trimmedDuration)
	case TPSDOD:
		return fmt.Sprintf(TPS_DOD_BY_POD, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case TPSWOW:
		return fmt.Sprintf(TPS_WOW_BY_POD, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)

	}
	return ""
}
func QueryPodRangePromql(duration string, queryType QueryType, contentKey string, serviceName string) string {
	contentKey = EscapeRegexp(contentKey)
	switch queryType {
	case TPSData:
		trimmedDuration := strings.TrimSuffix(duration, "m")
		return fmt.Sprintf(TPS_DATA_BY_POD, contentKey, serviceName, duration, trimmedDuration)
	case LatencyData:
		return fmt.Sprintf(LATENCY_DATA_BY_POD, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case ErrorData:
		return fmt.Sprintf(ERROR_DATA_BY_POD, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)
	default:
		return ""
	}

}

func QueryContainerIdPromql(duration string, queryType QueryType, serviceName string, contentKey string) string {
	contentKey = EscapeRegexp(contentKey)
	switch queryType {
	case AvgError:
		return fmt.Sprintf(AVG_ERROR_BY_CONTAINERID, serviceName, contentKey, duration, serviceName, contentKey, duration)
	case ErrorDOD:
		return fmt.Sprintf(ERROR_DOD_BY_CONTAINERID, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration)
	case ErrorWOW:
		return fmt.Sprintf(ERROR_WOW_BY_CONTAINERID, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration)
	case AvgLatency:
		return fmt.Sprintf(AVG_LATENCY_BY_CONTAINERID, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case LatencyDOD:
		return fmt.Sprintf(LATENCY_DOD_BY_CONTAINERID, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case LatencyWOW:
		return fmt.Sprintf(LATENCY_WOW_BY_CONTAINERID, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case AvgTPS:
		trimmedDuration := strings.TrimSuffix(duration, "m")
		return fmt.Sprintf(AVG_TPS_BY_CONTAINERID, contentKey, serviceName, duration, trimmedDuration)
	case TPSDOD:
		return fmt.Sprintf(TPS_DOD_BY_CONTAINERID, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case TPSWOW:
		return fmt.Sprintf(TPS_WOW_BY_CONTAINERID, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)

	}
	return ""
}
func QueryContainerIdRangePromql(duration string, queryType QueryType, contentKey string, serviceName string) string {
	contentKey = EscapeRegexp(contentKey)
	switch queryType {
	case TPSData:
		trimmedDuration := strings.TrimSuffix(duration, "m")
		return fmt.Sprintf(TPS_DATA_BY_CONTAINERID, contentKey, serviceName, duration, trimmedDuration)
	case LatencyData:
		return fmt.Sprintf(LATENCY_DATA_BY_CONTAINERID, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case ErrorData:
		return fmt.Sprintf(ERROR_DATA_BY_CONTAINERID, contentKey, serviceName, duration, contentKey, serviceName, duration)
	default:
		return ""
	}

}

func QueryPidPromql(duration string, queryType QueryType, serviceName string, contentKey string) string {
	contentKey = EscapeRegexp(contentKey)
	switch queryType {
	case AvgError:
		return fmt.Sprintf(AVG_ERROR_BY_PID, serviceName, contentKey, duration, serviceName, contentKey, duration)
	case ErrorDOD:
		return fmt.Sprintf(ERROR_DOD_BY_PID, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration)
	case ErrorWOW:
		return fmt.Sprintf(ERROR_WOW_BY_PID, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration, serviceName, contentKey, duration)
	case AvgLatency:
		return fmt.Sprintf(AVG_LATENCY_BY_PID, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case LatencyDOD:
		return fmt.Sprintf(LATENCY_DOD_BY_PID, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case LatencyWOW:
		return fmt.Sprintf(LATENCY_WOW_BY_PID, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case AvgTPS:
		trimmedDuration := strings.TrimSuffix(duration, "m")
		return fmt.Sprintf(AVG_TPS_BY_PID, contentKey, serviceName, duration, trimmedDuration)
	case TPSDOD:
		return fmt.Sprintf(TPS_DOD_BY_PID, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case TPSWOW:
		return fmt.Sprintf(TPS_WOW_BY_PID, contentKey, serviceName, duration, contentKey, serviceName, duration, contentKey, serviceName, duration)

	}
	return ""
}
func QueryPidRangePromql(duration string, queryType QueryType, contentKey string, serviceName string) string {
	contentKey = EscapeRegexp(contentKey)
	switch queryType {
	case TPSData:
		trimmedDuration := strings.TrimSuffix(duration, "m")
		return fmt.Sprintf(TPS_DATA_BY_PID, contentKey, serviceName, duration, trimmedDuration)
	case LatencyData:
		return fmt.Sprintf(LATENCY_DATA_BY_PID, contentKey, serviceName, duration, contentKey, serviceName, duration)
	case ErrorData:
		return fmt.Sprintf(ERROR_DATA_BY_PID, contentKey, serviceName, duration, contentKey, serviceName, duration)
	default:
		return ""
	}

}

func QueryLogPromql(duration string, queryType QueryType, containerIds []string) string {
	escapedKeys := make([]string, len(containerIds))
	for i, key := range containerIds {
		escapedKeys[i] = EscapeRegexp(key)
	}
	// 使用 strings.Join 生成正则表达式模式
	regexPattern := strings.Join(escapedKeys, "|")

	switch queryType {
	case AvgLog:
		return fmt.Sprintf(`(
  sum(
    increase(
      originx_logparser_level_count_total{
        pod_name=~"%s",level=~"error|critical"
      }[%s]
    )
  ) by(pod_name)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pod_name=~"%s"}[%s]
      )
    ) by(pod_name)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pod_name=~"%s"}[%s]
  )
) by(pod_name)`, regexPattern, duration, regexPattern, duration, regexPattern, duration)
	case LogDOD:
		return fmt.Sprintf(`((
  sum(
    increase(
      originx_logparser_level_count_total{
        pod_name=~"%s",level=~"error|critical"
      }[%s]
    )
  ) by(pod_name)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pod_name=~"%s"}[%s]
      )
    ) by(pod_name)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pod_name=~"%s"}[%s]
  )
) by(pod_name) -(
  sum(
    increase(
      originx_logparser_level_count_total{
        pod_name=~"%s",level=~"error|critical"
      }[%s] offset 24h
    )
  ) by(pod_name)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pod_name=~"%s"}[%s] offset 24h
      )
    ) by(pod_name)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pod_name=~"%s"}[%s]offset 24h
  )
) by(pod_name))/(
  sum(
    increase(
      originx_logparser_level_count_total{
        pod_name=~"%s",level=~"error|critical"
      }[%s]offset 24h
    )
  ) by(pod_name)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pod_name=~"%s"}[%s]offset 24h
      )
    ) by(pod_name)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pod_name=~"%s"}[%s]offset 24h
  )
) by(pod_name)`, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration)
	case LogWOW:
		return fmt.Sprintf(`((
  sum(
    increase(
      originx_logparser_level_count_total{
        pod_name=~"%s",level=~"error|critical"
      }[%s]
    )
  ) by(pod_name)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pod_name=~"%s"}[%s]
      )
    ) by(pod_name)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pod_name=~"%s"}[%s]
  )
) by(pod_name) -(
  sum(
    increase(
      originx_logparser_level_count_total{
        pod_name=~"%s",level=~"error|critical"
      }[%s] offset 7d
    )
  ) by(pod_name)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pod_name=~"%s"}[%s] offset 7d
      )
    ) by(pod_name)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pod_name=~"%s"}[%s]offset 7d
  )
) by(pod_name))/(
  sum(
    increase(
      originx_logparser_level_count_total{
        pod_name=~"%s",level=~"error|critical"
      }[%s]offset 7d
    )
  ) by(pod_name)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pod_name=~"%s"}[%s]offset 7d
      )
    ) by(pod_name)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pod_name=~"%s"}[%s]offset 7d
  )
) by(pod_name)`, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration)
	case ServiceAvgLog:
		return fmt.Sprintf(`(
  sum(
    increase(
      originx_logparser_level_count_total{
        pod_name=~"%s",level=~"error|critical"
      }[%s]
    )
  ) by(pod_name)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pod_name=~"%s"}[%s]
      )
    ) by(pod_name)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pod_name=~"%s"}[%s]
  )
)or 0 `, regexPattern, duration, regexPattern, duration, regexPattern, duration)
	}
	return ""
}
func QueryLogByContainerIdPromql(duration string, queryType QueryType, containerIds []string) string {
	escapedKeys := make([]string, len(containerIds))
	for i, key := range containerIds {
		escapedKeys[i] = EscapeRegexp(key)
	}
	// 使用 strings.Join 生成正则表达式模式
	regexPattern := strings.Join(escapedKeys, "|")

	switch queryType {
	case AvgLog:
		return fmt.Sprintf(`(
  sum(
    increase(
      originx_logparser_level_count_total{
        container_id=~"%s"
      }[%s]
    )
  ) by(container_id)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{container_id=~"%s"}[%s]
      )
    ) by(container_id)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{container_id=~"%s"}[%s]
  )
) by(container_id)or 0`, regexPattern, duration, regexPattern, duration, regexPattern, duration)
	case LogDOD:
		return fmt.Sprintf(`((
  sum(
    increase(
      originx_logparser_level_count_total{
        container_id=~"%s"
      }[%s]
    )
  ) by(container_id)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{container_id=~"%s"}[%s]
      )
    ) by(container_id)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{container_id=~"%s"}[%s]
  )
) by(container_id) -(
  sum(
    increase(
      originx_logparser_level_count_total{
        container_id=~"%s"
      }[%s] offset 24h
    )
  ) by(container_id)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{container_id=~"%s"}[%s] offset 24h
      )
    ) by(container_id)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{container_id=~"%s"}[%s] offset 24h
  )
) by(container_id))/(
  sum(
    increase(
      originx_logparser_level_count_total{
        container_id=~"%s"
      }[%s] offset 24h
    )
  ) by(container_id)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{container_id=~"%s"}[%s] offset 24h
      )
    ) by(container_id)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{container_id=~"%s"}[%s] offset 24h
  )
) by(container_id)`, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration)
	case LogWOW:
		return fmt.Sprintf(`((
  sum(
    increase(
      originx_logparser_level_count_total{
        container_id=~"%s"
      }[%s]
    )
  ) by(container_id)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{container_id=~"%s"}[%s]
      )
    ) by(container_id)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{container_id=~"%s"}[%s]
  )
) by(container_id) -(
  sum(
    increase(
      originx_logparser_level_count_total{
        container_id=~"%s"
      }[%s] offset 7d
    )
  ) by(container_id)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{container_id=~"%s"}[%s] offset 7d
      )
    ) by(container_id)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{container_id=~"%s"}[%s] offset 7d
  )
) by(container_id))/(
  sum(
    increase(
      originx_logparser_level_count_total{
        container_id=~"%s"
      }[%s] offset 7d
    )
  ) by(container_id)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{container_id=~"%s"}[%s] offset 7d
      )
    ) by(container_id)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{container_id=~"%s"}[%s] offset 7d
  )
) by(container_id)`, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration)
	}
	return ""
}
func QueryLogByPidPromql(duration string, queryType QueryType, containerIds []string) string {
	escapedKeys := make([]string, len(containerIds))
	for i, key := range containerIds {
		escapedKeys[i] = EscapeRegexp(key)
	}
	// 使用 strings.Join 生成正则表达式模式
	regexPattern := strings.Join(escapedKeys, "|")

	switch queryType {
	case AvgLog:
		return fmt.Sprintf(`(
  sum(
    increase(
      originx_logparser_level_count_total{
        pid=~"%s",level=~"error|critical"
      }[%s]
    )
  ) by(pid)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pid=~"%s"}[%s]
      )
    ) by(pid)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pid=~"%s"}[%s]
  )
) by(pid) or 0`, regexPattern, duration, regexPattern, duration, regexPattern, duration)
	case LogDOD:
		return fmt.Sprintf(`((
  sum(
    increase(
      originx_logparser_level_count_total{
        pid=~"%s",level=~"error|critical"
      }[%s]
    )
  ) by(pid)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pid=~"%s"}[%s]
      )
    ) by(pid)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pid=~"%s"}[%s]
  )
) by(pid) -(
  sum(
    increase(
      originx_logparser_level_count_total{
        pid=~"%s",level=~"error|critical"
      }[%s] offset 24h
    )
  ) by(pid)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pid=~"%s"}[%s] offset 24h
      )
    ) by(pid)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pid=~"%s"}[%s]offset 24h
  )
) by(pid))/(
  sum(
    increase(
      originx_logparser_level_count_total{
        pid=~"%s",level=~"error|critical"
      }[%s]offset 24h
    )
  ) by(pid)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pid=~"%s"}[%s]offset 24h
      )
    ) by(pid)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pid=~"%s"}[%s]offset 24h
  )
) by(pid)`, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration)
	case LogWOW:
		return fmt.Sprintf(`((
  sum(
    increase(
      originx_logparser_level_count_total{
        pid=~"%s",level=~"error|critical"
      }[%s]
    )
  ) by(pid)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pid=~"%s"}[%s]
      )
    ) by(pid)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pid=~"%s"}[%s]
  )
) by(pid) -(
  sum(
    increase(
      originx_logparser_level_count_total{
        pid=~"%s",level=~"error|critical"
      }[%s] offset 7d
    )
  ) by(pid)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pid=~"%s"}[%s] offset 7d
      )
    ) by(pid)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pid=~"%s"}[%s]offset 7d
  )
) by(pid))/(
  sum(
    increase(
      originx_logparser_level_count_total{
        pid=~"%s",level=~"error|critical"
      }[%s]offset 7d
    )
  ) by(pid)
    +
  (
    sum(
      increase(
        originx_logparser_exception_count_total{pid=~"%s"}[%s]offset 7d
      )
    ) by(pid)
      or
    0
  )
)
  or
sum(
  increase(
    originx_logparser_exception_count_total{pid=~"%s"}[%s]offset 7d
  )
) by(pid)`, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration, regexPattern, duration)
	}
	return ""
}

const SERVICES_POD_INSTANCE = `sum(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s"}) by (pod,svc_name,node_name,pid)`

const SERVICES_CONTAINER_INSTANCE = `sum(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",pod=""}) by (container_id,svc_name,node_name,pid)`

const SERVICES_PID_INSTANCE = `sum(kindling_span_trace_duration_nanoseconds_count{svc_name=~"%s",pod="",container_id=""}) by (pid,svc_name,node_name)`

func QueryServiceInstancePromql(queryType QueryType, svcNames []string) string {
	escapedKeys := make([]string, len(svcNames))
	for i, key := range svcNames {
		escapedKeys[i] = EscapeRegexp(key)
	}
	// 使用 strings.Join 生成正则表达式模式
	regexPattern := strings.Join(escapedKeys, "|")
	switch queryType {
	case ServiceInstancePod:
		return fmt.Sprintf(SERVICES_POD_INSTANCE, regexPattern)
	case ServiceInstanceContainer:
		return fmt.Sprintf(SERVICES_CONTAINER_INSTANCE, regexPattern)
	case ServiceInstancePid:
		return fmt.Sprintf(SERVICES_PID_INSTANCE, regexPattern)
	}
	return ""
}
