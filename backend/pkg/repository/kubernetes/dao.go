package kubernetes

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AuthTypeNone           string = "none"
	AuthTypeServiceAccount string = "serviceAccount"
	AuthTypeKubeConfig     string = "kubeConfig"

	// DefaultKubeConfigPath Default kubeconfig path
	DefaultKubeConfigPath string = "~/.kube/config"
)

type Repo interface {
	GetAlertManagerRule() (string, error)
	UpdateAlertManagerRule(alertRules string) error
}

func New(logger *zap.Logger, authType string, authFilePath string) (Repo, error) {
	restConfig, err := createRestConfig(authType, authFilePath)
	if err != nil {
		return nil, err
	}

	cli, err := client.New(restConfig, client.Options{})
	if err != nil {
		return nil, err
	}

	return &k8sApi{
		cli: cli,
	}, nil
}

type k8sApi struct {
	cli client.Client
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
