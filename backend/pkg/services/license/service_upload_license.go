// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"os"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) UploadLicense(req *request.UploadLicenseRequest) error {
	if len(req.PrivKeyPath) == 0 {
		req.PrivKeyPath = PRIVATE_KEY_PATH
	}

	if len(req.LicensePath) == 0 {
		req.LicensePath = LICENSE_PATH
	}

	if len(req.PubKeyPath) == 0 {
		req.PubKeyPath = PUBLIC_KEY_PATH
	}

	err := os.WriteFile(req.PrivKeyPath, []byte(req.DecryptPrivKey), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(req.PubKeyPath, []byte(req.PubKeyPath), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(req.LicensePath, []byte(req.License), 0644)
	if err != nil {
		return err
	}

	return nil
}
