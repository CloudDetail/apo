// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type UploadLicenseRequest struct {
	License        string `form:"license"`
	SignPubKey     string `form:"signPubKey"`
	DecryptPrivKey string `form:"decryptPrivKey"`
	LicensePath    string `form:"licensePath"`
	PubKeyPath     string `form:"pubKeyPath"`
	PrivKeyPath    string `form:"privKeyPath"`
}
