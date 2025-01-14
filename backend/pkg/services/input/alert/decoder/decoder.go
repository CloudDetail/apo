// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"fmt"

	ainput "github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

type InputType string

const (
	JSONType       InputType = "json"
	PrometheusType InputType = "prometheus"
)

type Decoder interface {
	Decode(sourceFrom ainput.SourceFrom, data []byte) ([]ainput.AlertEvent, error)
}

var decoders = map[InputType]Decoder{
	JSONType: JsonDecoder{},
}

type ErrDecoderNotFound struct {
	InputType string
}

func (e ErrDecoderNotFound) Error() string {
	return fmt.Sprintf("decoder not found: %s", e.InputType)
}

func Decode(sourceFrom ainput.SourceFrom, data []byte) ([]ainput.AlertEvent, error) {
	decoder, ok := decoders[InputType(sourceFrom.SourceType)]
	if !ok {
		return nil, ErrDecoderNotFound{InputType: sourceFrom.SourceType}
	}
	return decoder.Decode(sourceFrom, data)
}