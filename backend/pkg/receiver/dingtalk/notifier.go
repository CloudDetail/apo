package dingtalk

import (
	"context"
	"log/slog"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/services/alerts/notification"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
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

	groupLabels, find := notify.GroupLabels(ctx)
	if !find {
		groupLabels = make(model.LabelSet)
	}

	dingTalkNotification, err := builder.Build(
		&request.ForwardToDingTalkRequest{
			Receiver:          n.conf.AlertName,
			Status:            "firing",
			Alerts:            toAlerts(alerts),
			GroupLabels:       toKVs(groupLabels),
			CommonLabels:      request.KV{},
			CommonAnnotations: request.KV{},
			TruncatedAlerts:   0,
			ExternalURL:       n.tmpl.ExternalURL.String(),
		},
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

func toKVs(labelSet model.LabelSet) request.KV {
	labels := map[string]string{}
	for k, v := range labelSet {
		labels[string(k)] = string(v)
	}
	return labels
}

func toAlerts(alerts []*types.Alert) request.Alerts {
	var res []request.Alert
	for i := 0; i < len(alerts); i++ {
		res = append(res, request.Alert{
			Status:       string(alerts[i].Status()),
			Labels:       toKVs(alerts[i].Labels),
			Annotations:  toKVs(alerts[i].Annotations),
			StartsAt:     alerts[i].StartsAt.String(),
			EndsAt:       alerts[i].EndsAt.String(),
			GeneratorURL: alerts[i].GeneratorURL,
			Fingerprint:  alerts[i].Fingerprint().String(),
		})
	}
	return res
}
