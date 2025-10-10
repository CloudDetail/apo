// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

type LicenseResp struct {
	LicenseID           string   `json:"licenseId"`
	Active              int      `json:"active"`
	ExpireAt            int64    `json:"expiresAt"`
	MaxNode             int      `json:"maxNodes"`
	Features            []string `json:"features"`
	SeatsCurrentlyInUse int      `json:"seatsCurrentlyInUse"`
	SeatsInLicense      int      `json:"seatsInLicense"`
	IssuedAt            int64    `json:"issuedAt"`
}

type GetLicenseResp []LicenseResp
type GetCurrentLicenseResp LicenseResp
