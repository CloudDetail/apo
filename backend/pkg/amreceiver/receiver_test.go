package amreceiver

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestReceiver(t *testing.T) {
	zapLog := logger.NewLogger(logger.WithLevel("debug"))

	conf := amconfig.Config{}
	cfgBytes, err := os.ReadFile("./testdata/receiver.yaml")
	assert.NoError(t, err)

	err = yaml.Unmarshal(cfgBytes, &conf)
	assert.NoError(t, err)

	tmpl, err := template.FromGlobs(conf.Templates)
	assert.NoError(t, err)
	tmpl.ExternalURL, err = url.Parse("http://example.com")
	assert.NoError(t, err)
	rm, err := buildInnerReceivers(
		conf.Receivers,
		tmpl,
		util.NewZapSlogHandler(zapLog),
	)

	assert.NoError(t, err)
	assert.Equal(t, 4, len(rm.receivers))

	integrations, ok := rm.receivers["APO Alert Collector"]
	if !ok {
		t.Fatal("missing APO Alert Collector")
	}

	for _, integration := range integrations {
		ctx := notify.WithNow(context.Background(), time.Now())
		// Populate context with information likeed along the pipeline.
		ctx = notify.WithGroupKey(ctx, "1")
		ctx = notify.WithGroupLabels(ctx, model.LabelSet{"key": "value"})
		ctx = notify.WithReceiverName(ctx, "apoaa")

		ok, err := integration.Notify(ctx, &types.Alert{
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
		})

		assert.NoError(t, err)
		assert.False(t, ok)
	}
}
