// Copyright 2023 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package amnotify

import (
	"errors"
	"log/slog"

	commoncfg "github.com/prometheus/common/config"
	"github.com/prometheus/common/promslog"
	"gopkg.in/yaml.v3"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/notify/email"
	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/notify/wechat"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
)

type ReceiverMap struct {
	receivers map[string][]notify.Integration
}

func BuildReceivers(ncs []amconfig.Receiver, tmpl *template.Template, logger *slog.Logger, httpOpts ...commoncfg.HTTPClientOption) (ReceiverMap, error) {
	receivers := map[string][]notify.Integration{}
	var errs error
	for _, nc := range ncs {
		integrations, err := buildReceiverIntegrations(nc, tmpl, logger, httpOpts...)
		if err != nil {
			logger.Error("Error building integrations", "err", err)
			errs = errors.Join(errs, err)
		}
		receivers[nc.Name] = integrations
	}
	return ReceiverMap{receivers: receivers}, errs
}

// buildReceiverIntegrations builds a list of integration notifiers off of a
// receiver config.
func buildReceiverIntegrations(nc amconfig.Receiver, tmpl *template.Template, logger *slog.Logger, httpOpts ...commoncfg.HTTPClientOption) ([]notify.Integration, error) {
	if logger == nil {
		logger = promslog.NewNopLogger()
	}

	rc := &config.Receiver{}
	cfgBytes, err := yaml.Marshal(nc)
	if err != nil {
		// TODO do something
	}
	err = yaml.Unmarshal(cfgBytes, rc)
	if err != nil {
		// TODO do something
	}

	var (
		errs         types.MultiError
		integrations []notify.Integration
		add          = func(name string, i int, rs notify.ResolvedSender, f func(l *slog.Logger) (notify.Notifier, error)) {
			n, err := f(logger.With("integration", name))
			if err != nil {
				errs.Add(err)
				return
			}
			integrations = append(integrations, notify.NewIntegration(n, rs, name, i, nc.Name))
		}
	)

	for i, c := range rc.WebhookConfigs {
		add("webhook", i, c, func(l *slog.Logger) (notify.Notifier, error) { return webhook.New(c, tmpl, l, httpOpts...) })
	}
	for i, c := range rc.EmailConfigs {
		add("email", i, c, func(l *slog.Logger) (notify.Notifier, error) { return email.New(c, tmpl, l), nil })
	}

	for i, c := range nc.DingTalkConfigs {
		// TODO transform to webhook config
		cfg := &config.WebhookConfig{
			NotifierConfig: config.NotifierConfig{},
			HTTPConfig:     &commoncfg.HTTPClientConfig{},
			URL:            &config.SecretURL{},
			URLFile:        "",
			MaxAlerts:      0,
			Timeout:        0,
		}

		add("dingtalk", i, c, func(l *slog.Logger) (notify.Notifier, error) { return webhook.New(cfg, tmpl, l, httpOpts...) })
	}

	// for i, c := range nc.PagerdutyConfigs {
	// 	add("pagerduty", i, c, func(l *slog.Logger) (notify.Notifier, error) { return pagerduty.New(c, tmpl, l, httpOpts...) })
	// }
	// for i, c := range nc.OpsGenieConfigs {
	// 	add("opsgenie", i, c, func(l *slog.Logger) (notify.Notifier, error) { return opsgenie.New(c, tmpl, l, httpOpts...) })
	// }
	for i, c := range rc.WechatConfigs {
		add("wechat", i, c, func(l *slog.Logger) (notify.Notifier, error) { return wechat.New(c, tmpl, l, httpOpts...) })
	}
	// for i, c := range nc.SlackConfigs {
	// 	add("slack", i, c, func(l *slog.Logger) (notify.Notifier, error) { return slack.New(c, tmpl, l, httpOpts...) })
	// }
	// for i, c := range nc.VictorOpsConfigs {
	// 	add("victorops", i, c, func(l *slog.Logger) (notify.Notifier, error) { return victorops.New(c, tmpl, l, httpOpts...) })
	// }
	// for i, c := range nc.PushoverConfigs {
	// 	add("pushover", i, c, func(l *slog.Logger) (notify.Notifier, error) { return pushover.New(c, tmpl, l, httpOpts...) })
	// }
	// for i, c := range nc.SNSConfigs {
	// 	add("sns", i, c, func(l *slog.Logger) (notify.Notifier, error) { return sns.New(c, tmpl, l, httpOpts...) })
	// }
	// for i, c := range nc.TelegramConfigs {
	// 	add("telegram", i, c, func(l *slog.Logger) (notify.Notifier, error) { return telegram.New(c, tmpl, l, httpOpts...) })
	// }
	// for i, c := range nc.DiscordConfigs {
	// 	add("discord", i, c, func(l *slog.Logger) (notify.Notifier, error) { return discord.New(c, tmpl, l, httpOpts...) })
	// }
	// for i, c := range nc.WebexConfigs {
	// 	add("webex", i, c, func(l *slog.Logger) (notify.Notifier, error) { return webex.New(c, tmpl, l, httpOpts...) })
	// }
	// for i, c := range nc.MSTeamsConfigs {
	// 	add("msteams", i, c, func(l *slog.Logger) (notify.Notifier, error) { return msteams.New(c, tmpl, l, httpOpts...) })
	// }

	if errs.Len() > 0 {
		return nil, &errs
	}
	return integrations, nil
}
