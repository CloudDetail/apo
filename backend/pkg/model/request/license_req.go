// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type UploadLicenseRequest struct {
	Output      string `form:"output" json:"license"`
	LicensePath string `form:"licensePath" json:"licensePath" `
	PubKeyPath  string `form:"pubKeyPath" json:"pubKeyPath"`
	PrivKeyPath string `form:"privKeyPath" json:"privKeyPath"`
}
