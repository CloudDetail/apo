package kubernetes

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	promcfg "github.com/prometheus/alertmanager/config"
)

const (
	MailReceiver    = "email"
	WebHookReceiver = "webhook"
)

type AlertManagerConfig struct {
	Ref *promcfg.Config

	ReceiverMap  map[string]*request.AMConfigReceiver
	ReceiverList []*request.AMConfigReceiver
}

func ParseAlertManagerConfig(strContent string) (*AlertManagerConfig, error) {
	cfg, err := promcfg.Load(strContent)
	if err != nil {
		return nil, err
	}

	var receiverMap = make(map[string]*request.AMConfigReceiver)
	for i := 0; i < len(cfg.Receivers); i++ {
		receiverMap[cfg.Receivers[i].Name] = &request.AMConfigReceiver{
			Receiver: &cfg.Receivers[i],
			RType:    GetRTypeFromReceiver(cfg.Receivers[i]),
		}
	}

	return &AlertManagerConfig{
		Ref:         cfg,
		ReceiverMap: receiverMap,
	}, nil
}

func GetRTypeFromReceiver(r promcfg.Receiver) string {
	if r.DiscordConfigs != nil {
		return "discord"
	} else if r.EmailConfigs != nil {
		return "email"
	} else if r.WebhookConfigs != nil {
		return "webhook"
	} else if r.WechatConfigs != nil {
		return "wechat"
	} else if r.MSTeamsConfigs != nil {
		return "msteam"
	} else if r.OpsGenieConfigs != nil {
		return "opsGenie"
	} else if r.PagerdutyConfigs != nil {
		return "pagerduty"
	} else if r.PushoverConfigs != nil {
		return "pushover"
	} else if r.SNSConfigs != nil {
		return "sns"
	} else if r.SlackConfigs != nil {
		return "slack"
	} else if r.TelegramConfigs != nil {
		return "telegram"
	} else if r.VictorOpsConfigs != nil {
		return "victorOps"
	} else if r.WebexConfigs != nil {
		return "webex"
	}
	return "unknown"
}
