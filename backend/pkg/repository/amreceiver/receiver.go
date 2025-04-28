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

package amreceiver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	commoncfg "github.com/prometheus/common/config"
	"github.com/prometheus/common/promslog"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/notify/email"
	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/notify/wechat"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
)

type Receivers interface {
	HandleAlertCheckRecord(ctx context.Context, record *model.WorkflowRecord) error
	// TODO
	// HandleAlertResolvedRecord(ctx context.Context, record *model.WorkflowRecord) error

	GetAMConfigReceiver(filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int)
	AddAMConfigReceiver(receiver amconfig.Receiver) error
	UpdateAMConfigReceiver(receiver amconfig.Receiver, oldName string) error
	DeleteAMConfigReceiver(name string) error
}

type AMReceivers struct {
	database  database.Repo
	receivers map[string][]notify.Integration

	externalURL string
	logger      *slog.Logger
}

func (r *AMReceivers) GetAMConfigReceiver(filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int) {
	return r.database.GetAMConfigReceiver(filter, pageParam)
}

func (r *AMReceivers) AddAMConfigReceiver(receiver amconfig.Receiver) error {
	err := r.database.AddAMConfigReceiver(receiver)
	if err != nil {
		return err
	}

	receivers, _ := r.database.GetAMConfigReceiver(nil, nil)
	return r.UpdateReceivers(receivers)
}

func (r *AMReceivers) UpdateAMConfigReceiver(receiver amconfig.Receiver, oldName string) error {
	err := r.database.UpdateAMConfigReceiver(receiver, oldName)
	if err != nil {
		return err
	}
	receivers, _ := r.database.GetAMConfigReceiver(nil, nil)
	return r.UpdateReceivers(receivers)
}

func (r *AMReceivers) DeleteAMConfigReceiver(name string) error {
	err := r.database.DeleteAMConfigReceiver(name)
	if err != nil {
		return err
	}
	receivers, _ := r.database.GetAMConfigReceiver(nil, nil)
	return r.UpdateReceivers(receivers)
}

func (r *AMReceivers) UpdateReceivers(receivers []amconfig.Receiver) error {
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return err
	}
	tmpl.ExternalURL, err = url.Parse(r.externalURL)
	if err != nil {
		return err
	}

	newReceiver, err := buildAMReceivers(receivers, tmpl, r.logger)
	if err != nil {
		return err
	}
	r.receivers = newReceiver.receivers
	return nil
}

func (r *AMReceivers) HandleAlertCheckRecord(ctx context.Context, record *model.WorkflowRecord) error {
	if record.WorkflowName != "AlertCheck" {
		return nil
	}
	if record.Output != "false" {
		return nil
	}

	alert, ok := record.InputRef.(alert.AlertEvent)
	if !ok {
		return fmt.Errorf("unexpect inputRef, should be alert.AlertEvent, got %T", record.InputRef)
	}

	var errs error
	for name, integrations := range r.receivers {
		for _, integration := range integrations {
			// TODO set timeout and retry
			_, err := integration.Notify(ctx, alert.ToAMAlert(false))
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("[%s] send alert failed: %w", name, err))
			}
		}
	}

	return errs
}

func SetupReceiver(externalURL string, logger *zap.Logger, dbRepo database.Repo) (Receivers, error) {
	// TODO not support template now
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}
	tmpl.ExternalURL, err = url.Parse(externalURL)
	if err != nil {
		return nil, err
	}

	receivers, _ := dbRepo.GetAMConfigReceiver(nil, nil)

	amReceiver, err := buildAMReceivers(
		receivers,
		tmpl,
		util.NewZapSlogHandler(logger),
	)
	if err != nil {
		amReceiver.database = dbRepo
	}
	return amReceiver, err
}

func buildAMReceivers(ncs []amconfig.Receiver, tmpl *template.Template, logger *slog.Logger, httpOpts ...commoncfg.HTTPClientOption) (*AMReceivers, error) {
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
	return &AMReceivers{receivers: receivers, logger: logger}, errs
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
		return nil, err
	}
	err = yaml.Unmarshal(cfgBytes, rc)
	if err != nil {
		return nil, err
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
		// TODO transform into Function Call
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
