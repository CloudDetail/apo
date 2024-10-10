package amconfig

func HasEmailOrWebhookConfig(r Receiver) bool {
	if r.EmailConfigs != nil {
		return true
	} else if r.WebhookConfigs != nil {
		return true
	}

	return false
}
