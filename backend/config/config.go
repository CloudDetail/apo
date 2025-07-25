// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"strconv"
	"time"

	"github.com/CloudDetail/metadata/configs"
	"github.com/spf13/viper"
)

var config *Config

type Config struct {
	Server struct {
		Port                     int `mapstructure:"port"`
		AccessTokenExpireMinutes int `mapstructure:"access_token_expire_minutes"`
		RefreshTokenExpireHours  int `mapstructure:"refresh_token_expire_hours"`
	} `mapstructure:"server"`
	Logger struct {
		Level         string `mapstructure:"level"`
		EnableConsole bool   `mapstructure:"console_enable"`
		EnableFile    bool   `mapstructure:"file_enable"`
		FilePath      string `mapstructure:"file_path"`
		FileNum       int    `mapstructure:"file_num"`
		FileSize      int    `mapstructure:"file_size_mb"`
	} `mapstructure:"logger"`
	Database struct {
		Connection string `mapstructure:"connection"`
		MaxOpen    int    `mapstructure:"max_open"`
		MaxIdle    int    `mapstructure:"max_idle"`
		MaxLife    int    `mapstructure:"max_life_second"`
		MySql      struct {
			Host     string `mapstructure:"host"`
			Port     int    `mapstructure:"port"`
			Database string `mapstructure:"database"`
			UserName string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			Charset  string `mapstructure:"charset"`
		} `mapstructure:"mysql"`
		Sqllite struct {
			Database string `mapstructure:"database"`
		} `mapstructure:"sqllite"`
		Postgres struct {
			Host     string `mapstructure:"host"`
			Port     int    `mapstructure:"port"`
			Database string `mapstructure:"database"`
			UserName string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			SSLMode  string `mapstructure:"sslmode"`
			Timezone string `mapstructure:"timezone"`
		}
	} `mapstructure:"database"`
	ClickHouse struct {
		Address  string `mapstructure:"address"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Database string `mapstructure:"database"`
		Cluster  string `mapstructure:"cluster"`
		Replica  bool   `mapstructure:"replica"`
		// pool config
		MaxOpenConns           int `mapstructure:"max_open_conns"`
		MaxIdleConns           int `mapstructure:"max_idle_conns"`
		ConnMaxLifetimeMinutes int `mapstructure:"conn_max_lifetime_minutes"`
		DialTimeoutSeconds     int `mapstructure:"dial_timeout_seconds"`
	} `mapstructure:"clickhouse"`
	Promethues struct {
		Address string `mapstructure:"address"`
		Storage string `mapstructure:"storage"`
	} `mapstructure:"promethues"`
	Kubernetes struct {
		AuthType     string `mapstructure:"auth_type"`
		AuthFilePath string `mapstructure:"auth_file_path"`

		MetadataSettings MetadataSettings `mapstructure:"metadata_settings"`
	} `mapstructure:"kubernetes"`
	MetaServer struct {
		Enable           bool                     `mapstructure:"enable"`
		MetaSourceConfig configs.MetaSourceConfig `mapstructure:"meta_source_config"`
	} `mapstructure:"meta_server"`
	Dataplane struct {
		Address string `mapstructure:"address"`
	} `mapstructure:"dataplane"`
	Jaeger struct {
		Address string `mapstructure:"address"`
	}
	DeepFlow struct {
		ServerAddress string `mapstructure:"server_address"`
		ChAddress     string `mapstructure:"ch_address"`
		ChUsername    string `mapstructure:"ch_username"`
		ChPassword    string `mapstructure:"ch_password"`
	} `mapstructure:"deepflow"`
	User struct {
		AnonymousUser `mapstructure:"anonymous_user"`
	} `mapstructure:"user"`
	Dify          DifyConfig `mapstructure:"dify"`
	AlertReceiver struct {
		Enabled     bool   `mapstructure:"enabled"`
		ExternalURL string `mapstructure:"external_url"`
	} `mapstructure:"alert_receiver"`
	DataGroup struct {
		InitLookBackDays int `mapstructure:"init_look_back_days"`
		RefreshSeconds   int `mapstructure:"refresh_seconds"`
	} `mapstructure:"data_group"`
	Incident struct {
		Enabled bool `mapstructure:"enabled"`
	} `mapstructure:"incident"`
}

type AnonymousUser struct {
	Username string `mapstructure:"username"` // TODO deprecated
	Enable   bool   `mapstructure:"enable"`
	Role     string `mapstructure:"role"`
}

type MetadataSettings struct {
	Namespace         string `mapstructure:"namespace"`
	AlertRuleCMName   string `mapstructure:"alert_rule_configmap_name"`
	AlertRuleFileName string `mapstructure:"alert_rule_file_name"`

	AlertManagerCMName   string `mapstructure:"alert_manager_configmap"`
	AlertManagerFileName string `mapstructure:"alert_manager_file_name"`
	VectorCMName         string `mapstructure:"vector_configmap"`
	VectorFileName       string `mapstructure:"vector_file_name"`
}

type DifyConfig struct {
	User    string `mapstructure:"user"`
	URL     string `mapstructure:"url"`
	APIKeys struct {
		AlertCheck    string `mapstructure:"alert_check"`
		AlertClassify string `mapstructure:"alert_classify"`
		AlertAnalyze  string `mapstructure:"alert_analyze"`
	} `mapstructure:"api_keys"`
	FlowIDs struct {
		AlertCheck        string `mapstructure:"alert_check"`
		AlertEventAnalyze string `mapstructure:"alert_event_analyze"`
	} `mapstructure:"flow_ids"`
	MaxConcurrency int    `mapstructure:"max_concurrency"`
	CacheMinutes   int    `mapstructure:"cache_minutes"`
	TimeoutSecond  int    `mapstructure:"timeout_second"`
	Sampling       string `mapstructure:"sampling"`
	AutoCheck      bool   `mapstructure:"auto_check"`
	AutoAnalyze    bool   `mapstructure:"auto_analyze"`
	Retry          bool   `mapstructure:"retry"`
}

func Get() *Config {
	if config == nil {
		viper.SetConfigType("yaml")
		configFile, found := os.LookupEnv("APO_CONFIG")
		if !found {
			configFile = "./config/apo.yml"
		}
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
		if err := viper.Unmarshal(&config); err != nil {
			panic(err)
		}
	}
	return config
}

func GetCHCluster() string {
	return config.ClickHouse.Cluster
}

// GetClickHouseConnPoolConfig
// gets the ClickHouse connection configuration, supporting environment variable overrides.
func GetClickHouseConnPoolConfig() (maxOpenConns, maxIdleConns int, connMaxLifetime, dialTimeout time.Duration) {
	cfg := Get().ClickHouse

	maxOpenConns = cfg.MaxOpenConns
	if envVal := os.Getenv("APO_CH_MAX_OPEN_CONNS"); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil {
			maxOpenConns = val
		}
	}
	if maxOpenConns <= 0 {
		maxOpenConns = 20 // default
	}

	maxIdleConns = cfg.MaxIdleConns
	if envVal := os.Getenv("APO_CH_MAX_IDLE_CONNS"); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil {
			maxIdleConns = val
		}
	}
	if maxIdleConns <= 0 {
		maxIdleConns = 10 // default
	}

	connMaxLifetimeMinutes := cfg.ConnMaxLifetimeMinutes
	if envVal := os.Getenv("APO_CH_CONN_MAX_LIFETIME_MINUTES"); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil {
			connMaxLifetimeMinutes = val
		}
	}
	if connMaxLifetimeMinutes <= 0 {
		connMaxLifetimeMinutes = 60 // default
	}
	connMaxLifetime = time.Duration(connMaxLifetimeMinutes) * time.Minute

	dialTimeoutSeconds := cfg.DialTimeoutSeconds
	if envVal := os.Getenv("APO_CH_DIAL_TIMEOUT_SECONDS"); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil {
			dialTimeoutSeconds = val
		}
	}
	if dialTimeoutSeconds <= 0 {
		dialTimeoutSeconds = 5 // default
	}

	dialTimeout = time.Duration(dialTimeoutSeconds) * time.Second

	return
}
