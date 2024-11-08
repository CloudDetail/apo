package jaeger

import (
	"net/http"
)

type JaegerRepo interface {
	GetSingleTrace(traceId string) (string, error)
}

type jaegerRepo struct {
	cli *http.Client
}

func New() (JaegerRepo, error) {
	client := &http.Client{}
	return &jaegerRepo{cli: client}, nil
}
