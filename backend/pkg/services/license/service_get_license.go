// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"errors"

	"192.168.1.22/apo/license-manager/pkg/encrypt"
	"192.168.1.22/apo/license-manager/pkg/license"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
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
		return response.GetCurrentLicenseResp{}, model.NewErrWithMessage(errors.New("can not found local activated license"), code.ActivatedLicenseNotFound)
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
