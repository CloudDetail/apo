package kubernetes

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/CloudDetail/apo/backend/config"
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
)

type Repo interface {
	// Sync with K8sAPIServer
	SyncNow() error

	GetAlertRuleConfigFile(alertRuleFile string) (map[string]string, error)
	UpdateAlertRuleConfigFile(configFile string, content []byte) error

	GetAlertRules(configFile string, filter *request.AlertRuleFilter, pageParam *request.PageParam) ([]*request.AlertRule, int)
	AddOrUpdateAlertRule(configFile string, alertRule request.AlertRule) error
	DeleteAlertRule(configFile string, group, alert string) error
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

	api := &k8sApi{
		logger:           logger,
		cli:              cli,
		MetadataSettings: setting,

		Metadata: Metadata{
			AlertRulesMap: map[string]*AlertRules{},
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

// createRestConfig creates an Kubernetes API config from user configuration.
func createRestConfig(authType string, authFilePath string) (*rest.Config, error) {
	var authConf *rest.Config
	var err error

	var k8sHost string
	if authType != AuthTypeKubeConfig {
		host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
		if len(host) == 0 || len(port) == 0 {
			return nil, fmt.Errorf("unable to load k8s config, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
		}
		k8sHost = "https://" + net.JoinHostPort(host, port)
	}

	switch authType {
	case AuthTypeKubeConfig:
		if authFilePath == "" {
			authFilePath = DefaultKubeConfigPath
		}
		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: authFilePath}
		configOverrides := &clientcmd.ConfigOverrides{}
		authConf, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules, configOverrides).ClientConfig()

		if err != nil {
			return nil, fmt.Errorf("error connecting to k8s with auth_type=%s: %w", AuthTypeKubeConfig, err)
		}
	case AuthTypeNone:
		authConf = &rest.Config{
			Host: k8sHost,
		}
		authConf.Insecure = true
	case AuthTypeServiceAccount:
		// This should work for most clusters but other auth types can be added
		authConf, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("none of kubernetes auth config is set, ignore kubernetes repository")
	}

	authConf.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		// Don't use system proxy settings since the API is local to the
		// cluster
		if t, ok := rt.(*http.Transport); ok {
			t.Proxy = nil
		}
		return rt
	}

	return authConf, nil
}

var NoneRepo Repo = &NoneAPI{}

type NoneAPI struct{}

var ErrKubernetesRepoNotReady = errors.New("kubernetes repo is not ready")

func (n *NoneAPI) AddOrUpdateAlertRule(configFile string, alertRule request.AlertRule) error {
	return ErrKubernetesRepoNotReady
}

func (n *NoneAPI) DeleteAlertRule(configFile string, group string, alert string) error {
	return ErrKubernetesRepoNotReady
}

// GetAlertRuleConfigFile implements Repo.
func (n *NoneAPI) GetAlertRuleConfigFile(alertRuleFile string) (map[string]string, error) {
	return nil, ErrKubernetesRepoNotReady
}

// GetAlertRules implements Repo.
func (n *NoneAPI) GetAlertRules(configFile string, filter *request.AlertRuleFilter, pageParam *request.PageParam) ([]*request.AlertRule, int) {
	return []*request.AlertRule{}, 0
}

// SyncNow implements Repo.
func (n *NoneAPI) SyncNow() error {
	return ErrKubernetesRepoNotReady
}

// UpdateAlertRuleConfigFile implements Repo.
func (n *NoneAPI) UpdateAlertRuleConfigFile(configFile string, content []byte) error {
	return ErrKubernetesRepoNotReady
}
