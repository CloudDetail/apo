package kubernetes

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func NewTestRepo(t *testing.T) Repo {
	viper.SetConfigFile("testdata/config.yml")
	viper.ReadInConfig()

	authType := viper.GetString("kubernetes.authType")
	authFilePath := viper.GetString("kubernetes.authFilePath")

	zapLog := logger.NewLogger(logger.WithLevel("debug"))
	repo, err := New(zapLog, authType, authFilePath, config.MetadataSettings{})
	if err != nil {
		t.Fatalf("Error to connect clickhouse: %v", err)
	}
	return repo
}

func TestLoad(t *testing.T) {
	data := `global:
    resolve_timeout: 5m
    http_config:
        follow_redirects: true
        enable_http2: true
    smtp_hello: localhost
    smtp_require_tls: true
    pagerduty_url: https://events.pagerduty.com/v2/enqueue
    opsgenie_api_url: https://api.opsgenie.com/
    wechat_api_url: https://qyapi.weixin.qq.com/cgi-bin/
    victorops_api_url: https://alert.victorops.com/integrations/generic/20131114/alert/
    telegram_api_url: https://api.telegram.org
    webex_api_url: https://webexapis.com/v1/messages
route:
    receiver: 告警通知名测试编辑2
    group_by:
        - alertname
    continue: false
    group_wait: 30s
    group_interval: 10s
    repeat_interval: 1m
receivers:
    - name: 告警通知名测试编辑2
      webhook_configs:
            - send_resolved: false
              http_config:
                basic_auth:
                    username: 身份认证用户名
                    password: passwoard
                tls_config:
                    insecure_skip_verify: true
                follow_redirects: true
                enable_http2: true
                http_headers:
                    testkey:
                        values:
                            - testvalue
              url: http://localhost:300
              url_file: ""
              max_alerts: 0
templates: []
`

	_, err := amconfig.Load(data)
	if err != nil {
		t.Fatal(err)
	}

	jsondata := `{
    "name": "告警通知名带header2",
    "emailConfigs": [
        {
            "to": "luketing1999@gmail.com",
            "from": "luketing1999@gmail.com",
            "authUsername": "SMTP用户名",
            "authPassword": "SMTPpasswoard",
            "html": "HTML正文",
            "smarthost": "SMTP服务器Host:123",
            "tlsConfig": {
                "insecureSkipVerify": true
            }
        }
    ],
    "webhookConfigs": [
        {
            "url": "http://localhost:300",
            "httpConfig": {
                "basicAuth": {
                    "username": "身份认证用户名",
                    "password": "passwoard"
                },
                "httpHeaders": {
                    "headers": {
                        "headertestkey": {
                            "values": [
                                "headertestvalue"
                            ]
                        }
                    }
                },
                "tlsConfig": {
                    "insecureSkipVerify": true
                }
            }
        }
    ]
}`
	var amReceiver amconfig.Receiver
	err = json.Unmarshal([]byte(jsondata), &amReceiver)
	if err != nil {
		t.Fatal(err)
	}
	out, _ := yaml.Marshal(amReceiver)

	log.Printf("amConfig: %s", string(out))
}
