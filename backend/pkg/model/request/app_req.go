// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

import "mime/multipart"

type UploadLogoRequest struct {
	LogoText  string               `form:"logoText" binding:"required,max=15"`
	LogoImage multipart.FileHeader `form:"logoImage" binding:"required"`
}
