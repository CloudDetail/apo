package kubernetes

import (
	"encoding/json"
	"testing"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/spf13/viper"
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

func TestUnmarshal(t *testing.T) {

	jsondata := `{
    "url": "http://localhost:300",
    "httpConfig": {
        "basicAuth": {
            "username": "身份认证用户名",
            "password": "passwoard"
        },
        "httpHeaders": {
            "headertestkey": {
                    "values": [
                        "headertestvalue"
                    ]
            }
        },
        "tlsConfig": {
            "insecureSkipVerify": true
        }
    }
}`
	receiver := amconfig.WebhookConfig{}

	err := json.Unmarshal([]byte(jsondata), &receiver)
	if err != nil {
		t.Fatal(err)
	}

	out, err := json.Marshal(receiver)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("out: %s", string(out))
}
