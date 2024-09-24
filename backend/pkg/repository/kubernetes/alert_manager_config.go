package kubernetes

import (
	"encoding/json"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	MailReceiver    = "email"
	WebHookReceiver = "webhook"
)

type AlertManagerConfig struct {
	ReceiverList []*request.AMConfigReceiver

	UnsupportReceiver []*amconfig.Receiver
}

func ParseAlertManagerConfig(strContent string) (*AlertManagerConfig, error) {
	cfg, err := amconfig.Load(strContent)
	if err != nil {
		return nil, err
	}

	var receiverList = make([]*request.AMConfigReceiver, 0)
	var unsupportList = make([]*amconfig.Receiver, 0)
	for i := 0; i < len(cfg.Receivers); i++ {
		r := GetReceiverVOFromReceiverDef(cfg.Receivers[i])
		if r == nil {
			unsupportList = append(unsupportList, &cfg.Receivers[i])
		} else {
			receiverList = append(receiverList, r)
		}
	}

	return &AlertManagerConfig{
		ReceiverList:      receiverList,
		UnsupportReceiver: unsupportList,
	}, nil
}

func GetReceiverVOFromReceiverDef(r amconfig.Receiver) *request.AMConfigReceiver {
	rType := GetRTypeFromReceiver(r)
	switch rType {
	case "webhook":
		data, err := json.Marshal(r.WebhookConfigs)
		if err != nil {
			return nil
		}
		var cfgs []*amconfig.WebhookConfig
		err = json.Unmarshal(data, &cfgs)
		if err != nil {
			return nil
		}
		return &request.AMConfigReceiver{
			Name:           r.Name,
			RType:          rType,
			WebhookConfigs: cfgs,
		}
	case "email":
		data, err := json.Marshal(r.WebhookConfigs)
		if err != nil {
			return nil
		}
		var cfgs []*amconfig.EmailConfig
		err = json.Unmarshal(data, &cfgs)
		if err != nil {
			return nil
		}
		return &request.AMConfigReceiver{
			Name:         r.Name,
			RType:        rType,
			EmailConfigs: cfgs,
		}
	default:
		return nil
	}
}

func GetRTypeFromReceiver(r amconfig.Receiver) string {
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
