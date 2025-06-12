// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"io"
	"mime"
	"net/http"

	"bytes"
	"encoding/xml"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/microcosm-cc/bluemonday"
)

func (s *service) UploadLogo(req *request.UploadLogoRequest) error {
	maxFileSize := int64(2 << 20) // 2MB
	if req.LogoImage.Size > maxFileSize {
		return core.Error(code.GetLogoError, "logo file size exceeds 2M")
	}

	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/svg+xml":   true,
		"image/x-icon":    true,
		"application/xml": true,
		"text/xml":        true,
	}

	file, err := req.LogoImage.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	if _, err = file.Read(buffer); err != nil {
		return err
	}

	if _, err = file.Seek(0, 0); err != nil {
		return err
	}

	detectedType := http.DetectContentType(buffer)
	mimeType, _, err := mime.ParseMediaType(detectedType)
	if err != nil {
		mimeType = detectedType
	}

	if mimeType == "application/xml" || mimeType == "text/xml" {
		if isSVG(buffer) {
			mimeType = "image/svg+xml"
		}
	}

	if _, valid := allowedTypes[mimeType]; !valid {
		return core.Error(code.LogoTypeError, fmt.Sprintf("unsupported logo type %s", mimeType))
	}

	image, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if mimeType == "image/svg+xml" {
		sanitized, err := sanitizeSVG(image)
		if err != nil {
			return err
		}
		image = sanitized
	}

	logo := &model.Logo{
		LogoImage: image,
		LogoText:  req.LogoText,
	}

	return s.dbRepo.CreateLogo(logo)
}

func isSVG(data []byte) bool {
	r := bytes.NewReader(data)
	decoder := xml.NewDecoder(r)

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		if start, ok := token.(xml.StartElement); ok {
			if start.Name.Local == "svg" {
				for _, attr := range start.Attr {
					if attr.Name.Local == "xmlns" && attr.Value == "http://www.w3.org/2000/svg" {
						return true
					}
				}
			}
			break
		}
	}
	return false
}

func sanitizeSVG(data []byte) ([]byte, error) {
	p := bluemonday.NewPolicy()
	p.RequireNoFollowOnLinks(true)
	p.RequireNoReferrerOnLinks(true)

	p.AllowElements(
		"svg", "g", "path", "rect", "circle", "ellipse", "line", "polygon", "polyline",
		"text", "tspan", "defs", "linearGradient", "radialGradient", "stop", "image",
	)

	p.AllowAttrs(
		"width", "height", "x", "y", "cx", "cy", "r", "rx", "ry",
		"d", "points", "fill", "stroke", "stroke-width", "stroke-linecap",
		"stroke-linejoin", "font-size", "text-anchor", "href", "xlink:href", "viewBox",
	).Globally()

	p.AllowAttrs("xmlns", "xmlns:xlink").OnElements("svg")

	p.AllowAttrs("xlink:href").OnElements("image")

	sanitized := p.SanitizeBytes(data)
	return sanitized, nil
}
func (s *service) GetLogo() (response.GetLogoResp, error) {
	logo, err := s.dbRepo.GetLogo()
	return response.GetLogoResp(*logo), err
}
