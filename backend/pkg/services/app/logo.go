// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"io"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) UploadLogo(req *request.UploadLogoRequest) error {
	maxFileSize := int64(2 << 20)
	if req.LogoImage.Size > maxFileSize {
		return core.Error(code.GetLogoError, "logo file size exceeds 2M")
	}

	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
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

	mimeType := http.DetectContentType(buffer)
	if _, valid := allowedTypes[mimeType]; !valid {
		return core.Error(code.LogoTypeError, fmt.Sprintf("unsupported logo type %s", mimeType))
	}

	image, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	logo := &model.Logo{
		LogoImage: image,
		LogoText:  req.LogoText,
	}

	return s.dbRepo.CreateLogo(logo)
}
func (s *service) GetLogo() (response.GetLogoResp, error) {
	logo, err := s.dbRepo.GetLogo()
	return response.GetLogoResp(*logo), err
}
