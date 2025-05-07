package dingtalk

import (
	"context"
	"log/slog"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/services/alerts/notification"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
)

var _ notify.Notifier = &Notifier{}

type Notifier struct {
	conf *amconfig.DingTalkConfig
	tmpl *template.Template
	l    *slog.Logger
}

// Notify implements notify.Notifier.
func (n *Notifier) Notify(ctx context.Context, alerts ...*types.Alert) (bool, error) {
	// construct builder
	builder, err := notification.NewNotificationBuilder()
	if err != nil {
		n.l.Warn("get dingtalk builder failed", "err", err)
		return false, err
	}

	dingTalkNotification, err := builder.BuildByAlerts(
		n.conf.AlertName,
		"firing",
		"",
		0,
		alerts...,
	)
	if err != nil {
		n.l.Warn("build dingtalk notification failed", "err", err)
		return false, err
	}
	err = notification.SendNotification(dingTalkNotification, n.conf.URL, n.conf.Secret)
	if err != nil {
		n.l.Warn("send dingtalk notification failed", "err", err)
		return true, err
	}
	return false, nil
}

func New(conf *amconfig.DingTalkConfig, tmpl *template.Template, l *slog.Logger) *Notifier {
	return &Notifier{
		conf: conf,
		tmpl: tmpl,
		l:    l,
	}
}
