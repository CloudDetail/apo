package kubernetes

import (
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	AuthTypeNone           string = "none"
	AuthTypeServiceAccount string = "serviceAccount"
	AuthTypeKubeConfig     string = "kubeConfig"

	// DefaultKubeConfigPath Default kubeconfig path
	DefaultKubeConfigPath string = "~/.kube/config"

	// DefaultMetaSetting
	DefaultAPONS         string = "apo"
	DefaultCMNAME        string = "apo-victoria-metrics-alert-server-alert-rules-config"
	DefaultAlertRuleFile string = "alert-rules.yaml"
	DefaultAMCMName      string = "apo-alertmanager-config"
	DefaultAMConfigFile  string = "alertmanager.yaml"
)

var _ Repo = &k8sApi{}

type Repo interface {
	// Sync with K8sAPIServer
	SyncNow() error

	GetAlertRuleConfigFile(alertRuleFile string) (map[string]string, error)
	UpdateAlertRuleConfigFile(configFile string, content []byte) error

	GetAlertRules(configFile string, filter *request.AlertRuleFilter, pageParam *request.PageParam) ([]*request.AlertRule, int)
	AddOrUpdateAlertRule(configFile string, alertRule request.AlertRule) error
	DeleteAlertRule(configFile string, group, alert string) error

	GetAMConfigReceiver(configFile string, filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int)
	AddOrUpdateAMConfigReceiver(configFile string, receiver amconfig.Receiver) error
	DeleteAMConfigReceiver(configFile string, name string) error
}

func New(logger *zap.Logger, authType, authFilePath string, setting config.MetadataSettings) (Repo, error) {
	restConfig, err := createRestConfig(authType, authFilePath)
	if err != nil {
		logger.Info("failed to setup kubernetes repository, skip init", zap.Error(err))
		return NoneRepo, nil
	}

	ctrl.SetLogger(zapr.NewLogger(logger))

	cli, err := client.New(restConfig, client.Options{})
	if err != nil {
		return NoneRepo, err
	}

	if len(setting.Namespace) == 0 {
		setting.Namespace = DefaultAPONS
	}
	if len(setting.AlertRuleCMName) == 0 {
		setting.AlertRuleCMName = DefaultCMNAME
	}
	if len(setting.AlertRuleFileName) == 0 {
		setting.AlertRuleFileName = DefaultAlertRuleFile
	}
	if len(setting.AlertManagerCMName) == 0 {
		setting.AlertManagerCMName = DefaultAMCMName
	}
	if len(setting.AlertManagerFileName) == 0 {
		setting.AlertManagerFileName = DefaultAMConfigFile
	}

	api := &k8sApi{
		logger:           logger,
		cli:              cli,
		MetadataSettings: setting,

		Metadata: Metadata{
			AlertRulesMap: map[string]*AlertRules{},
			AMConfigMap:   map[string]*amconfig.Config{},
		},
	}

	api.SyncNow()

	return api, nil
}

type k8sApi struct {
	logger *zap.Logger
	cli    client.Client

	config.MetadataSettings
	Metadata
}
