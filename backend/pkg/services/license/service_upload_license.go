// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"fmt"
	"os"
	"path/filepath"

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

	err := checkAndCreateDir(req.PrivKeyPath)
	if err != nil {
		return err
	}

	err = checkAndCreateDir(req.PubKeyPath)
	if err != nil {
		return err
	}

	err = checkAndCreateDir(req.LicensePath)
	if err != nil {
		return err
	}

	err = os.WriteFile(req.PrivKeyPath, []byte(req.DecryptPrivKey), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(req.PubKeyPath, []byte(req.SignPubKey), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(req.LicensePath, []byte(req.License), 0644)
	if err != nil {
		return err
	}

	return nil
}

func checkAndCreateDir(path string) error {
	dir := filepath.Dir(path)

	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			return fmt.Errorf("failed to stat directory: %w", err)
		}
	}
	return nil
}
