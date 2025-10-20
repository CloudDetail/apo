// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"errors"
	"time"

	"192.168.1.22/apo/license-manager/pkg/encrypt"
	"192.168.1.22/apo/license-manager/pkg/license"
	mylicense "github.com/CloudDetail/apo/backend/pkg/model/license"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

const (
	licenseStatusInvalid = iota - 1
	licenseStatusNotActivated
	licenseStatusActivated
	licenseStatusExpired
)

func (s *service) GetLicenses() (response.GetLicenseResp, error) {
	licenses, err := s.dbRepo.GetLicense()
	if err != nil {
		return nil, err
	}

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -1)
	seatsCurrentlyInUse, _ := s.promRepo.GetNodesCount(startTime.UnixMicro(), endTime.UnixMicro())

	resp := make([]response.LicenseResp, 0, len(licenses))
	for _, lic := range licenses {
		licResp := response.LicenseResp{}
		if lic.Activated {
			licResp.Active = licenseStatusActivated
		} else {
			licResp.Active = licenseStatusNotActivated
		}
		bLic := &license.BasicLicense{}

		decryptor, err := encrypt.NewRSAFromKeys([]byte(lic.EncryptPublicKey), []byte(lic.DecryptPrivateKey))
		if err != nil {
			return nil, err
		}
		err = license.ParseWithLicense([]byte(lic.Token), bLic, decryptor, nil)
		if err != nil {
			return nil, err
		}

		licResp.Features = bLic.Features
		licResp.MaxNode = bLic.MaxNodes
		licResp.LicenseID = bLic.LicenseID
		if expire, err := bLic.CheckExpireDate(); err != nil && errors.Is(err, license.ErrLicenseNotActivated) {
			licResp.Active = licenseStatusNotActivated
		} else if expire {
			licResp.Active = licenseStatusExpired
		}
		licResp.ExpireAt = bLic.ActivatedAt + bLic.ExpireDuration
		licResp.IssuedAt = bLic.IssuedAt
		licResp.SeatsInLicense = bLic.MaxNodes
		licResp.SeatsCurrentlyInUse = int(seatsCurrentlyInUse)

		resp = append(resp, licResp)
	}
	return resp, nil
}

func (s *service) GetCurrentLicense() (response.GetCurrentLicenseResp, error) {
	lic, err := s.dbRepo.GetCurrentLicense()
	if err != nil {
		return response.GetCurrentLicenseResp{}, err
	}
	if lic.ID == 0 {
		return response.GetCurrentLicenseResp{
			Active: licenseStatusNotActivated,
		}, nil
	}

	licResp := response.LicenseResp{}
	if lic.Activated {
		licResp.Active = licenseStatusActivated
	} else {
		licResp.Active = licenseStatusNotActivated
	}
	bLic := &license.BasicLicense{}

	decryptor, err := encrypt.NewRSAFromKeys([]byte(lic.EncryptPublicKey), []byte(lic.DecryptPrivateKey))
	if err != nil {
		return response.GetCurrentLicenseResp{}, err
	}
	err = license.ParseWithLicense([]byte(lic.Token), bLic, decryptor, nil)
	if err != nil {
		return response.GetCurrentLicenseResp{}, err
	}

	licResp.Features = bLic.Features
	licResp.MaxNode = bLic.MaxNodes
	licResp.LicenseID = bLic.LicenseID
	expireAt := bLic.ActivatedAt + bLic.ExpireDuration
	licResp.ExpireAt = expireAt
	licResp.SeatsInLicense = bLic.MaxNodes
	licResp.IssuedAt = bLic.IssuedAt
	licResp.Type = bLic.Type

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -1)
	seatsCurrentlyInUse, _ := s.promRepo.GetNodesCount(startTime.UnixMicro(), endTime.UnixMicro())
	licResp.SeatsCurrentlyInUse = int(seatsCurrentlyInUse)

	if expire, err := bLic.CheckExpireDate(); err != nil && errors.Is(err, license.ErrLicenseNotActivated) {
		licResp.Active = licenseStatusNotActivated
	} else if expire {
		licResp.Active = licenseStatusExpired
	}

	return response.GetCurrentLicenseResp(licResp), nil
}

func (s *service) GetActivation() (mylicense.Activation, error) {
	activations, err := s.dbRepo.GetActivations()
	if err != nil {
		return mylicense.Activation{}, err
	}

	if len(activations) == 0 {
		return mylicense.Activation{}, nil
	}

	return activations[0], nil
}
