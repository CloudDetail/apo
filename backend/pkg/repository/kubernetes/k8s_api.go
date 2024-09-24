package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

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

func (k *k8sApi) SyncNow() error {
	var combineErr error
	err := k.syncAlertRule()
	if err != nil {
		k.logger.Warn("failed to sync alertRule with k8sAPI", zap.Error(err))
		combineErr = multierror.Append(combineErr, err)
	}

	err = k.syncAMConfig()
	if err != nil {
		k.logger.Warn("failed to sync alertManagerConfig with k8sAPI", zap.Error(err))
		combineErr = multierror.Append(combineErr, err)
	}

	return combineErr
}

func (k *k8sApi) getConfigMap(cm string, dataKey string) (map[string]string, error) {
	obj := &v1.ConfigMap{}
	key := client.ObjectKey{
		Namespace: k.Namespace,
		Name:      cm,
	}

	err := k.cli.Get(context.Background(), key, obj)
	if err != nil {
		return nil, err
	}

	if len(dataKey) > 0 {
		return map[string]string{
			dataKey: obj.Data[dataKey],
		}, nil
	}

	return obj.Data, nil
}

func (k *k8sApi) updateConfigMap(cm string, dataKey string, content []byte) error {
	obj := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cm,
			Namespace: k.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(context.Background(), k.cli, obj, func() error {
		if content == nil {
			delete(obj.Data, dataKey)
		} else {
			obj.Data[dataKey] = string(content)
		}
		return nil
	})
	return err
}

var NoneRepo Repo = &NoneAPI{}

type NoneAPI struct{}

var ErrKubernetesRepoNotReady = errors.New("kubernetes repo is not ready")

// SyncNow implements Repo.
func (n *NoneAPI) SyncNow() error {
	return ErrKubernetesRepoNotReady
}

func (n *NoneAPI) GetAlertRules(configFile string, filter *request.AlertRuleFilter, pageParam *request.PageParam) ([]*request.AlertRule, int) {
	return []*request.AlertRule{}, 0
}

func (n *NoneAPI) AddOrUpdateAlertRule(configFile string, alertRule request.AlertRule) error {
	return ErrKubernetesRepoNotReady
}

func (n *NoneAPI) DeleteAlertRule(configFile string, group string, alert string) error {
	return ErrKubernetesRepoNotReady
}

// GetAlertRules implements Repo.

// GetAlertRuleConfigFile implements Repo.
func (n *NoneAPI) GetAlertRuleConfigFile(alertRuleFile string) (map[string]string, error) {
	return nil, ErrKubernetesRepoNotReady
}

// UpdateAlertRuleConfigFile implements Repo.
func (n *NoneAPI) UpdateAlertRuleConfigFile(configFile string, content []byte) error {
	return ErrKubernetesRepoNotReady
}

// AddOrUpdateAMConfigReceiver implements Repo.
func (n *NoneAPI) AddOrUpdateAMConfigReceiver(configFile string, receiver amconfig.Receiver) error {
	return ErrKubernetesRepoNotReady
}

// DeleteAMConfigReceiver implements Repo.
func (n *NoneAPI) DeleteAMConfigReceiver(configFile string, name string) error {
	return ErrKubernetesRepoNotReady
}

// GetAMConfigReceiver implements Repo.
func (n *NoneAPI) GetAMConfigReceiver(configFile string, filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int) {
	return []amconfig.Receiver{}, 0
}
