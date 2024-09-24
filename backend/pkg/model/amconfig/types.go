package amconfig

func GetRTypeFromReceiver(r Receiver) string {
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
