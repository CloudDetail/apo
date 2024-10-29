package amconfig

func HasEmailOrWebhookConfig(r Receiver) bool {
	if r.EmailConfigs != nil {
		return true
	} else if r.WebhookConfigs != nil && len(r.WebhookConfigs) > 0 {
		return true
	}

	return false
}
