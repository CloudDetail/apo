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
