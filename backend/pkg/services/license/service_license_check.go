// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"errors"
	"time"

	"192.168.1.22/apo/license-manager/pkg/encrypt"
	"192.168.1.22/apo/license-manager/pkg/license"
	"go.uber.org/zap"
)

var ErrNoLicense = errors.New("no license detected")

func (s *service) CheckLicenseNodes() (exceed bool, err error) {
	bLic, err := s.getLicense()
	if err != nil {
		return false, err
	}
	if bLic == nil {
		s.logger.Error("", zap.Error(err))
		return false, ErrNoLicense
	}
	endTime := time.Now()
	startTime := endTime.Add(-time.Minute * 10)
	c, err := s.promRepo.GetNodesCount(startTime.UnixMicro(), endTime.UnixMicro())
	if err != nil {
		s.logger.Error("check license nodes error ", zap.Error(err))
		return false, err
	}

	if exceed := bLic.CheckMaxNodes(c); exceed {
		s.logger.Error("max node exceeds", zap.Int("current nodes", c), zap.Int("max nodes", bLic.MaxNodes))
		return true, nil
	}
	return false, nil
}

func (s *service) CheckLicenseDate() (expire bool, err error) {
	bLic, err := s.getLicense()
	if err != nil {
		return false, err
	}

	if bLic == nil {
		s.logger.Error("", zap.Error(err))
		return false, ErrNoLicense
	}

	if expire, err := bLic.CheckExpireDate(); err != nil {
		s.logger.Error("", zap.Error(err))
		return false, err
	} else if expire {
		return true, nil
	}
	return false, nil
}

func (s *service) checkLoop() {
	ticker := time.NewTicker(time.Minute * 10)

	for range ticker.C {
		s.CheckLicenseNodes()
		s.CheckLicenseDate()
	}
}

func (s *service) getLicense() (*license.BasicLicense, error) {
	lic, err := s.dbRepo.GetCurrentLicense()
	if err != nil {
		return &license.BasicLicense{}, err
	}
	if len(lic.Token) == 0 {
		return nil, nil
	}
	var bLic license.BasicLicense
	decryptor, err := encrypt.NewRSAFromKeys([]byte(lic.EncryptPublicKey), []byte(lic.DecryptPrivateKey))
	if err != nil {
		s.logger.Error("key error", zap.Error(err))
		return &license.BasicLicense{}, err
	}

	err = license.ParseWithLicense([]byte(lic.Token), &bLic, decryptor, nil)
	if err != nil {
		s.logger.Error("parse license error", zap.Error(err))
		return &license.BasicLicense{}, err
	}

	return &bLic, nil
}

func (s *service) CheckFeature(features ...string) (pass bool, err error) {
	bLic, err := s.getLicense()
	if err != nil {
		return false, err
	}

	if bLic == nil {
		s.logger.Error("", zap.Error(err))
		return false, ErrNoLicense
	}

	if has := bLic.CheckFeatures(features...); !has {
		return false, nil
	}

	return true, nil
}
