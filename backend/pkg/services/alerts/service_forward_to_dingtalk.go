// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"log"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/services/alerts/notification"
)

func (s *service) ForwardToDingTalk(ctx core.Context, req *request.ForwardToDingTalkRequest, uuid string) error {
	config, err := s.dbRepo.GetDingTalkReceiver(ctx, uuid)
	if err != nil {
		log.Println("get dingtalk receiver err:", err)
		return err
	}

	// construct builder
	builder, err := notification.NewNotificationBuilder()
	if err != nil {
		log.Println("get dingtalk builder err:", err)
		return err
	}
	dingTalkNotification, err := builder.Build(req)
	if err != nil {
		log.Println("build dingtalk notification err:", err)
		return err
	}
	err = notification.SendNotification(dingTalkNotification, config.URL, config.Secret)
	if err != nil {
		log.Println("send dingtalk notification err:", err)
		return err
	}
	return nil
}
