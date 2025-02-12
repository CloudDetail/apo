// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type UploadLicenseRequest struct {
	License        string `form:"license" json:"license"`
	SignPubKey     string `form:"signPubKey" json:"signPubKey"`
	DecryptPrivKey string `form:"decryptPrivKey" json:"decryptPrivKey"`
	LicensePath    string `form:"licensePath" json:"licensePath" `
	PubKeyPath     string `form:"pubKeyPath" json:"pubKeyPath"`
	PrivKeyPath    string `form:"privKeyPath" json:"privKeyPath"`
}
