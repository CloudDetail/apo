package dingtalk

import (
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/services/alerts/notification"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func TestBuildNotifyFaction(t *testing.T) {
	builder, err := notification.NewNotificationBuilder()
	assert.NoError(t, err)

	_, err = builder.Build(&request.ForwardToDingTalkRequest{
		Receiver: "dingtalkTest",
		Status:   "firing",
		Alerts: toAlerts([]*types.Alert{
			{
				Alert: model.Alert{
					Labels: model.LabelSet{
						"labels": "asd",
					},
					Annotations: model.LabelSet{
						"annos": "asdzxiouwqe",
					},
					StartsAt:     time.Now(),
					EndsAt:       time.Now(),
					GeneratorURL: "https://example.com",
				},
				UpdatedAt: time.Now(),
				Timeout:   false,
			},
		}),
		GroupLabels:       request.KV{},
		CommonLabels:      request.KV{},
		CommonAnnotations: request.KV{},
		TruncatedAlerts:   0,
		ExternalURL:       "",
	})

	assert.NoError(t, err)

	// TODO mock send
}
