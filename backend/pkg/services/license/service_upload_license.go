// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"encoding/json"

	"192.168.1.22/apo/license-manager/pkg/encrypt"
	"192.168.1.22/apo/license-manager/pkg/license"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	mylicense "github.com/CloudDetail/apo/backend/pkg/model/license"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) UploadLicense(req *request.UploadLicenseRequest) error {
	output := &license.OutputFile{}
	err := json.Unmarshal([]byte(req.Output), output)
	if err != nil {
		return err
	}

	encryptor, err := encrypt.NewRSAFromKeys([]byte(output.EncryptPublicKey), []byte(output.DecryptPrivateKey))
	if err != nil {
		return err
	}

	newToken, err := license.Activate([]byte(output.License), encryptor, nil)
	if err != nil {
		return core.Error(code.ActivateLicenseFailed, "activate license failed")
	}

	lic := mylicense.License{
		Token:             string(newToken),
		EncryptPublicKey:  output.EncryptPublicKey,
		DecryptPrivateKey: output.DecryptPrivateKey,
		SignPublicKey:     output.SignPublicKey,
		Activated:         true,
	}

	bLic := &license.BasicLicense{}
	err = license.ParseWithLicense(newToken, bLic, encryptor, nil)
	if err != nil {
		return nil
	}

	activations, err := s.dbRepo.GetActivations()
	if err != nil {
		return err
	}

	if len(activations) == 0 {
		err = s.dbRepo.CreateActivation(&mylicense.Activation{Mode: bLic.DistributionType})
		if err != nil {
			return err
		}
	}

	err = s.dbRepo.CreateLicense(&lic)
	if err != nil {
		return err
	}
	return nil
}
