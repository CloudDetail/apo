// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"go.uber.org/zap"
)

const (
	PRIVATE_KEY_PATH = "./license/keys/private_key.pem"
	PUBLIC_KEY_PATH  = "./license/keys/public_key.pem"
	LICENSE_PATH     = "./license/license.lic"
)


func (s *service) CheckLicenseNodes() {
	conf := config.Get()

	if conf.License.BasicLicense == nil {
		err := conf.ParseLicense()
		if err != nil {
			s.logger.Error("check license nodes error ", zap.Error(err))
		}
	}

	if conf.License.BasicLicense.MaxNodes == -1 {
		return 
	}

	endTime := time.Now()
	startTime := endTime.Add(- time.Minute * 10)
	c, err := s.promRepo.GetNodesCount(startTime.UnixMicro(), endTime.UnixMicro())
	if err != nil {
		s.logger.Error("check license nodes error ", zap.Error(err))
		return
	}

	err = conf.License.BasicLicense.CheckMaxNodes(c)
	if err != nil {
		s.logger.Warn("your nodes count has exceeded limit, please contact us to expand", zap.Error(err))
	}
}

func (s *service) CheckLicenseDate() {
	conf := config.Get()
	if conf.License.BasicLicense == nil {
		err := conf.ParseLicense()
		if err != nil {
			s.logger.Error("check license nodes error ", zap.Error(err))
		}
	}

	err := conf.License.BasicLicense.CheckExpireDate()
	if err != nil {
		s.logger.Warn("your license has been expired, please contact us to renew", zap.Error(err))
	}

	err = conf.License.BasicLicense.CheckTrailDate()
	if err != nil {
		s.logger.Warn("your license trail period has been expired, pleas contact us to renew", zap.Error(err))
	}
}

func (s *service) checkLoop() {
	ticker := time.NewTicker(time.Minute*10)

	for range ticker.C {
		s.CheckLicenseNodes()
		s.CheckLicenseDate()
	}
}