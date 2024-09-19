package prometheus

type EndpointKey struct {
	ContentKey string // URL
	SvcName    string // url所属的服务名
}

func (e EndpointKey) ConvertFromLabels(labels Labels) ConvertFromLabels {
	return EndpointKey{
		SvcName:    labels.SvcName,
		ContentKey: labels.ContentKey,
	}
}

type SQLKey struct {
	Service string `json:"service"`
	// DBSystem -> ${SQL Type}, e.g: Mysql
	DBSystem string `json:"dbSystem"`
	// DBName -> ${database}
	DBName string `json:"dbName"`
	// DBOperation -> ${operation} ${table}, e.g: SELECT trip
	DBOperation string `json:"dbOperation"`
	DBUrl       string `json:"dbUrl"`
}

func (k SQLKey) ConvertFromLabels(labels Labels) ConvertFromLabels {
	return SQLKey{
		Service:     labels.SvcName,
		DBSystem:    labels.DBSystem,
		DBName:      labels.DBName,
		DBOperation: labels.Name,
		DBUrl:       labels.DBUrl,
	}
}

type ServiceKey struct {
	SvcName string // url所属的服务名
}

func (S ServiceKey) ConvertFromLabels(labels Labels) ConvertFromLabels {
	return ServiceKey{
		SvcName: labels.SvcName,
	}
}
