// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package receiver

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"sync"
	"time"

	commoncfg "github.com/prometheus/common/config"
	"github.com/prometheus/common/promslog"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/receiver/dingtalk"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
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
	HandleAlertCheckRecord(ctx_core core.Context, ctx context.Context, record *model.WorkflowRecord) error

	GetAMConfigReceiver(ctx_core core.Context, filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int)
	AddAMConfigReceiver(ctx_core core.Context, receiver amconfig.Receiver) error
	UpdateAMConfigReceiver(ctx_core core.Context, receiver amconfig.Receiver, oldName string) error
	DeleteAMConfigReceiver(ctx_core core.Context, name string) error

	ListSlienceConfig(ctx_core core.Context) ([]slienceconfig.AlertSlienceConfig, error)
	GetSlienceConfigByAlertID(ctx_core core.Context, alertID string) (*slienceconfig.AlertSlienceConfig, error)
	SetSlienceConfigByAlertID(ctx_core core.Context, alertID string, forDuration string) error
	RemoveSlienceConfigByAlertID(ctx_core core.Context, alertID string) error
}

type InnerReceivers struct {
	database database.Repo
	ch       clickhouse.Repo

	receivers map[string][]notify.Integration

	externalURL *url.URL
	logger      *slog.Logger

	// alertID -> slienceconfig.AlertSlienceConfig
	slientCFGMap sync.Map
}

func SetupReceiver(externalURL string, logger *zap.Logger, dbRepo database.Repo, chRepo clickhouse.Repo) (Receivers, error) {
	// TODO not support custom template now
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}
	tmpl.ExternalURL, err = url.Parse(externalURL)
	if err != nil {
		return nil, err
	}

	// TODO ctx_core
	receivers, _, err := dbRepo.GetAMConfigReceiver(nil, nil, nil)
	if err != nil {
		return nil, err
	}

	amReceiver, err := buildInnerReceivers(
		receivers,
		tmpl,
		util.NewZapSlogHandler(logger),
	)

	if err != nil {
		return nil, err
	}
	if amReceiver != nil {
		amReceiver.database = dbRepo
		amReceiver.ch = chRepo
	}

	// ctx_core
	slienceCfgs, err := dbRepo.GetAlertSlience(nil)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(slienceCfgs); i++ {
		amReceiver.slientCFGMap.Store(slienceCfgs[i].AlertID, &slienceCfgs[i])
	}

	go amReceiver.cleanupSlience(context.Background(), time.Minute)
	return amReceiver, err
}

func buildInnerReceivers(ncs []amconfig.Receiver, tmpl *template.Template, logger *slog.Logger, httpOpts ...commoncfg.HTTPClientOption) (*InnerReceivers, error) {
	var errs error

	var innerReceivers = &InnerReceivers{
		receivers:   make(map[string][]notify.Integration),
		externalURL: tmpl.ExternalURL,
		logger:      logger,
	}
	for _, nc := range ncs {
		integrations, err := buildReceiverIntegrations(nc, tmpl, logger, httpOpts...)
		if err != nil {
			logger.Error("Error building integrations", "err", err)
			errs = errors.Join(errs, err)
		}
		innerReceivers.receivers[nc.Name] = integrations
	}

	return innerReceivers, errs
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
		c.AlertName = nc.Name
		add("dingtalk", i, c, func(l *slog.Logger) (notify.Notifier, error) { return dingtalk.New(c, tmpl, l), nil })
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

func (r *InnerReceivers) cleanupSlience(ctx context.Context, interval time.Duration) {
	cleanupTicker := time.NewTicker(interval)
	for {
		select {
		case <-cleanupTicker.C:
			now := time.Now()
			r.slientCFGMap.Range(func(key, value any) bool {
				val := value.(*slienceconfig.AlertSlienceConfig)
				if now.After(val.EndAt) {
					// ctx_core
					if err := r.database.DeleteAlertSlience(nil, val.ID); err == nil {
						r.slientCFGMap.Delete(key)
					}
				}
				return true
			})
		case <-ctx.Done():
			return
		}
	}
}
